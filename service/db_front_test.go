package service

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestDbFrontVars(t *testing.T) {
	// given
	tmpDir, _ := ioutil.TempDir("", "testing-db")
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// when
	d := NewFrontDb(NewDb(tmpDir))
	m := d.GetVars(func(m map[string]string) {})

	if len(m) > 0 {
		t.Error("map should be empty")
	}
}

func TestSubsequentVarUpdates(t *testing.T) {
	// given
	tmpDir, _ := ioutil.TempDir("", "testing-db")
	defer func() { _ = os.RemoveAll(tmpDir) }()
	t.Logf("TestSubsequentVarUpdates tmp dir: %s", tmpDir)

	// when
	d := NewFrontDb(NewDb(tmpDir))
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

func TestDbFrontListContainers(t *testing.T) {
	// given
	tmpDir, _ := ioutil.TempDir("", "testing-db")
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// when
	d := NewFrontDb(NewDb(tmpDir))
	containers := d.ListContainers()

	if len(containers) != 0 {
		t.Error("Should not have any definitions")
	}
}
