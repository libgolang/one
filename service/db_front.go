package service

import (
	"github.com/libgolang/log"
	"github.com/libgolang/one/model"
)

type front struct {
	db   Db
	jobs chan func(d Db)
}

// NewFrontDb constructor
func NewFrontDb(db Db) Db {
	f := &front{
		db:   db,
		jobs: make(chan func(Db)),
	}
	f.init()
	return f
}

func (f *front) init() {
	go func() {
		i := 0
		for cb := range f.jobs {
			log.Debug("Queue(%d): Transaction received ...", i)
			cb(f.db)
			log.Debug("Queue(%d): ... Transaction done", i)
			i++
		}
	}()
}

func (f *front) Trx(cb func(d Db)) {
	done := make(chan int)
	f.jobs <- func(d Db) {
		cb(d)
		done <- 1
	}
	<-done
}

func (f *front) Close() {
	f.db.Close()
}

func (f *front) ListNodes() map[string]*model.Node {
	var list map[string]*model.Node
	f.Trx(func(d Db) {
		list = d.ListNodes()
	})
	return list
}

func (f *front) ListDefinitions() map[string]*model.Definition {
	var list map[string]*model.Definition
	f.Trx(func(d Db) {
		list = d.ListDefinitions()
	})
	return list
}

func (f *front) ListContainers() map[string]*model.Container {
	var list map[string]*model.Container
	f.Trx(func(d Db) {
		list = d.ListContainers()
	})
	return list
}

func (f *front) GetDefinition(name string) (*model.Definition, error) {
	var list *model.Definition
	var err error
	f.Trx(func(d Db) {
		list, err = d.GetDefinition(name)
	})
	return list, err
}

func (f *front) GetVars(cb func(map[string]string)) map[string]string {
	log.Debug("Front GetVars Start")
	var res map[string]string
	f.Trx(func(d Db) {
		log.Debug("db GetVars Start")
		res = d.GetVars(cb)
		log.Debug("db GetVars Start")
	})
	log.Debug("Front GetVars End")
	return res
}

func (f *front) DeleteContainer(name string) {
	f.Trx(func(d Db) {
		d.DeleteContainer(name)
	})
}

func (f *front) NextAutoIncrement(ns, name string) int {
	var res int
	f.Trx(func(d Db) {
		res = d.NextAutoIncrement(ns, name)
	})
	return res
}

func (f *front) SaveContainer(cont *model.Container) error {
	var err error
	f.Trx(func(d Db) {
		err = d.SaveContainer(cont)
	})
	return err
}

func (f *front) GetNode(name string) (*model.Node, error) {
	var node *model.Node
	var err error
	f.Trx(func(d Db) {
		node, err = d.GetNode(name)
	})
	return node, err
}

func (f *front) SaveNode(node *model.Node) error {
	var err error
	f.Trx(func(d Db) {
		err = d.SaveNode(node)
	})
	return err
}
