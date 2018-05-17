package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strconv"

	"github.com/libgolang/log"
	"github.com/libgolang/one/model"
	"github.com/libgolang/one/utils"
)

const (
	// NodesDir constant holding the directory where node information is stored
	NodesDir = "nodes"
	// DefsDir constant holding the directory where definition information is stored
	DefsDir = "defs"
	// ContsDir constant holding the directory where container information is stored
	ContsDir = "conts"
	// LocksDir constant holding the directory where locks are mantained
	LocksDir = "locks"
)

// Db type
type Db interface {
	ListDefinitions() map[string]*model.Definition
	ListContainers() map[string]*model.Container
	ListNodes() map[string]*model.Node
	GetNode(name string) (*model.Node, error)
	SaveNode(node *model.Node) error
	GetDefinition(name string) (*model.Definition, error)
	GetVars(func(map[string]string)) map[string]string
	DeleteContainer(ID string)
	NextAutoIncrement(ns string, name string) int
	SaveContainer(cont *model.Container) error
	Trx(func(d Db))
	Close()
}

type db struct {
	dir string
	/*
		name       string
		clientURLs string
		peerURLs   string
		cluster    string
		etcd       *embed.Etcd
		ecli       *clientv3.Client
	*/
}

// NewDb Db constructor.
// dir is the path to a directory for storage.
func NewDb(dir /*, masterName, clientURLs, peerURLs, clusterStr*/ string) Db {
	// masterName the name of the master node.  clientURLs comma separated urls to listen for client connections.  e.g. http://10.0.0.1:2380,http://127.0.0.1:2380.  peerURLs comma separated urls to listen for peer connections.  e.g. http://10.0.0.1:2380,http://127.0.0.1:2380.  clusterStr is the cluster connection string with all master server peer addresses of the form "master01=http://10.0.1.10:2380,master02=http://10.0.1.11:2380".
	d := &db{
		dir: dir,
		/*
			name:       masterName,
			clientURLs: clientURLs,
			peerURLs:   peerURLs,
			cluster:    clusterStr,
		*/
	}
	return d
}

func (d *db) Trx(f func(d Db)) {
	f(d)
}

func (d *db) init() {
	// nothing to do
}

func (d *db) Close() {
	// nothing to do
}

func (d *db) ListNodes() map[string]*model.Node {
	result := make(map[string]*model.Node)
	d.listFromDirGeneric(NodesDir, reflect.TypeOf(model.Node{}), func(f string, it interface{}) bool {
		if obj, ok := it.(*model.Node); ok {
			result[obj.Name] = obj
		}
		return true // continue execution
	})

	return result
}

func (d *db) ListDefinitions() map[string]*model.Definition {
	//defer d.Lock(DefsDir)()
	result := make(map[string]*model.Definition)
	d.listFromDirGeneric(DefsDir, reflect.TypeOf(model.Definition{}), func(f string, it interface{}) bool {
		if def, ok := it.(*model.Definition); ok {
			result[def.Name] = def
		}
		return true // continue execution
	})
	return result
}

func (d *db) ListContainers() map[string]*model.Container {
	//defer d.Lock(ContsDir)()
	result := make(map[string]*model.Container)
	d.listFromDirGeneric(ContsDir, reflect.TypeOf(model.Container{}), func(f string, it interface{}) bool {
		if obj, ok := it.(*model.Container); ok {
			result[obj.Name] = obj
		}
		return true // continue execution
	})
	return result
}

func (d *db) GetDefinition(name string) (*model.Definition, error) {
	//defer d.Lock(DefsDir)()
	var found *model.Definition
	d.listFromDirGeneric(DefsDir, reflect.TypeOf(model.Definition{}), func(file string, it interface{}) bool {
		def, ok := it.(*model.Definition)
		if !ok {
			log.Warn("could not cast to model.Definition %s", file)
			return true // continue
		}
		if def.Name == name {
			found = def
			return false
		}
		return true
	})

	var err error
	if found == nil {
		err = fmt.Errorf("Definition %s not found", name)
	}

	return found, err
}

