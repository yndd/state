package config

import (
	"reflect"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
	"github.com/yndd/ndd-runtime/pkg/model"
	"github.com/yndd/state/pkg/ygotnddpstate"
)

type ConfigEntry interface {
	GetRunningConfig() ygot.ValidatedGoStruct
	SetRunningConfig(ygot.ValidatedGoStruct)
	GetModel() *model.Model
}

type configEntry struct {
	model         *model.Model
	runningConfig ygot.ValidatedGoStruct
}

func NewConfigEntry() ConfigEntry {
	return &configEntry{
		model: &model.Model{
			ModelData:       make([]*gnmi.ModelData, 0),
			StructRootType:  reflect.TypeOf((*ygotnddpstate.Device)(nil)),
			SchemaTreeRoot:  ygotnddpstate.SchemaTree["Device"],
			JsonUnmarshaler: ygotnddpstate.Unmarshal,
			EnumData:        ygotnddpstate.ΛEnum,
		},
	}
}

func (ce *configEntry) GetRunningConfig() ygot.ValidatedGoStruct {
	return ce.runningConfig
}

func (ce *configEntry) SetRunningConfig(c ygot.ValidatedGoStruct) {
	ce.runningConfig = c
}

func (ce *configEntry) GetModel() *model.Model {
	return ce.model
}
