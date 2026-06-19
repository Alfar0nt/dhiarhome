package docker

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
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
	httpClient   *http.Client
	baseURL      string
	portainerURL string
	portainerKey string
	portainerEnv int
	usePortainer bool
}

// Options holds configuration for creating a Docker client.
type Options struct {
	Endpoint string
	SkipTLS  bool
	CACert   string
	Cert     string
	Key      string
	// Portainer
	PortainerURL string
	PortainerKey string
	PortainerEnv int
}

// NewClient creates a Docker client with the given endpoint string.
// Kept for backward compatibility.
func NewClient(endpoint string) *Client {
	return NewClientWithOptions(Options{Endpoint: endpoint})
}

// NewClientWithOptions creates a Docker client with full TLS and Portainer support.
// Connection priority: Portainer > Remote Docker (TCP/TLS) > Local socket
func NewClientWithOptions(opts Options) *Client {
	c := &Client{}

	// Portainer mode takes priority
	if opts.PortainerURL != "" && opts.PortainerKey != "" {
		c.portainerURL = strings.TrimRight(opts.PortainerURL, "/")
		c.portainerKey = opts.PortainerKey
		c.portainerEnv = opts.PortainerEnv
		c.usePortainer = true
		c.httpClient = &http.Client{Timeout: 10 * time.Second}

		// If Portainer URL uses https with skip_tls
		if strings.HasPrefix(opts.PortainerURL, "https://") && opts.SkipTLS {
			c.httpClient.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
		}
		return c
	}

	// Standard Docker client (socket/tcp/tls)
	var tr *http.Transport
	baseURL := "http://localhost" // default for unix socket

	tlsConfig := &tls.Config{}
	hasTLS := false

	// Load TLS certificates if provided
	if opts.CACert != "" {
		caCert, err := os.ReadFile(opts.CACert)
		if err == nil {
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.RootCAs = caCertPool
			hasTLS = true
		}
	}

	if opts.Cert != "" && opts.Key != "" {
		cert, err := tls.LoadX509KeyPair(opts.Cert, opts.Key)
		if err == nil {
			tlsConfig.Certificates = []tls.Certificate{cert}
			hasTLS = true
		}
	}

	if opts.SkipTLS {
		tlsConfig.InsecureSkipVerify = true
		hasTLS = true
	}

	if strings.HasPrefix(opts.Endpoint, "unix://") {
		path := strings.TrimPrefix(opts.Endpoint, "unix://")
		tr = &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", path)
			},
		}
	} else if strings.HasPrefix(opts.Endpoint, "tcp://") {
		if hasTLS {
			// Convert tcp:// to https:// when TLS is configured
			baseURL = strings.Replace(opts.Endpoint, "tcp://", "https://", 1)
			tr = &http.Transport{TLSClientConfig: tlsConfig}
		} else {
			baseURL = strings.Replace(opts.Endpoint, "tcp://", "http://", 1)
			tr = &http.Transport{}
		}
	} else if strings.HasPrefix(opts.Endpoint, "http://") || strings.HasPrefix(opts.Endpoint, "https://") {
		baseURL = opts.Endpoint
		if strings.HasPrefix(opts.Endpoint, "https://") && hasTLS {
			tr = &http.Transport{TLSClientConfig: tlsConfig}
		} else {
			tr = &http.Transport{}
		}
	} else {
		// Fallback assuming it's a raw socket path
		path := opts.Endpoint
		tr = &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", path)
			},
		}
	}

	c.httpClient = &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
	}
	c.baseURL = strings.TrimRight(baseURL, "/")
	return c
}

// GetContainers fetches the list of containers from Docker or Portainer.
func (c *Client) GetContainers() ([]Container, error) {
	if c.usePortainer {
		return c.getPortainerContainers()
	}
	return c.getDockerContainers()
}

// getDockerContainers fetches containers directly from the Docker API.
func (c *Client) getDockerContainers() ([]Container, error) {
	url := c.baseURL + "/containers/json?all=1"
	resp, err := c.httpClient.Get(url)
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

// getPortainerContainers fetches containers via the Portainer API.
func (c *Client) getPortainerContainers() ([]Container, error) {
	envID := c.portainerEnv
	if envID == 0 {
		envID = 1 // default to endpoint 1
	}
	url := fmt.Sprintf("%s/api/endpoints/%d/docker/containers/json?all=1", c.portainerURL, envID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", c.portainerKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("portainer API returned status %d", resp.StatusCode)
	}

	var containers []Container
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil, err
	}
	return containers, nil
}
