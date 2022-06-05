package collector

import (
	"context"
	"errors"
	"sync"

	"github.com/karimra/gnmic/types"
	"github.com/yndd/cache/pkg/cache"
	"github.com/yndd/cache/pkg/origin"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/ndd-runtime/pkg/meta"
	"github.com/yndd/state/pkg/ygotnddpstate"
)

// Option can be used to manipulate Collector config.
type Option func(Collector)

type Collector interface {
	// add a logger to Collector
	WithLogger(log logging.Logger)
	// add a cache to Collector
	WithCache(c cache.Cache)
	// check if a target exists
	IsActive(target string) bool
	// start target collector
	ReconcileTarget(tc *types.TargetConfig) error
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

// WithLogger specifies how the collector logs messages.
func WithCache(c cache.Cache) Option {
	return func(d Collector) {
		d.WithCache(c)
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
	cache            cache.Cache
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

func (c *collector) WithCache(cache cache.Cache) {
	c.cache = cache
}

func (c *collector) IsActive(target string) bool {
	c.m.Lock()
	defer c.m.Unlock()
	_, ok := c.targetCollectors[target]
	return ok
}

func (c *collector) ReconcileTarget(tc *types.TargetConfig) error {
	cacheNsTargetName := meta.NamespacedName(tc.Name).GetPrefixNamespacedName(origin.State)
	log := c.log.WithValues("target", tc.Name, "cacheNsTargetName", cacheNsTargetName)
	log.Debug("Collector ReconcileTarget...")

	ce, err := c.cache.GetEntry(cacheNsTargetName)
	if err != nil {
		log.Debug("cache not ready", "error", err)
		return err
	}
	runningConfig, ok := ce.GetRunningConfig().(*ygotnddpstate.Device)
	if !ok {
		log.Debug("unexpected Object")
		return errors.New("unexpected Object")
	}
	log.Debug("Collector", "Running config", runningConfig)
	// validate if there are still state entries in the config, if not we should delete the target
	if len(runningConfig.StateEntry) == 0 {
		if c.IsActive(tc.Name) {
			err = c.StopTarget(tc.Name)
			if err != nil {
				return err
			}
		}
		log.Debug("handleUpdate No state config left")
		return nil
	}

	if c.IsActive(tc.Name) {
		log.Debug("handleSet", "Active", true)
		if err := c.StopTarget(tc.Name); err != nil {
			log.Debug("handleSet", "Stop success", false)
			return err
		}
		c.log.Debug("handleSet", "Stop success", true)
	}
	log.Debug("handleUpdate with running config", "runningConfig", runningConfig)

	// create a new target collector
	tColl, err := NewTargetCollector(c.ctx, tc, runningConfig,
		WithTargetCollectorLogger(c.log),
		WithTargetCollectorNATSAddr(c.natsAddr),
	)
	if err != nil {
		return err
	}

	c.m.Lock()
	defer c.m.Unlock()
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
