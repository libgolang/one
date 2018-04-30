package model

// Container strcuture
type Container struct {
	Name           string            `json:"name"`
	DefinitionName string            `json:"definitionName"`
	Image          string            `json:"image"`
	ContainerID    string            `json:"containerId"`
	Running        bool              `json:"running"`
	Labels         map[string]string `json:"labels"`
	HTTPPort       int               `json:"httpPort"`
	NodeName       string            `json:"nodeName"`
}
