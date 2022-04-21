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
	"strings"

	"github.com/nats-io/nats.go"
)

type stream struct {
	Name     string
	Subjects []string
}

// createStream creates a stream by using JetStreamContext
func createStream(js nats.JetStreamContext, str *stream) error {
	// Check if the stream already exists; if not, create it.
	stream, err := js.StreamInfo(str.Name)
	if err != nil {
		// ignore not found
		if !strings.Contains(err.Error(), "not found") {
			return err
		}
	}
	if stream == nil {
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     str.Name,
			Subjects: str.Subjects,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
