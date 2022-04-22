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
	"time"

	"github.com/karimra/gnmic/target"
	"github.com/karimra/gnmic/types"
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/pkg/errors"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/nddp-state/pkg/ygotnddpstate"
	"google.golang.org/grpc"
)

const (
	// timers
	defaultTimeout             = 5 * time.Second
	defaultTargetReceivebuffer = 1000
	defaultLockRetry           = 5 * time.Second
	defaultRetryTimer          = 10 * time.Second

	// errors
	errCreateGnmiClient          = "cannot create gnmi client"
	errCreateSubscriptionRequest = "cannot create subscription request"
)

// StateCollector defines the interfaces for the collector
type StateCollector interface {
	Start() error
	Stop() error
}

// Option can be used to manipulate Options.
type StateCollectorOption func(*stateCollector)

// WithLogger specifies how the collector logs messages.
func WithStateCollectorLogger(log logging.Logger) StateCollectorOption {
	return func(o *stateCollector) {
		o.log = log
	}
}

// stateCollector defines the parameters for the collector
type stateCollector struct {
	target              *target.Target
	subscriptions       []*Subscription
	ctx                 context.Context
	targetReceiveBuffer uint
	retryTimer          time.Duration

	stopCh chan bool

	//mutex sync.RWMutex
	log logging.Logger
}

// NewCollector creates a new GNMI collector
func NewStateCollector(t *types.TargetConfig, mc *ygotnddpstate.Device, opts ...StateCollectorOption) (StateCollector, error) {
	c := &stateCollector{
		subscriptions: []*Subscription{
			{
				Name:        "state-collector",
				StateConfig: mc,
				stopCh:      make(chan bool),
			},
		},
		stopCh: make(chan bool),
		//mutex:               sync.RWMutex{},
		targetReceiveBuffer: defaultTargetReceivebuffer,
		retryTimer:          defaultRetryTimer,
		ctx:                 context.Background(),
	}
	for _, opt := range opts {
		opt(c)
	}

	c.target = target.NewTarget(t)
	if err := c.target.CreateGNMIClient(c.ctx, grpc.WithBlock()); err != nil { // TODO add dialopts
		return nil, errors.Wrap(err, errCreateGnmiClient)
	}

	return c, nil
}

// Lock locks a gnmi collector
func (c *stateCollector) GetTarget() *target.Target {
	return c.target
}

// GetSubscription returns a bool based on a subscription name
func (c *stateCollector) GetSubscriptions() []*Subscription {
	return c.subscriptions
}

// GetSubscription returns a bool based on a subscription name
func (c *stateCollector) GetSubscription(subName string) *Subscription {
	for _, s := range c.GetSubscriptions() {
		if s.GetName() == subName {
			return s
		}
	}
	return nil
}

// StartGnmiSubscriptionHandler starts gnmi subscription
func (c *stateCollector) Start() error {
	log := c.log.WithValues("Target", c.target.Config.Name, "Address", c.target.Config.Address)
	log.Debug("Starting statecollector...")

	errChannel := make(chan error)
	go func() {
		if err := c.run(); err != nil {
			errChannel <- errors.Wrap(err, "error starting statecollector ")
		}
		errChannel <- nil
	}()
	return nil
}

// run state collector
func (c *stateCollector) run() error {
	log := c.log.WithValues("Target", c.target.Config.Name, "Address", c.target.Config.Address)
	log.Debug("Running statecollector...")

	c.ctx, c.subscriptions[0].cancelFn = context.WithCancel(c.ctx)

	//c.Lock()
	// this subscription is a go routine that runs until you send a stop through the stopCh
	go c.startSubscription(c.ctx, &gnmi.Path{}, c.GetSubscriptions())
	//c.Unlock()

	chanSubResp, chanSubErr := c.GetTarget().ReadSubscriptions()

	// run the response handler
	for {
		select {
		case resp := <-chanSubResp:
			c.handleSubscription(resp.Response)
		case tErr := <-chanSubErr:
			c.log.Debug("subscribe", "error", tErr)
			return errors.New("handle subscription error")
		case <-c.stopCh:
			c.log.Debug("Stopping collecor process...")
			return nil
		}
	}
}

// StartSubscription starts a subscription
func (c *stateCollector) startSubscription(ctx context.Context, prefix *gnmi.Path, s []*Subscription) error {
	log := c.log.WithValues("subscription", s[0].GetName(), "Paths", s[0].GetPaths())
	log.Debug("subscription start...")
	// initialize new subscription

	req, err := CreateSubscriptionRequest(prefix, s[0])
	if err != nil {
		c.log.Debug(errCreateSubscriptionRequest, "error", err)
		return errors.Wrap(err, errCreateSubscriptionRequest)
	}

	log.Debug("Subscription", "Request", req)
	go func() {
		c.target.Subscribe(ctx, req, s[0].GetName())
	}()
	log.Debug("subscription started ...")

	for {
		select {
		case <-s[0].stopCh: // execute quit
			c.log.Debug("subscription cancelled")
			s[0].cancelFn()
			//c.mutex.Lock()
			// TODO delete subscription from list
			//delete(c.subscriptions, subName)
			//c.mutex.Unlock()

			return nil
		}
	}
}

// StartGnmiSubscriptionHandler starts gnmi subscription
func (c *stateCollector) Stop() error {
	log := c.log.WithValues("Target", c.GetTarget().Config.Name)
	log.Debug("Stop Collector...")

	c.stopSubscription(c.ctx, c.GetSubscriptions()[0])
	c.stopCh <- true

	return nil
}

// StopSubscription stops a subscription
func (c *stateCollector) stopSubscription(ctx context.Context, s *Subscription) error {
	c.log.Debug("subscription stop...")
	//s.stopCh <- true // trigger quit
	s.cancelFn()
	c.log.Debug("subscription stopped")
	return nil
}
