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
	"github.com/yndd/cache/pkg/encoder"
	"github.com/yndd/ndd-runtime/pkg/meta"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *subServer) Get(ctx context.Context, req *gnmi.GetRequest) (*gnmi.GetResponse, error) {
	prefix := req.GetPrefix()
	cacheNsTargetName := meta.NamespacedName(prefix.GetTarget()).GetPrefixNamespacedName(prefix.GetOrigin())
	log := s.log.WithValues("origin", prefix.GetOrigin(), "target", prefix.GetTarget(), "cacheNsTargetName", cacheNsTargetName)

	log.Debug("Get...", "path", req.GetPath())

	ce, err := s.cache.GetEntry(cacheNsTargetName)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "cache not ready")
	}

	goStruct := ce.GetRunningConfig() // no DeepCopy required, since we get a deepcopy already
	model := ce.GetModel()
	ts := time.Now().UnixNano()

	ns, err := encoder.PopulateNotification(goStruct, req, model, ts, prefix)
	if err != nil {
		return nil, err
	}
	return &gnmi.GetResponse{
		Notification: ns,
	}, err
}
