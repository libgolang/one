package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"syscall"

	"github.com/rhamerica/one/model"
	"github.com/rhamerica/one/utils"
)

// Db type
type Db interface {
	ListDefinitions() []model.Definition
	ListContainers() []model.Container
	GetDefinition(name string) (*model.Definition, error)
	GetVars(func(map[string]string)) map[string]string
}

type db struct {
	dir          string
	defDir       string
	containerDir string
	varsFile     string
}

// NewDb constructor
func NewDb(dir string) Db {
	d := &db{
		dir:          dir,
		defDir:       path.Join(dir, "defs"),
		varsFile:     path.Join(dir, "vars.json"),
		containerDir: path.Join(dir, "cont"),
	}
	return d
}

func (d *db) ListDefinitions() []model.Definition {
	result := make([]model.Definition, 0)
	obj := model.Definition{}
	d.listFromDir(
		d.defDir,
		obj,
		func(it interface{}) {
			result = append(result, it.(model.Definition))
		},
	)
	return result
}

func (d *db) ListContainers() []model.Container {
	result := make([]model.Container, 0)
	obj := model.Definition{}
	collector := func(it interface{}) {
		result = append(result, it.(model.Container))
	}
	d.listFromDir(d.containerDir, obj, collector)
	return result
}

func (d *db) listFromDir(dir string, it interface{}, collector func(i interface{})) {
	if !utils.FileExists(dir) {
		fmt.Println("Definition directory() does not exist.  Creating!", dir)
		utils.Mkdir(dir)
	}

	files, _ := ioutil.ReadDir(dir)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if path.Ext(file.Name()) != ".json" {
			continue
		}
		contents, err := ioutil.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			panic(err)
		}
		v := reflect.Indirect(reflect.ValueOf(it)).Interface() // clone
		if err = json.Unmarshal(contents, v); err != nil {
			panic(err)
		}
		collector(v)
	}
}

func (d *db) GetDefinition(name string) (*model.Definition, error) {
	for _, def := range d.ListDefinitions() {
		if def.Name == name {
			return &def, nil
		}
	}
	return nil, fmt.Errorf("Definition %s not found", name)
}

func (d *db) GetVars(cb func(map[string]string)) map[string]string {
	f, err := os.OpenFile(d.varsFile, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "error closing file %s", d.varsFile)
		}
	}()
	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
	if err != nil {
		panic(err)
	}

	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	if len(bytes) == 0 {
		bytes = []byte("{}") // empty map
	}

	var values map[string]string
	if err = json.Unmarshal(bytes, &values); err != nil {
		panic(err)
	}

	cb(values)

	if bytes, err = json.Marshal(values); err != nil {
		panic(err)
	}

	// override file contents
	_, _ = f.Seek(0, 0)
	_ = f.Truncate(0)
	if _, err = f.Write(bytes); err != nil {
		panic(err)
	}

	return values
}
