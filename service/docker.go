package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/libgolang/log"
	"github.com/libgolang/one/model"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"golang.org/x/net/context"
)

const (
	initPort = 11000
	initID   = 0
)

// Docker interface
type Docker interface {
	ContainerList() []model.Container
	IsRunningByDefName(defName string) bool
	//IsRunningByName(defName string) bool
	ContainerExists(defName string) bool
	ContainerRemove(defName string)
	ContainerRemoveByName(defName string)
	ContainerGetByDefName(defName string) *model.Container
	ContainerGetByName(name string) *model.Container
	ContainerRunByDefinition(def *model.Definition) *model.Container
	ContainerRun(def *model.Container)
	ContainerStopByDefName(defName string)
	ContainerRemoveByDefName(defName string)
}

type docker struct {
	hostIP string
	ctx    context.Context
	cli    *client.Client
	db     Db
}

// NewDocker Docker constructor host is the ip or host name
// used to access the hsot. apiHost is the ip or host name of
// the docker api (empty string to use unix socket).  apiVersion
// used to match the api version on the docker server.
func NewDocker(host /*, apiHost, apiVersion*/ string, db Db) Docker {
	//_ = os.Setenv("DOCKER_HOST", apiHost)
	//_ = os.Setenv("DOCKER_API_VERSION", apiVersion)
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	return &docker{host, ctx, cli, db}
}

func (d *docker) ContainerRemove(name string) {
	log.Info("ContainerRemove(%s)", name)
	c := d.ContainerGetByDefName(name)
	if c != nil {
		_ = d.cli.ContainerRemove(d.ctx, c.ContainerID, types.ContainerRemoveOptions{})
	}
}

func (d *docker) ContainerRemoveByName(name string) {
	log.Info("ContainerRemove(%s)", name)
	c := d.ContainerGetByName(name)
	if c != nil {
		_ = d.cli.ContainerRemove(d.ctx, c.ContainerID, types.ContainerRemoveOptions{})
	}
}

// ContainerGetByDefName gets a container or nil if not found
func (d *docker) ContainerGetByDefName(defName string) *model.Container {
	list := d.ContainerList()
	for _, c := range list {
		if c.DefinitionName == defName {
			return &c
		}
	}
	return nil
}
func (d *docker) ContainerGetByName(name string) *model.Container {
	list := d.ContainerList()
	for _, c := range list {
		if c.Name == name {
			return &c
		}
	}
	return nil
}

func (d *docker) ContainerList() []model.Container {
	result := make([]model.Container, 0)

	list, err := d.cli.ContainerList(d.ctx, types.ContainerListOptions{All: true})
	if err != nil {
		log.Error("Error listing container from docker daemon: %s", err)
	} else {
		for _, cont := range list {
			_, ok := cont.Labels["one.managed"]
			if !ok {
				continue
			}
			modelContainer := model.Container{
				Name:           string([]rune(cont.Names[0])[1:]),
				ContainerID:    cont.ID,
				Image:          cont.Image,
				Labels:         cont.Labels,
				DefinitionName: cont.Labels["one.definitionName"],
				Running:        cont.State == "running",
			}
			//log.Debug("Found: %s", modelContainer)
			result = append(result, modelContainer)
		}
	}
	return result
}

func (d *docker) IsRunningByDefName(defName string) bool {
	list := d.ContainerList()
	for _, cont := range list {
		if cont.DefinitionName == defName && cont.Running {
			//fmt.Printf("Container is running: %s\n", defName)
			return true
		}
	}
	return false
}

/*
func (d *docker) IsRunningByName(name string) bool {
	list := d.ContainerList()
	var isRunning = false
	for _, cont := range list {
		if cont.Name == name && cont.Running {
			isRunning = true
		}
	}
	log.Debug("IsRunningByName(%s) %t", name, isRunning)
	return isRunning
}
*/

func (d *docker) ContainerExists(defName string) bool {
	list := d.ContainerList()
	for _, cont := range list {
		if cont.DefinitionName == defName {
			return true
		}
	}
	return false
}

func (d *docker) ContainerRunByDefinition(def *model.Definition) *model.Container {
	log.Info("ContainerRunByDefinition(%s)", def)

	cont := &model.Container{}
	cont.Image = def.Image
	cont.DefinitionName = def.Name
	cont.HTTPPort = d.db.NextAutoIncrement("http.port", "http.port")
	cont.Name = fmt.Sprintf("%s-%d", def.Name, d.db.NextAutoIncrement("inc.container", def.Name))
	cont.Running = false

	d.ContainerRun(cont)
	return cont
}

