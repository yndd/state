package config

import (
	"fmt"
	"sync"
)

type ConfigStore interface {
	//Lock()
	//Unlock()
	HasTarget(target string) bool
	Get(target string) (ConfigEntry, error)
	Add(target string) error
	Delete(target string) error
	GetTargets() []string
}

type configStore struct {
	m sync.RWMutex

	config map[string]ConfigEntry
}

// NewConfigStore creates a new Config map
func NewConfigStore() ConfigStore {
	return &configStore{
		config: map[string]ConfigEntry{},
	}
}

// HasTarget checks if the given target is present in the config map
func (c *configStore) HasTarget(target string) bool {
	c.m.Lock()
	defer c.m.Unlock()
	_, ok := c.config[target]
	return ok
}

// Get retrieves the target configuration. Throws an error if the target does not exist
func (c *configStore) Get(target string) (ConfigEntry, error) {
	c.m.RLock()
	defer c.m.RUnlock()
	tc, ok := c.config[target]
	if !ok {
		return nil, fmt.Errorf("config target '%s' does not exist", target)
	}
	return tc, nil
}

// Add creates a new empty ConfigEntry referenced under the given target
func (c *configStore) Add(target string) error {
	c.m.Lock()
	defer c.m.Unlock()
	c.config[target] = NewConfigEntry()
	return nil
}

// Delete removes the config entry referenced vis the target string
func (c *configStore) Delete(target string) error {
	c.m.Lock()
	defer c.m.Unlock()
	delete(c.config, target)
	return nil
}

// GetTargets retrieves a slice of available targets in the config map
func (c *configStore) GetTargets() []string {
	targets := []string{}
	c.m.Lock()
	defer c.m.Unlock()
	for target := range c.config {
		targets = append(targets, target)
	}
	return targets
}
