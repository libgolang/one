package service

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestVars(t *testing.T) {
	// given
	tmpDir, _ := ioutil.TempDir("", "testing-db")
	defer func() {
		os.RemoveAll(tmpDir)
	}()

	// when
	d := NewDb(tmpDir)
	m := d.GetVars(func(m map[string]string) {})

	if len(m) > 0 {
		t.Error("map should be empty")
	}
}

func TestListContainers(t *testing.T) {
	// given
	tmpDir, _ := ioutil.TempDir("", "testing-db")
	defer func() {
		os.RemoveAll(tmpDir)
	}()

	// when
	d := NewDb(tmpDir)
	containers := d.ListContainers()

	if len(containers) != 0 {
		t.Error("Should not have any definitions")
	}
}
