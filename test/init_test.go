package test

import (
	"hopcolony.io/hopcolony/initialize"
	"testing"
)

func TestInitialize(t *testing.T) {
	_, err := initialize.Initialize(initialize.ProjectConfig{ConfigFile: ".."})

	if err != initialize.ConfigNotFound {
		t.Error("Opened project with non-existing file path")
	}

	_, err = initialize.Initialize(initialize.ProjectConfig{Username: "username"})

	if err != initialize.InvalidConfig {
		t.Error("Created project with only username")
	}

	_, err = initialize.Initialize(initialize.ProjectConfig{Username: "username", Project: "project"})

	if err != initialize.InvalidConfig {
		t.Error("Created project with only username and project")
	}

	project, err := initialize.Initialize(initialize.ProjectConfig{Username: "username", Project: "project", Token: "token"})

	if err != nil {
		t.Errorf("Error in project creation: %s", err)
	}

	if project.Config.Username != "username" {
		t.Errorf(`Expected username to be "username" but got "%s"`, project.Config.Username)
	}

	if project.Config.Project != "project" {
		t.Errorf(`Expected project to be "project" but got "%s"`, project.Config.Project)
	}

	if project.Config.Token != "token" {
		t.Errorf(`Expected token to be "token" but got "%s"`, project.Config.Token)
	}
}
