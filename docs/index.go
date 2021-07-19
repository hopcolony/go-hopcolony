package docs

import (
	"encoding/json"
	"fmt"
)

type Index struct {
	Name    string
	NumDocs int
	Status  string `json:"status"`
}

type IndexSnapshot struct {
	Docs    []Document
	Success bool
	Reason  string
}

type IndexReference struct {
	client HopDocClient
	Index  string
}

type BoolQuery struct {
	Must   []interface{} `json:"must"`
	Filter []interface{} `json:"filter"`
}
type Query struct {
	Bool BoolQuery `json:"bool"`
}
type CompoundBody struct {
	Size  int   `json:"size"`
	From  int   `json:"from"`
	Query Query `json:"query"`
}

func (i *IndexReference) CompoundBody(size int, from int) CompoundBody {
	var compoundBody CompoundBody
	compoundBody.Size = size
	compoundBody.From = from
	compoundBody.Query.Bool.Must = make([]interface{}, 0)
	compoundBody.Query.Bool.Filter = make([]interface{}, 0)
	return compoundBody
}

type Hits struct {
	Hits []Document `json:"hits"`
}

type IndexGetResponse struct {
	Hits Hits `json:"hits"`
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
		return 0, err
	}
	var result map[string]interface{}
	json.Unmarshal(resp, &result)
	return int(result["count"].(float64)), nil
}
