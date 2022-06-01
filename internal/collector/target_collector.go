/*
Copyright 2021 NDD.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package collector

import (
	"context"
	"encoding/json"
	"time"

	"github.com/karimra/gnmic/target"
	"github.com/karimra/gnmic/types"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/state/pkg/ygotnddpstate"
	"google.golang.org/grpc"
)

const (
	// timers
	defaultTimeout             = 5 * time.Second
	defaultTargetReceiveBuffer = 1000
	defaultLockRetry           = 5 * time.Second
	defaultRetryTimer          = 10 * time.Second
	// nats
	streamName     = "nddpstate"
	streamSubjects = "nddpstate.>"

	// errors
	errCreateGnmiClient          = "cannot create gnmi client"
	errCreateSubscriptionRequest = "cannot create subscription request"
)

// TargetCollector defines the interfaces for the collector
type TargetCollector interface {
	Start(ctx context.Context) error
	Stop() error
}

// Option can be used to manipulate TargetCollector.
type TargetCollectorOption func(*targetCollector)

// WithLogger specifies how the collector logs messages.
func WithTargetCollectorLogger(log logging.Logger) TargetCollectorOption {
	return func(o *targetCollector) {
		o.log = log
	}
}

func WithTargetCollectorNATSAddr(addr string) TargetCollectorOption {
	return func(o *targetCollector) {
		o.natsAddr = addr
	}
}

// targetCollector defines the parameters for the collector
type targetCollector struct {
	// target the state is collected from
	target *target.Target
	// update channel nats producer goroutines read from
	updateCh chan *stateMsg
	// comma separated nats server addresses
	natsAddr string
	// subscriptions derived from State CR
	subscriptions []*Subscription
	// channel to signal stopping of the state collector
	stopCh chan struct{}
	// logger
	log logging.Logger
}

// NewTargetCollector creates a new GNMI collector for a target defined by target config tc,
// this function creates the gNMI client as well.
func NewTargetCollector(ctx context.Context, tc *types.TargetConfig, mc *ygotnddpstate.Device, opts ...TargetCollectorOption) (TargetCollector, error) {
	sc := &targetCollector{
		// TODO: add second subscription
		subscriptions: []*Subscription{
			{
				Name:        "target-collector",
				StateConfig: mc,
			},
		},
		updateCh: make(chan *stateMsg),
		stopCh:   make(chan struct{}),
	}
	for _, opt := range opts {
		opt(sc)
	}
	if tc.BufferSize == 0 {
		tc.BufferSize = defaultTargetReceiveBuffer
	}
	if tc.RetryTimer <= 0 {
		tc.RetryTimer = defaultRetryTimer
	}
	sc.target = target.NewTarget(tc)
	if err := sc.target.CreateGNMIClient(ctx, grpc.WithBlock()); err != nil { // TODO add dialopts
		return nil, errors.Wrap(err, errCreateGnmiClient)
	}

	return sc, nil
}

// Lock locks a gnmi collector
func (c *targetCollector) GetTarget() *target.Target {
	return c.target
}

// GetSubscription returns a bool based on a subscription name
func (c *targetCollector) GetSubscriptions() []*Subscription {
	return c.subscriptions
}

// GetSubscription returns a bool based on a subscription name
func (c *targetCollector) GetSubscription(subName string) *Subscription {
	for _, s := range c.GetSubscriptions() {
		if s.GetName() == subName {
			return s
		}
	}
	return nil
}

// Start starts the target collector, i.e the nats producer and the gnmi subscription
func (c *targetCollector) Start(ctx context.Context) error {
	log := c.log.WithValues("Target", c.target.Config.Name, "Address", c.target.Config.Address)
	log.Debug("Starting target collector", "target", c.target.Config.Name)

	go c.natsProducerWorker(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := c.run(ctx)
				switch err {
				case nil:
					// stop channel closed
					return
				case context.Canceled:
					// target collector stopped
					return
				default:
					log.Debug("error starting target collector", "error", err)
					time.Sleep(time.Second)
					continue
				}
			}
		}
	}()
	return nil
}

// run state collector
func (c *targetCollector) run(ctx context.Context) error {
	log := c.log.WithValues("Target", c.target.Config.Name, "Address", c.target.Config.Address)
	log.Debug("Running target collector...")

	ctx, c.subscriptions[0].cfn = context.WithCancel(ctx)

	// this subscription is a go routine that runs until the cancel function is called
	go c.startSubscription(ctx)

	chanSubResp, chanSubErr := c.GetTarget().ReadSubscriptions()
	// run the response handler loop
	for {
		select {
		// subscribe response or error cases
		case resp := <-chanSubResp:
			c.handleSubscribeResponse(resp.Response)
		case tErr := <-chanSubErr:
			c.log.Debug("subscribe", "error", tErr)
			return errors.New("handle subscription error")

		// stop cases
		// canceled when the subscription is stopped
		case <-ctx.Done(): // canceled when the subscription is stopped
			c.log.Debug("subscription stopped", "error", ctx.Err())
			return ctx.Err()
		// the whole target collector is stopped
		case <-c.stopCh: // the whole target collector is stopped
			c.stopSubscription(c.subscriptions[0])
			c.log.Debug("Stopping target collector process...")
			return nil
		}
	}
}

// StartSubscription starts a subscription
func (c *targetCollector) startSubscription(ctx context.Context) error {
	s := c.GetSubscriptions()
	log := c.log.WithValues("subscription", s[0].GetName(), "Paths", s[0].GetPaths())
	log.Debug("subscription starting", "target", c.target.Config.Name)
	// create subscription request
	req, err := s[0].createSubscribeRequest()
	if err != nil {
		c.log.Debug(errCreateSubscriptionRequest, "error", err)
		return errors.Wrap(err, errCreateSubscriptionRequest)
	}

	log.Debug("Subscription", "Request", req)
	go c.target.Subscribe(ctx, req, s[0].GetName())
	log.Debug("subscription started", "target", c.target.Config.Name)
	return nil
}

// Stop stops the target collector
func (c *targetCollector) Stop() error {
	log := c.log.WithValues("Target", c.GetTarget().Config.Name)
	log.Debug("Stoping target collector...", "target", c.target.Config.Name)

	c.stopSubscription(c.GetSubscriptions()[0])
	close(c.stopCh)

	return nil
}

// StopSubscription stops a subscription
func (c *targetCollector) stopSubscription(s *Subscription) error {
	c.log.Debug("subscription stop...")
	s.cfn()
	c.log.Debug("subscription stopped")
	return nil
}

func (c *targetCollector) natsProducerWorker(ctx context.Context) {
	var nc *nats.Conn
	var jsc nats.JetStreamContext
	var err error
STARTCONN:
	select {
	case <-ctx.Done():
		c.log.Debug("nats producer stopped", "error", ctx.Err())
		return
	default:
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
				c.log.Debug("nats producer stopped", "error", ctx.Err())
				return
			case sm := <-c.updateCh:
				c.log.Debug("state msg", "msg", sm)
				b, err := json.Marshal(sm)
				if err != nil {
					c.log.Info("failed to marshal msg", "error", err)
					continue
				}
				c.log.Debug("JetStream", "subject", sm.Subject, "data", string(b))
				_, err = jsc.Publish(sm.Subject, b)
				if err != nil {
					c.log.Info("JetStream Publish error", "error", err)
					nc.Close()
					goto STARTCONN
				}
			}
		}
	}
}
