package service

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/libgolang/one/model"
)

func TestVarsWithNoData(t *testing.T) {
	// given
	tmpDir, _ := ioutil.TempDir("", "testing-db")
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// when
	d := NewDb(tmpDir)
	m := d.GetVars(func(m map[string]string) {})

	if len(m) > 0 {
		t.Error("map should be empty")
	}
}

func TestVarsWithData(t *testing.T) {
	// given
	tmpDir, _ := ioutil.TempDir("", "testing-db")
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// when
	d := NewDb(tmpDir)
	m := d.GetVars(func(m map[string]string) {
		m["xyz1"] = "xyz1"
	})

	m = d.GetVars(func(m map[string]string) {
		m["xyz2"] = "xyz2"
	})

	if len(m) != 2 {
		t.Errorf("Map should have two items, instead the count is %d", len(m))
	}
	if _, ok := m["xyz1"]; !ok {
		t.Error("Map should have an xyz1 key")
	}
	if _, ok := m["xyz2"]; !ok {
		t.Error("Map should have an xyz2 key")
	}
}

func TestListContainers(t *testing.T) {
	// given
	tmpDir, _ := ioutil.TempDir("", "testing-db")
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// when
	d := NewDb(tmpDir)
	containers := d.ListContainers()

	if len(containers) != 0 {
		t.Error("Should not have any definitions")
	}
}

func TestListNodes(t *testing.T) {
	// given
	tmpDir, _ := ioutil.TempDir("", "testing-db")
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// when
	d := NewDb(tmpDir)

	node := &model.Node{}
	node.Addr = "127.0.0.1"
	node.Name = "xyz"

	err := d.SaveNode(node)
	if err != nil {
		t.Errorf("error saving node: %s", err)
	}
	nodes := d.ListNodes()

	if len(nodes) != 1 {
		t.Error("Should have one node")
	}
}
