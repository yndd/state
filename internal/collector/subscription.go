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

	gapi "github.com/karimra/gnmic/api"
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/yndd/ndd-yang/pkg/yparser"
	"github.com/yndd/nddp-state/pkg/ygotnddpstate"
)

// Subscription defines the parameters for the subscription
type Subscription struct {
	Name        string
	StateConfig *ygotnddpstate.Device

	cfn context.CancelFunc
}

func (s *Subscription) GetName() string {
	return s.Name
}

func (s *Subscription) GetPaths() []*gnmi.Path {
	paths := []*gnmi.Path{}

	for _, m := range s.StateConfig.StateEntry {
		for _, p := range m.Path {
			paths = append(paths, yparser.Xpath2GnmiPath(p, 0))
		}
	}
	return paths
}

func (s *Subscription) SetName(n string) {
	s.Name = n
}

func (s *Subscription) SetCancelFn(c context.CancelFunc) {
	s.cfn = c
}

// createSubscribeRequest create a gnmi subscription
func (s *Subscription) createSubscribeRequest() (*gnmi.SubscribeRequest, error) {
	// create subscription
	gnmiOpts := []gapi.GNMIOption{
		gapi.SubscriptionListModeSTREAM(),
		gapi.EncodingASCII(),
	}
	for _, st := range s.StateConfig.StateEntry {
		for _, p := range st.Path {
			gnmiOpts = append(gnmiOpts,
				gapi.Subscription(
					gapi.Path(p),
					gapi.SubscriptionModeON_CHANGE(),
				),
			)
		}
	}
	return gapi.NewSubscribeRequest(gnmiOpts...)
}
