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
	"errors"

	"github.com/nats-io/nats.go"
)

const (
	defaultNATSAddr = "nats.ndd-system.svc.cluster.local"
	// natsAddrTmpl    = "%s.%s.svc.%s" // TODO: use to figure out nats address instead of using the default value
)

// createStream creates a stream if it does not exist using JetStreamContext
func createStream(js nats.JetStreamContext, str *nats.StreamConfig) error {
	// Check if the stream already exists; if not, create it.
	stream, err := js.StreamInfo(str.Name)
	if err != nil {
		// ignore Notfound error and continue
		if !errors.Is(err, nats.ErrStreamNotFound) {
			return err
		}
	}
	// stream does not exist, create it
	if stream == nil {
		_, err = js.AddStream(str)
		if err != nil {
			return err
		}
	}
	return nil
}
