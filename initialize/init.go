package initialize

import (
	b64 "encoding/base64"
	"errors"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v3"
)

type HopConfig struct {
	Username string `yaml:"username"`
	Project  string `yaml:"project"`
	Token    string `yaml:"token"`
	Identity string
}

func getNamespace(username string) string {
	encodedUsername := b64.StdEncoding.EncodeToString([]byte(username))
	lowerEncodedUsername := strings.ToLower(encodedUsername)

	return "a" + strings.Replace(lowerEncodedUsername, "=", "-", -1) + "a"
}

func computeIdentity(username, project string) string {
	data := getNamespace(username) + "." + project
	return b64.StdEncoding.EncodeToString([]byte(data))
}

func newHopConfig(username, project, token string) HopConfig {
	identity := computeIdentity(username, project)
	return HopConfig{
		Username: username,
		Project:  project,
		Token:    token,
		Identity: identity,
	}
}

func newHopConfigFromFile(filename string) (*HopConfig, error) {
	file, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	c := &HopConfig{}
	err = yaml.Unmarshal(file, c)

	if err != nil {
		return nil, err
	}

	return c, nil
}

var InvalidConfig error = errors.New("If you provide one of [username, project, token] or [namespace, project, token], you need to provide the 3 of them")
var ConfigNotFound error = errors.New("Hop Config not found. Run 'hopctl login' or place a .hop.config file here.")

type ProjectConfig struct {
	Username   string
	Project    string
	Token      string
	ConfigFile string
}
type Project struct {
	Config HopConfig
}

func newProject(config ProjectConfig) (*Project, error) {
	if config.Username != "" || config.Project != "" || config.Token != "" {
		if config.Username != "" && config.Project != "" && config.Token != "" {
			return &Project{Config: newHopConfig(
				config.Username,
				config.Project,
				config.Token,
			)}, nil
		} else {
			return nil, InvalidConfig
		}
	} else {
		config, err := newHopConfigFromFile(config.ConfigFile)

		if err != nil {
			return nil, ConfigNotFound
		}

		return &Project{Config: *config}, nil

	}
}

var project *Project

func Initialize(config ProjectConfig) (*Project, error) {
	return newProject(config)
}
