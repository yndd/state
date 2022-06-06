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

	"github.com/openconfig/ygot/ygot"
	"github.com/yndd/cache/pkg/cache"
	"github.com/yndd/cache/pkg/model"
	"github.com/yndd/cache/pkg/origin"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/ndd-runtime/pkg/meta"
	"github.com/yndd/registrator/registrator"
	"github.com/yndd/state/internal/collector"
	"github.com/yndd/state/pkg/ygotnddpstate"
	targetv1 "github.com/yndd/target/apis/target/v1"
	"github.com/yndd/target/pkg/targetinstance"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	Client      client.Client
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
	client client.Client // used to get the target credentials

	ctx context.Context
	log logging.Logger
}

func New(ctx context.Context, o *Options, opts ...Option) StateTargetController {
	log := o.Logger
	log.Debug("new target state controller")

	c := &stateTargetController{
		log:     o.Logger,
		client:  o.Client,
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
func (c *stateTargetController) GetTargetInstance(nsTargetName string) targetinstance.TargetInstance {
	c.m.Lock()
	defer c.m.Unlock()
	t, ok := c.targets[nsTargetName]
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

	if c.GetTargetInstance(nsTargetName) != nil {
		log.Debug("start target, nothing to do as target was already active ...")
		return
	}
	// initialize the target since no data/context of the target exists

	// the target we get on the channel has <namespace.target> semantics
	targetName := meta.NamespacedName(nsTargetName).GetName()
	namespace := meta.NamespacedName(nsTargetName).GetNameSpace()

	ti := targetinstance.NewTargetInstance(c.ctx, &targetinstance.TiOptions{
		Logger:       c.log,
		Namespace:    namespace,
		NsTargetName: nsTargetName,
		TargetName:   targetName,
		Cache:        c.options.Cache,
		Client:       c.client, // used to get target credentials
		Registrator:  c.options.Registrator,
	})
	c.addTargetInstance(nsTargetName, ti)

	// initialize the state target cache
	stateCacheNsTargetName := meta.NamespacedName(nsTargetName).GetPrefixNamespacedName(origin.State)
	log.Debug("start target; init target cache...", "stateCacheNsTargetName", stateCacheNsTargetName)
	ce := cache.NewCacheEntry(stateCacheNsTargetName)
	ce.SetModel(c.options.TargetModel)
	if err := initRunningConfig(ce); err != nil {
		log.Debug("start target: init running config failed", "error", err)
	}
	c.options.Cache.AddEntry(ce)

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

	stateCacheNsTargetName := meta.NamespacedName(nsTargetName).GetPrefixNamespacedName(origin.State)
	if err := c.options.Cache.DeleteEntry(stateCacheNsTargetName); err != nil {
		log.Debug("delete target from cache", "error", err)
	}

	c.deleteTargetInstance(nsTargetName)
	log.Debug("deleted target...")
}

func initRunningConfig(ce cache.CacheEntry) error {
	// initialize the go struct
	d := &ygotnddpstate.Device{
		//StateEntry: map[string]*ygotnddpstate.YnddState_StateEntry{},
	}
	j, err := ygot.EmitJSON(d, &ygot.EmitJSONConfig{})
	if err != nil {
		return err
	}
	// get the model
	m := ce.GetModel()
	goStruct, err := m.NewConfigStruct([]byte(j), true)
	if err != nil {
		return err
	}
	// set running config to an empty object
	ce.SetRunningConfig(goStruct)
	return nil
}
