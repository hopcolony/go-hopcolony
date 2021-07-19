package docs

import (
	"encoding/json"
	"fmt"

	"github.com/mitchellh/mapstructure"
)

type UpdateData struct {
	Key   string
	Value interface{}
}

type Document struct {
	Source  map[string]interface{} `json:"_source"`
	Index   string                 `json:"_index"`
	Id      string                 `json:"_id"`
	Version int                    `json:"_version"`
}

func (ds *Document) Map() map[string]interface{} {
	return ds.Source
}

func (ds *Document) DataTo(in interface{}) error {
	err := mapstructure.Decode(ds.Source, in)
	return err
}

type DocumentSnapshot struct {
	Doc     *Document
	Success bool
	Reason  string
}

type DocumentReference struct {
	client HopDocClient
	Index  string
	Id     string
}

func (d *DocumentReference) Get() DocumentSnapshot {
	resp, err := d.client.Get(fmt.Sprintf("/%s/_doc/%s", d.Index, d.Id))
	if err != nil {
		return DocumentSnapshot{Success: false, Reason: err.Error()}
	}

	var document Document
	json.Unmarshal(resp, &document)

	return DocumentSnapshot{&document, true, ""}
}

func (d *DocumentReference) SetData(data interface{}) DocumentSnapshot {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return DocumentSnapshot{Success: false, Reason: err.Error()}
	}

	resp, err := d.client.Post(fmt.Sprintf("/%s/_doc/%s", d.Index, d.Id), jsonData)
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

func (d *DocumentReference) Update(updates []UpdateData) DocumentSnapshot {
	data := make(map[string]interface{})
	data["doc"] = make(map[string]interface{})
	for _, update := range updates {
		data["doc"].(map[string]interface{})[update.Key] = update.Value
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return DocumentSnapshot{Success: false, Reason: err.Error()}
	}

	_, err = d.client.Post(fmt.Sprintf("/%s/_doc/%s/_update", d.Index, d.Id), jsonData)
	if err != nil {
		return DocumentSnapshot{Success: false, Reason: err.Error()}
	}

	return d.Get()
}

func (d *DocumentReference) Delete() DocumentSnapshot {
	err := d.client.Delete(fmt.Sprintf("/%s/_doc/%s", d.Index, d.Id))
	if err != nil {
		return DocumentSnapshot{Success: false, Reason: err.Error()}
	}
	return DocumentSnapshot{nil, true, ""}
}
