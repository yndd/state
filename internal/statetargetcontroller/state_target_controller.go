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

package statetargetcontroller

import (
	"context"
	"sync"

	"github.com/yndd/cache/pkg/cache"
	"github.com/yndd/cache/pkg/model"
	"github.com/yndd/cache/pkg/origin"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/ndd-runtime/pkg/meta"
	"github.com/yndd/ndd-runtime/pkg/resource"
	"github.com/yndd/registrator/registrator"
	"github.com/yndd/state/internal/collector"
	targetv1 "github.com/yndd/target/apis/target/v1"
	"github.com/yndd/target/pkg/targetinstance"
	"k8s.io/client-go/rest"
)

// StateTargetController defines the interfaces for the target state controller
type StateTargetController interface {
	// start the target instance
	StartTarget(nsTargetName string)
	// stops the target instance
	StopTarget(nsTargetName string)
	// get a target instance from the target state controller
	GetTargetInstance(targetName string) targetinstance.TargetInstance
}

type Options struct {
	Logger      logging.Logger
	Registrator registrator.Registrator
	Collector   collector.Collector
	TargetModel *model.Model
	VendorType  targetv1.VendorType
	Cache       cache.Cache
}

// Option can be used to manipulate Collector config.
type Option func(StateTargetController)

// stateTargetController implements the StateTargetController interface
type stateTargetController struct {
	options *Options
	m       sync.RWMutex
	targets map[string]targetinstance.TargetInstance

	// kubernetes
	client resource.ClientApplicator // used to get the target credentials

	ctx context.Context
	log logging.Logger
}

func New(ctx context.Context, config *rest.Config, o *Options, opts ...Option) StateTargetController {
	log := o.Logger
	log.Debug("new target state controller")

	c := &stateTargetController{
		log:     o.Logger,
		options: o, // contains all options
		m:       sync.RWMutex{},
		targets: make(map[string]targetinstance.TargetInstance),
		ctx:     ctx,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// get a target instance from the target configuration controller
func (c *stateTargetController) GetTargetInstance(targetName string) targetinstance.TargetInstance {
	c.m.Lock()
	defer c.m.Unlock()
	t, ok := c.targets[targetName]
	if !ok {
		return nil
	}
	return t
}

// add a target instance to the target configuration controller
func (c *stateTargetController) addTargetInstance(nsTargetName string, t targetinstance.TargetInstance) {
	c.m.Lock()
	defer c.m.Unlock()
	c.targets[nsTargetName] = t
}

// delete a target instance from the target configuration controller
func (c *stateTargetController) deleteTargetInstance(nsTargetName string) error {
	c.m.Lock()
	defer c.m.Unlock()
	if ti, ok := c.targets[nsTargetName]; ok {
		if err := ti.StopTargetCollector(); err != nil {
			return err
		}
		if err := ti.StopTargetReconciler(); err != nil {
			return err
		}
		ti.DeRegister()
	}
	delete(c.targets, nsTargetName)
	return nil
}

func (c *stateTargetController) StartTarget(nsTargetName string) {
	log := c.log.WithValues("nsTargetName", nsTargetName)
	log.Debug("start target...")
	// the target we get on the channel has <namespace.target> semantics
	targetName := meta.NamespacedName(nsTargetName).GetName()
	namespace := meta.NamespacedName(nsTargetName).GetNameSpace()

	ti := targetinstance.NewTargetInstance(c.ctx, &targetinstance.TiOptions{
		Logger:       c.log,
		Namespace:    namespace,
		NsTargetName: nsTargetName,
		TargetName:   targetName,
		Cache:        c.options.Cache,
		Client:       c.client,
		Registrator:  c.options.Registrator,
	})
	c.addTargetInstance(nsTargetName, ti)

	// initialize the state target cache
	stateCacheNsTargetName := meta.NamespacedName(nsTargetName).GetPrefixNamespacedName(origin.State)
	cce := cache.NewCacheEntry(stateCacheNsTargetName)
	cce.SetModel(c.options.TargetModel)
	c.options.Cache.AddEntry(cce)

	// start the target in the collector if there is a running config
	tc, err := ti.GetTargetConfig()
	if err != nil {
		c.log.Debug("cannot get target config", "error", err)
	}
	if err := c.options.Collector.ReconcileTarget(tc); err != nil {
		c.log.Debug("start target state", "error", err)
	}
	// register service discovery
	ti.Register()
	log.Debug("started target...")
}

func (c *stateTargetController) StopTarget(nsTargetName string) {
	log := c.log.WithValues("nsTargetName", nsTargetName)
	log.Debug("delete target...")
	// delete the target instance -> stops the collectors, reconciler
	c.options.Collector.StopTarget(nsTargetName)

	c.deleteTargetInstance(nsTargetName)
	log.Debug("deleted target...")
}
