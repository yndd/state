/*
Copyright 2021 NDDO.

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

package stategnmihandler

import (
	"context"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/yndd/cache/pkg/cache"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/state/internal/collector"
	"github.com/yndd/state/internal/statetargetcontroller"
)

type Options struct {
	Logger                logging.Logger
	Cache                 cache.Cache
	Collector             collector.Collector
	StateTargetController statetargetcontroller.StateTargetController
}

type SubServer interface {
	Get(ctx context.Context, req *gnmi.GetRequest) (*gnmi.GetResponse, error)
	Set(ctx context.Context, p *gnmi.Path, upd *gnmi.Update) (*gnmi.SetResponse, error)
	Delete(ctx context.Context, p *gnmi.Path, del *gnmi.Path) (*gnmi.SetResponse, error)
}

func New(o *Options) SubServer {
	s := &subServer{
		log:                   o.Logger,
		cache:                 o.Cache,
		collector:             o.Collector,
		stateTargetController: o.StateTargetController,
	}
	return s
}

type subServer struct {
	log                   logging.Logger
	cache                 cache.Cache
	collector             collector.Collector
	stateTargetController statetargetcontroller.StateTargetController
}
