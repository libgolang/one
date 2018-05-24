package model

// Container strcuture
type Container struct {
	Name           string            `json:"name"`
	DefinitionName string            `json:"definitionName"`
	Image          string            `json:"image"`
	NodeName       string            `json:"nodeName"`
	ContainerID    string            `json:"containerId"`
	Running        bool              `json:"running"`
	Labels         map[string]string `json:"labels"`
	Volumes        map[string]string `json:"volumes"`
	HTTPPort       int               `json:"httpPort"`     // the HTTP Port on the container
	NodeHTTPPort   int               `json:"nodeHttpPort"` // the rnadomly generated HTTP Port to access HTTPPort
	Ports          []string          `json:"ports"`
	Env            map[string]string `json:"env"`
	Cmd            []string          `json:"cmd"`
	Caps           []string          `json:"caps"`
}
