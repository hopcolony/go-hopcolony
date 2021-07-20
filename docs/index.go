package docs

import (
	"encoding/json"
	"fmt"
)

type Index struct {
	Name    string
	NumDocs int
	Status  string
}

type IndexSnapshot struct {
	Docs    []Document
	Success bool
	Reason  string
}

type Query struct {
	Field    string
	Operator string
	Value    interface{}
}

type IndexReference struct {
	client  HopDocClient
	Index   string
	Queries []Query
}

type CompoundBody struct {
	Size  int `json:"size"`
	From  int `json:"from"`
	Query struct {
		Bool struct {
			Must   []map[string]interface{} `json:"must"`
			Filter []map[string]interface{} `json:"filter"`
		} `json:"bool"`
	} `json:"query"`
}

func operatorToComparison(operator string) string {
	switch operator {
	case "<":
		return "lt"
	case "<=":
		return "lte"
	case ">":
		return "gt"
	default:
		return "gte"
	}
}
func (i *IndexReference) CompoundBody(size int, from int) CompoundBody {
	var compoundBody CompoundBody
	compoundBody.Size = size
	compoundBody.From = from
	must := make([]map[string]interface{}, 0)
	for _, q := range i.Queries {
		query := make(map[string]interface{})
		switch q.Operator {
		case "==":
			query["match"] = map[string]interface{}{
				q.Field: q.Value,
			}
		case "<", "<=", ">", ">=":
			query["range"] = map[string]interface{}{
				q.Field: map[string]interface{}{
					operatorToComparison(q.Operator): q.Value,
				},
			}
		}
		must = append(must, query)
	}
	compoundBody.Query.Bool.Must = must
	compoundBody.Query.Bool.Filter = make([]map[string]interface{}, 0)
	return compoundBody
}

type IndexGetResponse struct {
	Hits struct {
		Hits []Document
	}
}

func (i *IndexReference) Get() *IndexSnapshot {
	jsonData, err := json.Marshal(i.CompoundBody(100, 0))
	if err != nil {
		return &IndexSnapshot{Success: false, Reason: err.Error()}
	}

	resp, err := i.client.Post(fmt.Sprintf("/%s/_search", i.Index), jsonData)
	if err != nil {
		return &IndexSnapshot{Success: false, Reason: err.Error()}
	}

	var result IndexGetResponse
	json.Unmarshal(resp, &result)

	return &IndexSnapshot{result.Hits.Hits, true, ""}
}

func (i *IndexReference) Add(data interface{}) DocumentSnapshot {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return DocumentSnapshot{Success: false, Reason: err.Error()}
	}

	resp, err := i.client.Post(fmt.Sprintf("/%s/_doc", i.Index), jsonData)
	if err != nil {
		return DocumentSnapshot{Success: false, Reason: err.Error()}
	}

	var document Document
	err = json.Unmarshal(resp, &document)
	if err != nil {
		return DocumentSnapshot{Success: false, Reason: err.Error()}
	}

	var result map[string]interface{}
	json.Unmarshal(jsonData, &result)
	document.Source = result

	return DocumentSnapshot{&document, true, ""}
}

func (i *IndexReference) Delete() error {
	return i.client.Delete("/" + i.Index)
}

func (i *IndexReference) Document(id string) *DocumentReference {
	return &DocumentReference{i.client, i.Index, id}
}

func (i *IndexReference) Count() (int, error) {
	resp, err := i.client.Get(fmt.Sprintf("/%s/_count", i.Index))
	if err != nil {
		return 0, fmt.Errorf("could not get index count: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(resp, &result)
	return int(result["count"].(float64)), nil
}

func (i *IndexReference) Where(field, operator string, value interface{}) *IndexReference {
	i.Queries = append(i.Queries, Query{field, operator, value})
	return i
}
