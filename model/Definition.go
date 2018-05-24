package model

// Definition model
type Definition struct {
	Name     string            `json:"name"`
	Image    string            `json:"image"`
	Count    int               `json:"count"`
	HTTPPort int               `json:"httpPort"` // container port. it will be mapped to nodehttpport in Container model
	Ports    []string          `json:"ports"`
	Volumes  map[string]string `json:"volumes"`
	Env      map[string]string `json:"env"`
	Caps     []string          `json:"caps"`
	Cmd      []string          `json:"cmd"`
}
