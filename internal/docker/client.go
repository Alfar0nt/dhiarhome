package docker

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"
)

type Container struct {
	Id     string   `json:"Id"`
	Names  []string `json:"Names"`
	State  string   `json:"State"`
	Status string   `json:"Status"`
}

type Client struct {
	httpClient *http.Client
}

func NewClient(socketPath string) *Client {
	path := strings.TrimPrefix(socketPath, "unix://")
	
	// Create a transport that uses Unix domain sockets
	tr := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial("unix", path)
		},
	}
	
	return &Client{
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   2 * time.Second,
		},
	}
}

// GetContainers fetches the list of containers.
func (c *Client) GetContainers() ([]Container, error) {
	// Docker API endpoint for listing all containers
	resp, err := c.httpClient.Get("http://localhost/containers/json?all=1")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var containers []Container
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil, err
	}

	return containers, nil
}
