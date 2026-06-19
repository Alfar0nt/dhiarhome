package docker

// Container represents a Docker container (mock/demo only).
type Container struct {
	Id     string   `json:"Id"`
	Names  []string `json:"Names"`
	State  string   `json:"State"`
	Status string   `json:"Status"`
}

// MockContainers returns a hardcoded list of containers for demo purposes.
func MockContainers() []Container {
	return []Container{
		{Names: []string{"/nginx"}, State: "running", Status: "Up 2 days"},
		{Names: []string{"/pihole"}, State: "exited", Status: "Exited (0) 5 hours ago"},
		{Names: []string{"/portainer"}, State: "running", Status: "Up 14 days"},
		{Names: []string{"/plex"}, State: "running", Status: "Up 7 days"},
		{Names: []string{"/nextcloud"}, State: "running", Status: "Up 3 days"},
	}
}
