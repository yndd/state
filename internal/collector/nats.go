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
	"fmt"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/openconfig/gnmi/path"
	"github.com/openconfig/gnmi/proto/gnmi"
)

const (
	natsServer = "nats.ndd-system.svc.cluster.local"
	natsStream = "nddpstate"
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
		if !errors.Is(err, nats.ErrStreamNotFound) {
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

func publishGnmiUpdate(js nats.JetStreamContext, target string, u *gnmi.Update) error {
	b, err := getByteSlice(u.GetVal())
	if err != nil {
		return err
	}
	if _, err := js.Publish(getSubjectFromPath(target, u.GetPath()), b); err != nil {
		return err
	}
	return nil
}

func getSubjectFromPath(target string, p *gnmi.Path) string {
	pathStr := strings.Join(path.ToStrings(p, false), ".")
	return strings.Join([]string{natsStream, target, pathStr}, ".")
}

// GetByteSlice return a []byte of the gnmi typed value
func getByteSlice(updValue *gnmi.TypedValue) ([]byte, error) {
	if updValue == nil {
		return nil, nil
	}
	switch updValue.Value.(type) {
	case *gnmi.TypedValue_AsciiVal:
		return []byte(updValue.GetAsciiVal()), nil
	case *gnmi.TypedValue_BoolVal:
		return []byte(fmt.Sprintf("%v", updValue.GetBoolVal())), nil
	case *gnmi.TypedValue_BytesVal:
		return updValue.GetBytesVal(), nil
	case *gnmi.TypedValue_DecimalVal:
		return []byte(fmt.Sprintf("%v", updValue.GetDecimalVal())), nil
	case *gnmi.TypedValue_FloatVal:
		return []byte(fmt.Sprintf("%v", updValue.GetFloatVal())), nil
	case *gnmi.TypedValue_IntVal:
		return []byte(fmt.Sprintf("%v", updValue.GetIntVal())), nil
	case *gnmi.TypedValue_StringVal:
		return []byte(updValue.GetStringVal()), nil
	case *gnmi.TypedValue_UintVal:
		return []byte(fmt.Sprintf("%v", updValue.GetUintVal())), nil
	case *gnmi.TypedValue_JsonIetfVal:
		return updValue.GetJsonIetfVal(), nil
	case *gnmi.TypedValue_JsonVal:
		return updValue.GetJsonVal(), nil
	case *gnmi.TypedValue_LeaflistVal:
		//value = updValue.GetLeaflistVal()
	case *gnmi.TypedValue_ProtoBytes:
		//value = updValue.GetProtoBytes()
	case *gnmi.TypedValue_AnyVal:
		//value = updValue.GetAnyVal()
	}

	return nil, nil
}
