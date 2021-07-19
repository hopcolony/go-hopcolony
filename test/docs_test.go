package test

import (
	"os"
	"reflect"
	"testing"

	"hopcolony.io/hopcolony/docs"
	"hopcolony.io/hopcolony/initialize"
)

type Data struct {
	Purpose string `json:"purpose"`
}

var (
	project initialize.Project
	db      docs.HopDoc

	index   string
	uid     string
	dataMap map[string]interface{}
	data    Data
)

func TestDocsInitialize(t *testing.T) {
	var err error
	project, err = initialize.Initialize(initialize.ProjectConfig{Username: os.Getenv("HOP_USER_NAME"),
		Project: os.Getenv("HOP_PROJECT_NAME"), Token: os.Getenv("HOP_TOKEN")})

	if err != nil {
		t.Errorf("Error in project creation: %s", err)
	}

	db, err = docs.New()
	if err != nil {
		t.Errorf("Error while creating new docs client: %v", err)
	}

	if db.Project.Config.Identity != project.Config.Identity {
		t.Errorf(`Expected db identity to be "%s" but got "%s"`, project.Config.Identity, db.Project.Config.Identity)
	}

	index = ".hop.tests"
	uid = "hopcolony"
	dataMap = map[string]interface{}{
		"purpose": "Test Hop Docs!",
	}
	data = Data{Purpose: "Test Hop Docs with structs!"}
}

func TestDocsStatus(t *testing.T) {
	status, err := db.Status()
	if err != nil {
		t.Errorf("Error while requesting status: %v", err)
	}

	if status == "red" {
		t.Errorf(`Invalid red status`)
	}
}

func TestDocsCreateDocumentMap(t *testing.T) {
	snapshot := db.Index(index).Document(uid).SetData(dataMap)
	if !snapshot.Success {
		t.Errorf(`SetData not succeded for reason: %s`, snapshot.Reason)
	}

	if snapshot.Doc.Index != index {
		t.Errorf(`Expected doc index to be "%s" but got "%s"`, index, snapshot.Doc.Index)
	}
	if snapshot.Doc.Id != uid {
		t.Errorf(`Expected doc id to be "%s" but got "%s"`, uid, snapshot.Doc.Id)
	}

	if !reflect.DeepEqual(snapshot.Doc.Map(), dataMap) {
		t.Errorf(`Snapshot Map data to be "%v" but got "%v"`, dataMap, snapshot.Doc.Map())
	}

	var example Data
	if err := snapshot.Doc.DataTo(&example); err != nil {
		t.Errorf(`DataTo conversion not succeded for reason: %v`, err)
	}

	if example.Purpose != dataMap["purpose"] {
		t.Errorf(`Snapshot DataTo to be "%v" but got "%v"`, data, example)
	}
}

func TestDocsUpdateDocument(t *testing.T) {
	newPurpose := "another purpose"
	snapshot := db.Index(index).Document(uid).Update([]docs.UpdateData{{Key: "purpose", Value: newPurpose}})
	if !snapshot.Success {
		t.Errorf(`Get not succeded for reason: %s`, snapshot.Reason)
	}
	if snapshot.Doc.Index != index {
		t.Errorf(`Expected doc index to be "%s" but got "%s"`, index, snapshot.Doc.Index)
	}
	if snapshot.Doc.Id != uid {
		t.Errorf(`Expected doc id to be "%s" but got "%s"`, uid, snapshot.Doc.Id)
	}
	var example Data
	if err := snapshot.Doc.DataTo(&example); err != nil {
		t.Errorf(`DataTo conversion not succeded for reason: %v`, err)
	}

	if example.Purpose != newPurpose {
		t.Errorf(`Snapshot DataTo to be "%v" but got "%v"`, newPurpose, example.Purpose)
	}
}

