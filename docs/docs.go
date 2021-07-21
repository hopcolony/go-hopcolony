package docs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"hopcolony.io/hopcolony/initialize"
)

type ClusterHealth struct {
	ClusterName   string `json:"cluster_name"`
	Status        string `json:"status"`
	NumberOfNodes int    `json:"number_of_nodes"`
}

func New() (*HopDoc, error) {
	project := initialize.GetProject()
	client, err := newHopDocClient(project, "docs.hopcolony.io", 443)
	if err != nil {
		return nil, err
	}

	return &HopDoc{project, client}, nil
}

type HopDoc struct {
	Project initialize.Project
	client  HopDocClient
}

func (h *HopDoc) Close() {
	h.client.close()
}

func (h *HopDoc) Status() (string, error) {
	b, err := h.client.Get("/_cluster/health")
	if err != nil {
		return "", err
	}

	var clusterHealth ClusterHealth
	err = json.Unmarshal(b, &clusterHealth)
	if err != nil {
		return "", err
	}

	return clusterHealth.Status, nil
}

func (h *HopDoc) Index(index string) *IndexReference {
	return &IndexReference{h.client, index, make([]Query, 0)}
}

func (h *HopDoc) Get() ([]Index, error) {
	indices := make([]Index, 0)
	resp, err := h.client.Get("/_cluster/health?level=indices")
	if err != nil {
		return nil, fmt.Errorf("could not get cluster's health: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(resp, &result)
	dotr, _ := regexp.Compile(`\..*`)
	ilmr, _ := regexp.Compile(`lm-history-.*`)
	for name, status := range result["indices"].(map[string]interface{}) {
		// Filter indices starting with . or ilm-history
		if dotr.MatchString(name) || ilmr.MatchString(name) {
			continue
		}
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
		return nil, fmt.Errorf("could not do GET request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code of GET is %s", resp.Status)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read body in GET: %v", err)
	}
	return body, nil
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
		return nil, fmt.Errorf("could not do POST request: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("status code of POST is %s", resp.Status)
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read body in POST: %v", err)
	}

	return b, nil
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
		return nil, fmt.Errorf("could not do PUT request: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("status code of PUT is %s", resp.Status)
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read body in PUT: %v", err)
	}

	return b, nil
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
		return fmt.Errorf("could not do DELETE request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code of PUT is %s", resp.Status)
	}

	return nil
}

func newHopDocClient(project initialize.Project, host string, port int) (HopDocClient, error) {
	return HopDocClient{Project: project, Host: host, Port: port, identity: project.Config.Identity,
		baseUrl: fmt.Sprintf("https://%s:%d/%s/api", host, port, project.Config.Identity), httpClient: &http.Client{}}, nil
}