func (d *docker) ContainerStopByDefName(defName string) {
	log.Info("CotainerStopByDefName(%s)", defName)
	if cont := d.ContainerGetByDefName(defName); cont != nil {
		var dur = time.Second * 10
		if err := d.cli.ContainerStop(d.ctx, cont.ContainerID, &dur); err != nil {
			panic(err)
		}
	}
}

func (d *docker) ContainerRemoveByDefName(defName string) {
	log.Info("ContainerRemoveByDefName(%s)", defName)
	if cont := d.ContainerGetByDefName(defName); cont != nil && !cont.Running {
		opts := types.ContainerRemoveOptions{}
		err := d.cli.ContainerRemove(d.ctx, cont.ContainerID, opts)
		if err != nil {
			panic(err)
		}
	}
}

func (d *docker) getNextRandPort() int {
	m := d.db.GetVars(func(m map[string]string) {
		portStr, ok := m["lastHttpPort"]
		if !ok {
			portStr = strconv.Itoa(initPort - 1)
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			panic(err)
		}

		port++

		m["lastHttpPort"] = strconv.Itoa(port)
	})

	portStr := m["lastHttpPort"]

	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(err)
	}
	log.Debug("getNextRandPort():%d", port)
	return port
}

func (d *docker) getNextID() int {
	m := d.db.GetVars(func(m map[string]string) {
		str, ok := m["lastId"]
		if !ok {
			str = strconv.Itoa(initID - 1)
		}

		id, err := strconv.Atoi(str)
		if err != nil {
			panic(err)
		}

		id++
		m["lastId"] = strconv.Itoa(id)
	})

	str := m["lastId"]

	id, err := strconv.Atoi(str)
	if err != nil {
		panic(err)
	}
	log.Debug("getNextID():%d", id)
	return id
}

func (d *docker) ContainerRun(cont *model.Container) {
	log.Info("ContainerRun(%s)", cont)

	//
	// docker api
	//
	name := cont.Name

	// labels
	labels := make(map[string]string)
	labels["one.definitionName"] = cont.DefinitionName
	labels["one.managed"] = "true"

	// caps
	env := make([]string, 0)
	for k, v := range cont.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// porMap : type PortMap map[Port][]PortBinding
	portMap := make(nat.PortMap)
	for _, portPortAndProtocol := range cont.Ports { // 53:53/udp
		parts := strings.Split(portPortAndProtocol, ":")
		hostPort := parts[0]
		portAndProtocol := parts[1]
		port := nat.Port(portAndProtocol)
		binding := nat.PortBinding{HostIP: d.hostIP, HostPort: hostPort}
		if bindings, ok := portMap[port]; ok {
			portMap[port] = append(bindings, binding)
		} else {
			portMap[port] = append([]nat.PortBinding{}, binding)
		}
	}
	// http port
	httpRandPort := 0
	if cont.HTTPPort > 0 {
		httpRandPort = d.getNextRandPort()
		port := nat.Port(
			fmt.Sprintf("%s/tcp", strconv.Itoa(cont.HTTPPort)),
		)
		portMap[port] = append([]nat.PortBinding{}, nat.PortBinding{HostIP: d.hostIP, HostPort: strconv.Itoa(httpRandPort)})
	}

	// volumes
	volumes := make([]string, 0)
	for hostDir, contDir := range cont.Volumes {
		hostDir = strings.TrimSpace(hostDir)
		contDir = strings.TrimSpace(contDir)
		volumes = append(volumes, fmt.Sprintf("%s:%s", hostDir, contDir))
	}

	config := &container.Config{}
	config.Image = cont.Image
	config.Env = env
	config.Cmd = cont.Cmd
	config.Labels = labels
	hostConfig := &container.HostConfig{}
	hostConfig.CapAdd = cont.Caps
	hostConfig.PortBindings = portMap
	hostConfig.Binds = volumes
	netConfig := &network.NetworkingConfig{}

	_, err := d.cli.ImagePull(d.ctx, cont.Image, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	//_, _ = io.Copy(os.Stdout, reader)

	created, err := d.cli.ContainerCreate(d.ctx, config, hostConfig, netConfig, name)
	if err != nil {
		panic(err)
	}
	if err = d.cli.ContainerStart(d.ctx, created.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	if inspect, err := d.cli.ContainerInspect(d.ctx, created.ID); err != nil {
		panic(err)
	} else if !inspect.State.Running {
		panic(fmt.Sprintf("Container %s not running", created.ID))
	}

	//cont := &model.Container{}
	//cont.Name = name
	//cont.Image = image
	cont.ContainerID = created.ID
	//cont.Labels = labels
	cont.Running = true
	cont.HTTPPort = httpRandPort
	//return cont
}
