package collector

import (
	"context"
	"sync"

	"github.com/karimra/gnmic/types"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/nddp-state/pkg/ygotnddpstate"
)

// Option can be used to manipulate Collector config.
type Option func(Collector)

type Collector interface {
	// add a logger to Collector
	WithLogger(log logging.Logger)
	// check if a target exists
	IsActive(target string) bool
	// start target collector
	StartTarget(t *types.TargetConfig, mc *ygotnddpstate.Device) error
	// stop target collector
	StopTarget(target string) error
	// stop all target collectors
	Stop() error
}

// WithLogger specifies how the collector logs messages.
func WithLogger(log logging.Logger) Option {
	return func(d Collector) {
		d.WithLogger(log)
	}
}

// stateMsg is the msg send from a stateCollector to the collector through an update channel
type stateMsg struct {
	Subject   string `json:"-"`
	Timestamp int64  `json:"timestamp,omitempty"`
	Operation string `json:"operation,omitempty"`
	Data      []byte `json:"data,omitempty"`
}

// collector is the implementation of Collector interface
type collector struct {
	m sync.Mutex
	// state collectors indexed by target name
	targetCollectors map[string]TargetCollector
	ctx              context.Context
	cfn              context.CancelFunc
	natsAddr         string
	log              logging.Logger
}

// New creates a new Collector interface
func New(ctx context.Context, opts ...Option) Collector {
	c := &collector{
		targetCollectors: map[string]TargetCollector{},
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.natsAddr == "" {
		c.natsAddr = defaultNATSAddr
	}
	c.ctx, c.cfn = context.WithCancel(ctx)
	return c
}

func (c *collector) WithLogger(log logging.Logger) {
	c.log = log
}

func (c *collector) IsActive(target string) bool {
	c.m.Lock()
	defer c.m.Unlock()
	_, ok := c.targetCollectors[target]
	return ok
}

func (c *collector) StartTarget(tc *types.TargetConfig, mc *ygotnddpstate.Device) error {
	c.m.Lock()
	defer c.m.Unlock()
	// create a new target collector
	tColl, err := NewTargetCollector(c.ctx, tc, mc,
		WithTargetCollectorLogger(c.log),
		WithTargetCollectorNATSAddr(c.natsAddr),
	)
	if err != nil {
		return err
	}

	c.targetCollectors[tc.Name] = tColl
	if err := tColl.Start(c.ctx); err != nil {
		return err
	}
	return nil
}

func (c *collector) StopTarget(target string) error {
	c.m.Lock()
	defer c.m.Unlock()

	tColl, ok := c.targetCollectors[target]
	if !ok {
		return nil
	}
	// delete state collector
	delete(c.targetCollectors, target)
	return tColl.Stop()
}

func (c *collector) Stop() error {
	c.m.Lock()
	defer c.m.Unlock()

	for target, tColl := range c.targetCollectors {
		delete(c.targetCollectors, target)
		err := tColl.Stop()
		if err != nil {
			c.log.Debug("failed to stop target collector", "target", target, "error", err)
		}
	}
	return nil
}