func TestDocsCreateDocumentData(t *testing.T) {
	snapshot := db.Index(index).Document(uid).SetData(data)
	if !snapshot.Success {
		t.Errorf(`SetData not succeded for reason: %s`, snapshot.Reason)
	}

	if snapshot.Doc.Index != index {
		t.Errorf(`Expected doc index to be "%s" but got "%s"`, index, snapshot.Doc.Index)
	}
	if snapshot.Doc.Id != uid {
		t.Errorf(`Expected doc id to be "%s" but got "%s"`, uid, snapshot.Doc.Id)
	}

	var example Data
	if err := snapshot.Doc.DataTo(&example); err != nil {
		t.Errorf(`DataTo conversion not succeded for reason: %v`, err)
	}

	if !reflect.DeepEqual(example, data) {
		t.Errorf(`Snapshot DataTo to be "%v" but got "%v"`, data, example)
	}
}

func TestDocsGetDocument(t *testing.T) {
	snapshot := db.Index(index).Document(uid).Get()
	if !snapshot.Success {
		t.Errorf(`Get not succeded for reason: %s`, snapshot.Reason)
	}

	if snapshot.Doc.Index != index {
		t.Errorf(`Expected doc index to be "%s" but got "%s"`, index, snapshot.Doc.Index)
	}
	if snapshot.Doc.Id != uid {
		t.Errorf(`Expected doc id to be "%s" but got "%s"`, uid, snapshot.Doc.Id)
	}

	var example Data
	if err := snapshot.Doc.DataTo(&example); err != nil {
		t.Errorf(`DataTo conversion not succeded for reason: %v`, err)
	}

	if !reflect.DeepEqual(example, data) {
		t.Errorf(`Snapshot DataTo to be "%v" but got "%v"`, data, example)
	}
}

func TestDocsDeleteDocument(t *testing.T) {
	snapshot := db.Index(index).Document(uid).Delete()
	if !snapshot.Success {
		t.Errorf(`Delete not succeded for reason: %s`, snapshot.Reason)
	}
}

func TestDocsFindNonExistingDocument(t *testing.T) {
	snapshot := db.Index(index).Document(uid).Get()
	if snapshot.Success {
		t.Errorf(`Document Get succeded but does not exist: %s`, snapshot.Reason)
	}

	snapshot = db.Index(index).Document(uid).Update([]docs.UpdateData{{Key: "data", Value: "test"}})
	if snapshot.Success {
		t.Errorf(`Document Update succeded but document does not exist: %s`, snapshot.Reason)
	}

	snapshot = db.Index(index).Document(uid).Delete()
	if snapshot.Success {
		t.Errorf(`Document Delete succeded but document does not exist: %s`, snapshot.Reason)
	}

	indexSnapshot := db.Index(".does.not.exist").Get()
	if indexSnapshot.Success {
		t.Errorf(`Get from non existing index succeded: %s`, indexSnapshot.Reason)
	}
}

func TestDocsCreateDocumentWithoutIndexMap(t *testing.T) {
	snapshot := db.Index(index).Add(dataMap)
	if !snapshot.Success {
		t.Errorf(`Add not succeded for reason: %s`, snapshot.Reason)
	}

	if snapshot.Doc.Index != index {
		t.Errorf(`Expected doc index to be "%s" but got "%s"`, index, snapshot.Doc.Index)
	}

	if !reflect.DeepEqual(snapshot.Doc.Map(), dataMap) {
		t.Errorf(`Snapshot Map data to be "%v" but got "%v"`, dataMap, snapshot.Doc.Map())
	}

	var example Data
	if err := snapshot.Doc.DataTo(&example); err != nil {
		t.Errorf(`DataTo conversion not succeded for reason: %v`, err)
	}

	if example.Purpose != dataMap["purpose"] {
		t.Errorf(`Snapshot DataTo to be "%v" but got "%v"`, data, example)
	}

	snapshot = db.Index(index).Document(snapshot.Doc.Id).Delete()
	if !snapshot.Success {
		t.Errorf(`Delete not succeded for reason: %s`, snapshot.Reason)
	}
}

func TestDocsDeleteIndex(t *testing.T) {
	err := db.Index(index).Delete()
	if err != nil {
		t.Errorf(`Index Delete not succeded for reason: %v`, err)
	}
}

func TestDocsIndexNotThere(t *testing.T) {
	indices, err := db.Get()
	if err != nil {
		t.Errorf(`Index Get not succeded for reason: %v`, err)
	}
	for _, i := range indices {
		if i.Name == index {
			t.Errorf(`Index with name "%s" should not exist`, index)
		}
	}
}
