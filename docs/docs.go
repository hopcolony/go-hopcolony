package docs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"hopcolony.io/hopcolony/initialize"
)

type ClusterHealth struct {
	ClusterName   string `json:"cluster_name"`
	Status        string `json:"status"`
	NumberOfNodes int    `json:"number_of_nodes"`
}

func New() (HopDoc, error) {
	project := initialize.GetProject()
	client, err := newHopDocClient(project, "docs.hopcolony.io", 443)
	if err != nil {
		return HopDoc{}, err
	}

	return HopDoc{project, client}, nil
}

type HopDoc struct {
	Project initialize.Project
	client  HopDocClient
}

func (h *HopDoc) Close() {
	h.client.close()
}

func (h *HopDoc) Status() (string, error) {
	bytes, err := h.client.Get("/_cluster/health")
	if err != nil {
		return "", err
	}

	var clusterHealth ClusterHealth
	err = json.Unmarshal(bytes, &clusterHealth)
	if err != nil {
		return "", err
	}

	return clusterHealth.Status, nil
}

func (h *HopDoc) Index(index string) *IndexReference {
	return &IndexReference{h.client, index}
}

func (h *HopDoc) Get() ([]Index, error) {
	indices := make([]Index, 0)
	resp, err := h.client.Get("/_cluster/health?level=indices")
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	json.Unmarshal(resp, &result)
	for name, status := range result["indices"].(map[string]interface{}) {
		index := Index{Name: name}
		index.Status = status.(map[string]interface{})["status"].(string)
		num_docs, err := h.Index(name).Count()
		if err != nil {
			return nil, err
		}
		index.NumDocs = num_docs
		indices = append(indices, index)
	}
	return indices, nil
}

type HopDocClient struct {
	Project    initialize.Project
	Host       string
	Port       int
	identity   string
	baseUrl    string
	httpClient *http.Client
}

func (h *HopDocClient) close() {}

func (h *HopDocClient) Get(path string) ([]byte, error) {
	req, err := http.NewRequest("GET", h.baseUrl+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Token", h.Project.Config.Token)
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return body, nil
	}
	return nil, errors.New("Status code for GET request was: " + strconv.Itoa(resp.StatusCode))
}

func (h *HopDocClient) Post(path string, body []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", h.baseUrl+path, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Token", h.Project.Config.Token)
	req.Header.Add("Content-type", "application/json")
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		return respBody, nil
	}
	return nil, fmt.Errorf("POST with error: %s", respBody)
}

func (h *HopDocClient) Put(path string, body []byte) ([]byte, error) {
	req, err := http.NewRequest("PUT", h.baseUrl+path, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Token", h.Project.Config.Token)
	req.Header.Add("Content-type", "application/json")
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		return respBody, nil
	}
	return nil, fmt.Errorf("PUT with error: %s", respBody)
}

func (h *HopDocClient) Delete(path string) error {
	req, err := http.NewRequest("DELETE", h.baseUrl+path, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Token", h.Project.Config.Token)
	req.Header.Add("Content-type", "application/json")
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		return nil
	}
	return fmt.Errorf("DELETE with error: %s", respBody)
}

func newHopDocClient(project initialize.Project, host string, port int) (HopDocClient, error) {
	return HopDocClient{Project: project, Host: host, Port: port, identity: project.Config.Identity,
		baseUrl: fmt.Sprintf("https://%s:%d/%s/api", host, port, project.Config.Identity), httpClient: &http.Client{}}, nil
}
