package config

import (
	"fmt"
	"sync"
)

type Config interface {
	//Lock()
	//Unlock()
	HasTarget(target string) bool
	Get(target string) (ConfigEntry, error)
	Add(target string) error
	Delete(target string) error
	GetTargets() []string
}

type config struct {
	m sync.RWMutex

	config map[string]ConfigEntry
}

func New() Config {
	return &config{
		config: map[string]ConfigEntry{},
	}
}

func (c *config) HasTarget(target string) bool {
	c.m.Lock()
	defer c.m.Unlock()
	_, ok := c.config[target]
	return ok
}

func (c *config) Get(target string) (ConfigEntry, error) {
	c.m.RLock()
	defer c.m.RUnlock()
	tc, ok := c.config[target]
	if !ok {
		return nil, fmt.Errorf("config target '%s' does not exist", target)
	}
	return tc, nil
}

func (c *config) Add(target string) error {
	c.m.Lock()
	defer c.m.Unlock()
	c.config[target] = NewConfigEntry()
	return nil
}

func (c *config) Delete(target string) error {
	c.m.Lock()
	defer c.m.Unlock()
	delete(c.config, target)
	return nil
}

func (c *config) GetTargets() []string {
	targets := []string{}
	c.m.Lock()
	defer c.m.Unlock()
	for target := range c.config {
		targets = append(targets, target)
	}
	return targets
}
