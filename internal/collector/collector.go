package collector

import (
	"strings"
	"sync"

	"github.com/karimra/gnmic/types"
	"github.com/nats-io/nats.go"
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

func New(opts ...Option) (Collector, error) {
	c := &collector{
		collector: map[string]StateCollector{},
	}
	for _, opt := range opts {
		opt(c)
	}

	nc, err := nats.Connect(natsServer)
	if err != nil {
		return nil, err
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}

	if err := createStream(js, &stream{
		Name:     natsStream,
		Subjects: []string{strings.Join([]string{natsStream, ">"}, ".")},
	}); err != nil {
		c.log.Debug("create stream", "error", err.Error())
		return nil, err
	}

	return c, nil
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
		WithStateCollectorLogger(c.log))
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