func (d *db) GetVars(cb func(map[string]string)) map[string]string {
	//defer d.Lock("vars-json")()

	varsFile := path.Join(d.dir, "vars.json")

	// Read JSON from file
	var bytes []byte
	var err error
	if _, err := os.Stat(varsFile); os.IsNotExist(err) {
		bytes = []byte("{}") // empty map
	} else {
		bytes, err = ioutil.ReadFile(varsFile)
		if err != nil {
			panic(err)
		}
	}

	// Unmarshal
	var values map[string]string
	if err = json.Unmarshal(bytes, &values); err != nil {
		panic(err)
	}

	// pass to callback
	cb(values)

	// marshall back to file
	if bytes, err = json.Marshal(values); err != nil {
		panic(err)
	}
	if err = ioutil.WriteFile(varsFile, bytes, 0664); err != nil {
		panic(err)
	}

	return values
}

func (d *db) DeleteContainer(name string) {
	collector := func(file string, it interface{}) bool {
		c, ok := it.(*model.Container)
		if ok && c.Name == name {
			return false
		}
		return true
	}
	d.listFromDirGeneric(ContsDir, reflect.TypeOf(model.Container{}), collector)
}

func (d *db) NextAutoIncrement(ns, name string) int {
	key := fmt.Sprintf("%s.%s", ns, name)
	m := d.GetVars(func(m map[string]string) {
		str := m[key]
		n, err := strconv.Atoi(str)
		if err != nil {
			n = 0
		}
		n++
		m[key] = strconv.Itoa(n)
	})
	n, err := strconv.Atoi(m[key])
	if err != nil {
		panic(err)
	}
	return n
}

func (d *db) SaveContainer(cont *model.Container) error {
	bytes, err := json.Marshal(cont)
	if err != nil {
		return err
	}
	dir := d.mkdirIfMissing(ContsDir)
	fileName := path.Join(dir, fmt.Sprintf("%s.json", cont.Name))
	if err = ioutil.WriteFile(fileName, bytes, 0664); err != nil {
		return err
	}
	return nil
}

func (d *db) GetNode(name string) (*model.Node, error) {
	nodes := d.ListNodes()
	node, ok := nodes[name]
	if !ok {
		return nil, fmt.Errorf("Node Not found")
	}
	return node, nil
}

func (d *db) SaveNode(node *model.Node) error {
	log.Info("SavingNode %s", node.Name)
	bytes, err := json.Marshal(node)
	if err != nil {
		return err
	}
	dir := d.mkdirIfMissing(NodesDir)
	fileName := path.Join(dir, fmt.Sprintf("%s.json", node.Name))
	if err = ioutil.WriteFile(fileName, bytes, 0664); err != nil {
		return err
	}
	return nil
}

func (d *db) mkdirIfMissing(subDir string) string {
	dir := path.Join(d.dir, subDir)
	if !utils.FileExists(dir) {
		log.Warn("Creating missing Directory %s", dir)
		utils.Mkdir(dir)
	}
	return dir
}

func (d *db) listFromDirGeneric(subDir string, elementType reflect.Type, collector func(fileName string, it interface{}) bool) {

	// Check Dir
	dir := d.mkdirIfMissing(subDir)

	files, _ := ioutil.ReadDir(dir)
	for _, file := range files {
		// make sure is not dir
		if file.IsDir() {
			continue
		}

		// check file estension
		ext := path.Ext(file.Name())
		if ext != ".json" {
			log.Warn("File has wrong extension %s", ext)
			continue
		}

		// Read File Contents
		fullPath := path.Join(dir, file.Name())
		log.Debug("reading db file %s", fullPath)
		contents, err := ioutil.ReadFile(fullPath)
		if err != nil {
			log.Warn("Unable to read file %s: %s", fullPath, err)
			continue
		}

		//
		fileName := file.Name()
		objPtr := reflect.New(elementType)
		// Unmarshal to the dynamicly created type
		if err := json.Unmarshal(contents, objPtr.Interface()); err != nil {
			log.Warn("Unable to unmarshal definition: %s", fileName)
			continue
		}

		if !collector(fileName, objPtr.Interface()) {
			break
		}
	}
}
