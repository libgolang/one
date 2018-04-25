package model

// Container strcuture
type Container struct {
	Name           string
	Image          string
	ContainerID    string
	Running        bool
	Labels         map[string]string
	HTTPPort       int
	DefinitionName string
	NodeName       string
}
