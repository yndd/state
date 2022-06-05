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
	"time"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/yndd/cache/pkg/validator"
	"github.com/yndd/ndd-runtime/pkg/meta"
	"github.com/yndd/ndd-yang/pkg/yparser"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// errors
	errTargetNotFoundInCache = "could not find target in cache"
)

func (s *subServer) Set(ctx context.Context, p *gnmi.Path, upd *gnmi.Update) (*gnmi.SetResponse, error) {
	cacheNsTargetName := meta.NamespacedName(p.GetTarget()).GetPrefixNamespacedName(p.GetOrigin())
	log := s.log.WithValues("origin", p.GetOrigin(), "target", p.GetTarget(), "cacheNsTargetName", cacheNsTargetName)

	log.Debug("Set Update/Replace...", "update", upd)

	// get cahe entry for target
	ce, err := s.cache.GetEntry(cacheNsTargetName)
	if err != nil {
		log.Debug("Set Update/Replace cache entry node found", "error", err)
		return nil, status.Errorf(codes.NotFound, errTargetNotFoundInCache)
	}
	// validates and updates the running config
	if err := validator.ValidateUpdate(ce, []*gnmi.Update{upd}, true, false, validator.Origin_GnmiServer); err != nil {
		log.Debug("Set Update/Replace  validate update failed", "error", err)
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	//return s.handleUpdate(p)
	ti := s.stateTargetController.GetTargetInstance(p.GetTarget())
	if ti == nil {
		// target does not exist -> we can return an error
		log.Debug("Set Update/Replace  target instance does not exist")
		return nil, status.Errorf(codes.NotFound, errTargetNotFoundInCache)
	}

	tc, err := ti.GetTargetConfig()
	if err != nil {
		// target does not exist -> we can return an error
		log.Debug("Set Update/Replace  target config not found")
		return nil, status.Errorf(codes.NotFound, errTargetNotFoundInCache)
	}

	if err := s.collector.ReconcileTarget(tc); err != nil {
		log.Debug("Set Update/Replace ReconcileTarget error", "error", err)
		return nil, status.Errorf(codes.Internal, "ReconcileTarget error: %v", err)
	}

	return &gnmi.SetResponse{
		Response: []*gnmi.UpdateResult{
			{
				Timestamp: time.Now().UnixNano(),
			},
		}}, nil
}

func (s *subServer) Delete(ctx context.Context, p *gnmi.Path, del *gnmi.Path) (*gnmi.SetResponse, error) {
	cacheNsTargetName := meta.NamespacedName(p.GetTarget()).GetPrefixNamespacedName(p.GetOrigin())
	log := s.log.WithValues("origin", p.GetOrigin(), "target", p.GetTarget(), "cacheNsTargetName", cacheNsTargetName)

	log.Debug("Set Delete...", "path", yparser.GnmiPath2XPath(del, true))
	// TODO check if origin != State -> return error
	//return s.handleUpdate(p)

	// get cache entry for the  target
	ce, err := s.cache.GetEntry(cacheNsTargetName)
	if err != nil {
		log.Debug("Set Update/Replace cache entry node found", "error", err)
		return nil, status.Errorf(codes.NotFound, errTargetNotFoundInCache)
	}

	// delete the entry from the cache/running config
	if err := validator.ValidateDelete(ce, []*gnmi.Path{del}, validator.Origin_GnmiServer); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	ti := s.stateTargetController.GetTargetInstance(p.GetTarget())
	if ti == nil {
		// target does not exist -> we can return an error
		return nil, status.Errorf(codes.InvalidArgument, errTargetNotFoundInCache)
	}

	tc, err := ti.GetTargetConfig()
	if err != nil {
		// target does not exist -> we can return an error
		return nil, status.Errorf(codes.InvalidArgument, errTargetNotFoundInCache)
	}

	if err := s.collector.ReconcileTarget(tc); err != nil {
		return nil, status.Errorf(codes.Internal, "start target error: %v", err)
	}

	return &gnmi.SetResponse{
		Response: []*gnmi.UpdateResult{
			{
				Timestamp: time.Now().UnixNano(),
			},
		}}, nil
}
