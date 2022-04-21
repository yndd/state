package collector

import (
	"sync"

	"github.com/karimra/gnmic/types"
	"github.com/openconfig/gnmi/cache"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/nddp-state/pkg/ygotnddpstate"
)

// Option can be used to manipulate Options.
type Option func(Collector)

type Collector interface {
	WithLogger(log logging.Logger)
	IsActive(target string) bool
	Start(t *types.TargetConfig, mc *ygotnddpstate.Device) error
	Stop(target string) error
}

// WithLogger specifies how the collector logs messages.
func WithLogger(log logging.Logger) Option {
	return func(d Collector) {
		d.WithLogger(log)
	}
}

type collector struct {
	m         sync.Mutex
	collector map[string]StateCollector
	cache     *cache.Cache

	log logging.Logger
}

func New(cache *cache.Cache, opts ...Option) Collector {
	c := &collector{
		collector: map[string]StateCollector{},
		cache:     cache,
	}
	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *collector) WithLogger(log logging.Logger) {
	c.log = log
}

func (c *collector) IsActive(target string) bool {
	c.m.Lock()
	defer c.m.Unlock()
	if _, ok := c.collector[target]; !ok {
		return false
	}
	return true
}

func (c *collector) Start(t *types.TargetConfig, mc *ygotnddpstate.Device) error {
	c.m.Lock()
	defer c.m.Unlock()

	nmc, err := NewStateCollector(t, mc,
		WithStateCollectorLogger(c.log),
		WithStateCollectorCache((*cache.Cache)(c.cache)))
	if err != nil {
		return err
	}

	c.collector[t.Name] = nmc
	if err := c.collector[t.Name].Start(); err != nil {
		return err
	}
	return nil
}

func (c *collector) Stop(target string) error {
	c.m.Lock()
	defer c.m.Unlock()

	mc, ok := c.collector[target]
	if !ok {
		return nil
	}

	if err := mc.Stop(); err != nil {
		return err
	}
	return nil
}
