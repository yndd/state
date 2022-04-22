package collector

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/karimra/gnmic/types"
	"github.com/nats-io/nats.go"
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

// stateMsg is the msg send from a stateCollector to the collector through an update channel
type stateMsg struct {
	Subject   string `json:"-"`
	Timestamp int64  `json:"timestamp,omitempty"`
	Operation string `json:"operation,omitempty"`
	Data      []byte `json:"data,omitempty"`
}

type collector struct {
	m         sync.Mutex
	collector map[string]StateCollector
	// updateCh is the channel through which
	// the collector will receive stateMsgs from
	// the stateCollectors
	updateCh  chan *stateMsg
	natsAddr  string
	namespace string
	log       logging.Logger
}

func New(opts ...Option) Collector {
	c := &collector{
		collector: map[string]StateCollector{},
		updateCh:  make(chan *stateMsg),
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.natsAddr == "" {
		c.natsAddr = defaultNATSAddr
	}
	c.namespace, _ = os.LookupEnv("POD_NAMESPACE")
	c.log.Info("pod namespace", "ns", c.namespace)
	go c.natsProducerWorker(context.TODO())
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
		WithStateCollectorUpdateCh(c.updateCh),
		WithNamespace(c.namespace),
	)
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

func (c *collector) natsProducerWorker(ctx context.Context) {
	var nc *nats.Conn
	var jsc nats.JetStreamContext
	var err error
STARTCONN:
	nc, err = nats.Connect(c.natsAddr)
	if err != nil {
		time.Sleep(time.Second)
		goto STARTCONN
	}
	defer nc.Close()
	jsc, err = nc.JetStream()
	if err != nil {
		c.log.Info("inconsistent JetStream Options", "error", err)
	}
	err = createStream(jsc, &nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{streamSubjects},
	})
	if err != nil {
		nc.Close()
		time.Sleep(time.Second)
		goto STARTCONN
	}
	for {
		select {
		case <-ctx.Done():
			return
		case sm := <-c.updateCh:
			// TOOD: move json marshaling to the stateCollector
			// rather than in the worker
			c.log.Info("state msg", "msg", sm)
			b, err := json.Marshal(sm)
			if err != nil {
				c.log.Info("failed to marshal msg", "err", err)
				continue
			}
			c.log.Info("JetStream", "subject", sm.Subject)
			c.log.Info("JetStream", "data", string(b))
			_, err = jsc.Publish(sm.Subject, b)
			if err != nil {
				c.log.Info("JetStream Publish error", "error", err)
				nc.Close()
				goto STARTCONN
			}
		}
	}
}
