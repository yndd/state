package config

import (
	"strings"
	"testing"
)

var (
	names = []string{"foo", "bar"}
)

func TestAdd(t *testing.T) {

	cs, err := initTestConfigStore()
	if err != nil {
		t.Error(err)
	}

	// get targets from config store
	targets := cs.GetTargets()

	// check len of getTargets equals len of names
	if len(targets) != len(names) {
		t.Errorf("expected %d [%s] targets but got %d [%s]", len(names), strings.Join(names, ", "), len(targets), strings.Join(targets, ", "))
	}

	// check all targets exist
	for _, name := range names {
		if !contains(targets, name) {
			t.Errorf("expected %s to be a valid target, but its missing in [%s]", name, strings.Join(targets, ", "))
		}
	}
}

func TestHasTarget(t *testing.T) {
	cs, err := initTestConfigStore()
	if err != nil {
		t.Error(err)
	}

	// check initilaized targets are there
	for _, name := range names {
		if !cs.HasTarget(name) {
			t.Errorf("no target \"%s\" in configstore although it is expected", name)
		}
	}
	// check non existing targets are not there
	for _, name := range []string{"non-existing-target", "othernonexistingtarget"} {
		if cs.HasTarget(name) {
			t.Errorf("target \"%s\" should not exist but does exist", name)
		}
	}
}

func TestDelete(t *testing.T) {
	cs, err := initTestConfigStore()
	if err != nil {
		t.Error(err)
	}

	// delete a non existing entry
	err = cs.Delete("somenonexistingentry")
	// actually this goes through as a no-op ... so we dont expect any error
	if err != nil {
		t.Error(err)
	}

	// check the delte did not mess up the valid targets
	actualTargets := cs.GetTargets()
	// still all added entries should be in the configstrore
	if len(names) != len(actualTargets) {
		t.Errorf("expected %d [%s] targets but got %d [%s]", len(names), strings.Join(names, ", "), len(actualTargets), strings.Join(actualTargets, ", "))
	}

	expected_size := len(names)
	// delete all the names one after the other
	for _, name := range names {
		expected_size -= 1
		err := cs.Delete(name)
		if err != nil {
			t.Error(err)
		}
	}

	// check configstore is empty
	actualTargets = cs.GetTargets()
	if len(actualTargets) != 0 {
		t.Errorf("expected zero targets but got %d [%s]", len(actualTargets), strings.Join(actualTargets, ", "))
	}
}

func TestGet(t *testing.T) {
	cs, err := initTestConfigStore()
	if err != nil {
		t.Error(err)
	}

	pointers := map[string]ConfigEntry{}
	// retrieve entries, store in map
	for _, name := range names {
		entry, err := cs.Get(name)
		if err != nil {
			t.Error(err)
		}
		pointers[name] = entry
	}

	// make sure the entries are still the same on consecutive calls
	for _, name := range names {
		entry, err := cs.Get(name)
		if err != nil {
			t.Error(err)
		}
		if entry != pointers[name] {
			t.Errorf("retrieved different configentries on consecutive calls to get")
		}
	}
}

// initTestConfigStore initializes a ConfigStore with the entries of the names slice
func initTestConfigStore() (ConfigStore, error) {
	cs := NewConfigStore()

	for _, name := range names {
		err := cs.Add(name)
		if err != nil {
			return nil, err
		}
	}
	return cs, nil
}

// contains checks if a string is present as an entry in a string slice
func contains(s []string, searchterm string) bool {
	for _, entry := range s {
		if entry == searchterm {
			return true
		}
	}
	return false
}
