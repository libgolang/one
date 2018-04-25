package service

import (
	"fmt"
	"os"
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
	ContainerExists(defName string) bool
	ContainerRemove(defName string)
	ContainerGetByDefName(defName string) *model.Container
	ContainerRun(def *model.Definition) *model.Container
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
func NewDocker(host, apiHost, apiVersion string, db Db) Docker {
	_ = os.Setenv("DOCKER_HOST", apiHost)
	_ = os.Setenv("DOCKER_API_VERSION", apiVersion)
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

func (d *docker) ContainerList() []model.Container {
	list, err := d.cli.ContainerList(d.ctx, types.ContainerListOptions{All: true})
	if err != nil {
		panic(err)
	}
	result := make([]model.Container, len(list))
	for _, cont := range list {
		defName, ok := cont.Labels["definitionName"]
		if !ok {
			continue
		}
		modelContainer := model.Container{
			Name:           cont.Names[0],
			ContainerID:    cont.ID,
			Image:          cont.Image,
			Labels:         cont.Labels,
			DefinitionName: defName,
			Running:        cont.State == "running",
		}
		//fmt.Printf("%s", modelContainer)
		result = append(result, modelContainer)
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

func (d *docker) ContainerExists(defName string) bool {
	list := d.ContainerList()
	for _, cont := range list {
		if cont.DefinitionName == defName {
			return true
		}
	}
	return false
}

func (d *docker) ContainerRun(def *model.Definition) *model.Container {
	log.Info("ContainerRun(%s)", def)

	//
	// docker api
	//
	name := fmt.Sprintf("%s-%d", def.Name, d.getNextID())
	image := def.Image

	// labels
	labels := make(map[string]string)
	labels["definitionName"] = def.Name

	// caps
	env := make([]string, 0)
	for k, v := range def.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// porMap : type PortMap map[Port][]PortBinding
	portMap := make(nat.PortMap)
	for _, portPortAndProtocol := range def.Ports { // 53:53/udp
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
	if def.HTTPPort > 0 {
		httpRandPort = d.getNextRandPort()
		port := nat.Port(
			fmt.Sprintf("%s/tcp", strconv.Itoa(def.HTTPPort)),
		)
		portMap[port] = append([]nat.PortBinding{}, nat.PortBinding{HostIP: d.hostIP, HostPort: strconv.Itoa(httpRandPort)})
	}

	// volumes
	volumes := make([]string, 0)
	for hostDir, contDir := range def.Volumes {
		hostDir = strings.TrimSpace(hostDir)
		contDir = strings.TrimSpace(contDir)
		volumes = append(volumes, fmt.Sprintf("%s:%s", hostDir, contDir))
	}

	config := &container.Config{}
	config.Image = def.Image
	config.Env = env
	config.Cmd = def.Cmd
	config.Labels = labels
	hostConfig := &container.HostConfig{}
	hostConfig.CapAdd = def.Caps
	hostConfig.PortBindings = portMap
	hostConfig.Binds = volumes
	netConfig := &network.NetworkingConfig{}

	_, err := d.cli.ImagePull(d.ctx, def.Image, types.ImagePullOptions{})
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

	cont := &model.Container{}
	cont.Name = name
	cont.Image = image
	cont.ContainerID = created.ID
	cont.Labels = labels
	cont.Running = true
	cont.HTTPPort = httpRandPort

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
