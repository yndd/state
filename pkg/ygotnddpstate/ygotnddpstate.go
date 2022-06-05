/*
Package ygotnddpstate is a generated package which contains definitions
of structs which represent a YANG schema. The generated schema can be
compressed by a series of transformations (compression was false
in this case).

This package was generated by /Users/henderiw/Documents/codeprojects/tmp/ygot/genutil/names.go
using the following YANG input files:
	- /Users/henderiw/Documents/codeprojects/yndd/yang-models//combined/yndd-state.yang
Imported modules were sourced from:
	- /Users/henderiw/Documents/codeprojects/yndd/yang-models/combined/...
*/
package ygotnddpstate

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/openconfig/goyang/pkg/yang"
	"github.com/openconfig/ygot/ygot"
	"github.com/openconfig/ygot/ytypes"
)

// Binary is a type that is used for fields that have a YANG type of
// binary. It is used such that binary fields can be distinguished from
// leaf-lists of uint8s (which are mapped to []uint8, equivalent to
// []byte in reflection).
type Binary []byte

// YANGEmpty is a type that is used for fields that have a YANG type of
// empty. It is used such that empty fields can be distinguished from boolean fields
// in the generated code.
type YANGEmpty bool

// UnionInt8 is an int8 type assignable to unions of which it is a subtype.
type UnionInt8 int8

// UnionInt16 is an int16 type assignable to unions of which it is a subtype.
type UnionInt16 int16

// UnionInt32 is an int32 type assignable to unions of which it is a subtype.
type UnionInt32 int32

// UnionInt64 is an int64 type assignable to unions of which it is a subtype.
type UnionInt64 int64

// UnionUint8 is a uint8 type assignable to unions of which it is a subtype.
type UnionUint8 uint8

// UnionUint16 is a uint16 type assignable to unions of which it is a subtype.
type UnionUint16 uint16

// UnionUint32 is a uint32 type assignable to unions of which it is a subtype.
type UnionUint32 uint32

// UnionUint64 is a uint64 type assignable to unions of which it is a subtype.
type UnionUint64 uint64

// UnionFloat64 is a float64 type assignable to unions of which it is a subtype.
type UnionFloat64 float64

// UnionString is a string type assignable to unions of which it is a subtype.
type UnionString string

// UnionBool is a bool type assignable to unions of which it is a subtype.
type UnionBool bool

// UnionUnsupported is an interface{} wrapper type for unsupported types. It is
// assignable to unions of which it is a subtype.
type UnionUnsupported struct {
	Value interface{}
}

var (
	SchemaTree map[string]*yang.Entry
	ΛEnumTypes map[string][]reflect.Type
)

func init() {
	var err error
	initΛEnumTypes()
	if SchemaTree, err = UnzipSchema(); err != nil {
		panic("schema error: " + err.Error())
	}
}

// Schema returns the details of the generated schema.
func Schema() (*ytypes.Schema, error) {
	uzp, err := UnzipSchema()
	if err != nil {
		return nil, fmt.Errorf("cannot unzip schema, %v", err)
	}

	return &ytypes.Schema{
		Root:       &Device{},
		SchemaTree: uzp,
		Unmarshal:  Unmarshal,
	}, nil
}

// UnzipSchema unzips the zipped schema and returns a map of yang.Entry nodes,
// keyed by the name of the struct that the yang.Entry describes the schema for.
func UnzipSchema() (map[string]*yang.Entry, error) {
	var schemaTree map[string]*yang.Entry
	var err error
	if schemaTree, err = ygot.GzipToSchema(ySchema); err != nil {
		return nil, fmt.Errorf("could not unzip the schema; %v", err)
	}
	return schemaTree, nil
}

// Unmarshal unmarshals data, which must be RFC7951 JSON format, into
// destStruct, which must be non-nil and the correct GoStruct type. It returns
// an error if the destStruct is not found in the schema or the data cannot be
// unmarshaled. The supplied options (opts) are used to control the behaviour
// of the unmarshal function - for example, determining whether errors are
// thrown for unknown fields in the input JSON.
func Unmarshal(data []byte, destStruct ygot.GoStruct, opts ...ytypes.UnmarshalOpt) error {
	tn := reflect.TypeOf(destStruct).Elem().Name()
	schema, ok := SchemaTree[tn]
	if !ok {
		return fmt.Errorf("could not find schema for type %s", tn)
	}
	var jsonTree interface{}
	if err := json.Unmarshal([]byte(data), &jsonTree); err != nil {
		return err
	}
	return ytypes.Unmarshal(schema, destStruct, jsonTree, opts...)
}

// Device represents the /device YANG schema element.
type Device struct {
	StateEntry map[string]*YnddState_StateEntry `path:"stateEntry" module:"yndd-state"`
}

// IsYANGGoStruct ensures that Device implements the yang.GoStruct
// interface. This allows functions that need to handle this struct to
// identify it as being generated by ygen.
func (*Device) IsYANGGoStruct() {}

// NewStateEntry creates a new entry in the StateEntry list of the
// Device struct. The keys of the list are populated from the input
// arguments.
func (t *Device) NewStateEntry(Name string) (*YnddState_StateEntry, error) {

	// Initialise the list within the receiver struct if it has not already been
	// created.
	if t.StateEntry == nil {
		t.StateEntry = make(map[string]*YnddState_StateEntry)
	}

	key := Name

	// Ensure that this key has not already been used in the
	// list. Keyed YANG lists do not allow duplicate keys to
	// be created.
	if _, ok := t.StateEntry[key]; ok {
		return nil, fmt.Errorf("duplicate key %v for list StateEntry", key)
	}

	t.StateEntry[key] = &YnddState_StateEntry{
		Name: &Name,
	}

	return t.StateEntry[key], nil
}

// GetOrCreateStateEntry retrieves the value with the specified keys from
// the receiver Device. If the entry does not exist, then it is created.
// It returns the existing or new list member.
func (t *Device) GetOrCreateStateEntry(Name string) *YnddState_StateEntry {

	key := Name

	if v, ok := t.StateEntry[key]; ok {
		return v
	}
	// Panic if we receive an error, since we should have retrieved an existing
	// list member. This allows chaining of GetOrCreate methods.
	v, err := t.NewStateEntry(Name)
	if err != nil {
		panic(fmt.Sprintf("GetOrCreateStateEntry got unexpected error: %v", err))
	}
	return v
}

// GetStateEntry retrieves the value with the specified key from
// the StateEntry map field of Device. If the receiver is nil, or
// the specified key is not present in the list, nil is returned such that Get*
// methods may be safely chained.
func (t *Device) GetStateEntry(Name string) *YnddState_StateEntry {

	if t == nil {
		return nil
	}

	key := Name

	if lm, ok := t.StateEntry[key]; ok {
		return lm
	}
	return nil
}

// DeleteStateEntry deletes the value with the specified keys from
// the receiver Device. If there is no such element, the function
// is a no-op.
func (t *Device) DeleteStateEntry(Name string) {
	key := Name

	delete(t.StateEntry, key)
}

// AppendStateEntry appends the supplied YnddState_StateEntry struct to the
// list StateEntry of Device. If the key value(s) specified in
// the supplied YnddState_StateEntry already exist in the list, an error is
// returned.
func (t *Device) AppendStateEntry(v *YnddState_StateEntry) error {
	if v.Name == nil {
		return fmt.Errorf("invalid nil key received for Name")
	}

	key := *v.Name

	// Initialise the list within the receiver struct if it has not already been
	// created.
	if t.StateEntry == nil {
		t.StateEntry = make(map[string]*YnddState_StateEntry)
	}

	if _, ok := t.StateEntry[key]; ok {
		return fmt.Errorf("duplicate key for list StateEntry %v", key)
	}

	t.StateEntry[key] = v
	return nil
}

// PopulateDefaults recursively populates unset leaf fields in the Device
// with default values as specified in the YANG schema, instantiating any nil
// container fields.
func (t *Device) PopulateDefaults() {
	if t == nil {
		return
	}
	ygot.BuildEmptyTree(t)
	for _, e := range t.StateEntry {
		e.PopulateDefaults()
	}
}

// Validate validates s against the YANG schema corresponding to its type.
func (t *Device) ΛValidate(opts ...ygot.ValidationOption) error {
	if err := ytypes.Validate(SchemaTree["Device"], t, opts...); err != nil {
		return err
	}
	return nil
}

// Validate validates s against the YANG schema corresponding to its type.
func (t *Device) Validate(opts ...ygot.ValidationOption) error {
	return t.ΛValidate(opts...)
}

// ΛEnumTypeMap returns a map, keyed by YANG schema path, of the enumerated types
// that are included in the generated code.
func (t *Device) ΛEnumTypeMap() map[string][]reflect.Type { return ΛEnumTypes }

// ΛBelongingModule returns the name of the module that defines the namespace
// of Device.
func (*Device) ΛBelongingModule() string {
	return ""
}

// YnddState_StateEntry represents the /yndd-state/stateEntry YANG schema element.
type YnddState_StateEntry struct {
	Name   *string  `path:"name" module:"yndd-state"`
	Path   []string `path:"path" module:"yndd-state"`
	Prefix *string  `path:"prefix" module:"yndd-state"`
}

// IsYANGGoStruct ensures that YnddState_StateEntry implements the yang.GoStruct
// interface. This allows functions that need to handle this struct to
// identify it as being generated by ygen.
func (*YnddState_StateEntry) IsYANGGoStruct() {}

// PopulateDefaults recursively populates unset leaf fields in the YnddState_StateEntry
// with default values as specified in the YANG schema, instantiating any nil
// container fields.
func (t *YnddState_StateEntry) PopulateDefaults() {
	if t == nil {
		return
	}
	ygot.BuildEmptyTree(t)
}

// ΛListKeyMap returns the keys of the YnddState_StateEntry struct, which is a YANG list entry.
func (t *YnddState_StateEntry) ΛListKeyMap() (map[string]interface{}, error) {
	if t.Name == nil {
		return nil, fmt.Errorf("nil value for key Name")
	}

	return map[string]interface{}{
		"name": *t.Name,
	}, nil
}

// Validate validates s against the YANG schema corresponding to its type.
func (t *YnddState_StateEntry) ΛValidate(opts ...ygot.ValidationOption) error {
	if err := ytypes.Validate(SchemaTree["YnddState_StateEntry"], t, opts...); err != nil {
		return err
	}
	return nil
}

// Validate validates s against the YANG schema corresponding to its type.
func (t *YnddState_StateEntry) Validate(opts ...ygot.ValidationOption) error {
	return t.ΛValidate(opts...)
}

// ΛEnumTypeMap returns a map, keyed by YANG schema path, of the enumerated types
// that are included in the generated code.
func (t *YnddState_StateEntry) ΛEnumTypeMap() map[string][]reflect.Type { return ΛEnumTypes }

// ΛBelongingModule returns the name of the module that defines the namespace
// of YnddState_StateEntry.
func (*YnddState_StateEntry) ΛBelongingModule() string {
	return "yndd-state"
}

var (
	// ySchema is a byte slice contain a gzip compressed representation of the
	// YANG schema from which the Go code was generated. When uncompressed the
	// contents of the byte slice is a JSON document containing an object, keyed
	// on the name of the generated struct, and containing the JSON marshalled
	// contents of a goyang yang.Entry struct, which defines the schema for the
	// fields within the struct.
	ySchema = []byte{
		0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xec, 0x56, 0xcd, 0x6e, 0xf2, 0x30,
		0x10, 0xbc, 0xe7, 0x29, 0x2c, 0x9f, 0xf9, 0x44, 0xa2, 0x2f, 0xfc, 0x34, 0x37, 0x5a, 0xa8, 0x2a,
		0xd1, 0x3f, 0x95, 0x5e, 0x7a, 0xaa, 0xac, 0x78, 0x01, 0xab, 0xe0, 0x20, 0xc7, 0x69, 0x89, 0x2a,
		0xde, 0xbd, 0x4a, 0x1c, 0x42, 0x12, 0x62, 0x93, 0xb4, 0x47, 0xb8, 0xc1, 0x7a, 0xec, 0x19, 0xef,
		0xec, 0x66, 0xfd, 0x6d, 0x21, 0x84, 0x10, 0x7e, 0x24, 0x6b, 0xc0, 0x1e, 0xc2, 0x14, 0x3e, 0x99,
		0x0f, 0xb8, 0xa3, 0xa2, 0x53, 0xc6, 0x29, 0xf6, 0x90, 0x93, 0xfd, 0xbd, 0x09, 0xf8, 0x9c, 0x2d,
		0xb0, 0x87, 0xec, 0x2c, 0x30, 0x66, 0x02, 0x7b, 0x48, 0x1d, 0x91, 0x06, 0x42, 0x49, 0x24, 0x4c,
		0xb8, 0x14, 0x71, 0x29, 0x5e, 0xa2, 0x28, 0x60, 0x3a, 0x65, 0x44, 0x99, 0x2e, 0x0f, 0x57, 0x69,
		0xf3, 0x85, 0x67, 0x01, 0x73, 0xb6, 0x3d, 0x62, 0x2a, 0xb1, 0xc5, 0x9c, 0xd2, 0x7f, 0x29, 0x65,
		0x85, 0x2d, 0x45, 0xcd, 0x82, 0x48, 0xf8, 0x50, 0x7b, 0x82, 0x52, 0x04, 0xf1, 0x57, 0x20, 0x12,
		0x51, 0x78, 0xa3, 0xc8, 0x3a, 0xf5, 0xc0, 0x3b, 0x12, 0x8e, 0xc4, 0x22, 0x5a, 0x03, 0x97, 0xd8,
		0x43, 0x52, 0x44, 0xa0, 0x01, 0x16, 0x50, 0x45, 0x6d, 0x47, 0xe0, 0x5d, 0x29, 0xb2, 0xab, 0xdc,
		0xbc, 0x9a, 0xf8, 0x7c, 0x81, 0xab, 0x6b, 0x6b, 0xae, 0xb3, 0x4f, 0x4a, 0x8a, 0xd2, 0x08, 0xcc,
		0x4c, 0xb0, 0x35, 0xcb, 0x3a, 0x33, 0x9a, 0x98, 0xd2, 0xce, 0x9c, 0xa6, 0x26, 0xb5, 0x36, 0xab,
		0xb5, 0x69, 0xad, 0xcd, 0xab, 0x37, 0x51, 0x63, 0x66, 0x7e, 0xfa, 0x6b, 0xbc, 0x81, 0x66, 0x79,
		0x0b, 0xa5, 0x60, 0x7c, 0x61, 0xca, 0xd9, 0xbe, 0x95, 0x86, 0x56, 0x33, 0x5d, 0x35, 0x9a, 0xf0,
		0x86, 0xc8, 0xe5, 0xe9, 0x5a, 0x4a, 0x51, 0x97, 0x5a, 0x3a, 0xdf, 0x5a, 0xd2, 0x28, 0xb8, 0x67,
		0xa1, 0x1c, 0x49, 0x29, 0xcc, 0x2a, 0x1e, 0x18, 0x9f, 0xac, 0x20, 0xc9, 0x43, 0xa8, 0xaf, 0x03,
		0x85, 0x24, 0xdb, 0x02, 0xd2, 0x19, 0xba, 0x6e, 0x7f, 0xe0, 0xba, 0xf6, 0xe0, 0xff, 0xc0, 0xbe,
		0xea, 0xf5, 0x9c, 0xbe, 0xd3, 0x33, 0x6c, 0x7e, 0x12, 0x14, 0x04, 0xd0, 0xeb, 0x64, 0x2c, 0xf1,
		0x68, 0xb5, 0xfa, 0x4b, 0x57, 0x98, 0x6b, 0xf2, 0xd0, 0x17, 0xc6, 0x79, 0x71, 0xe9, 0x8c, 0xb3,
		0xfc, 0xca, 0x1a, 0x87, 0xfa, 0x14, 0x62, 0xcd, 0x70, 0x36, 0x37, 0xd3, 0xe9, 0x26, 0xfa, 0x55,
		0xf3, 0x98, 0x9b, 0xa6, 0x2a, 0x7e, 0xc4, 0x79, 0x20, 0x89, 0x64, 0x01, 0xaf, 0xd7, 0x18, 0xfa,
		0x4b, 0x58, 0x93, 0x6c, 0xa4, 0xe0, 0xee, 0xc1, 0xd7, 0xae, 0xf6, 0x3d, 0x98, 0xbd, 0x28, 0x45,
		0xe4, 0xcb, 0xec, 0x59, 0x83, 0xdf, 0x38, 0xa5, 0xb3, 0x04, 0xff, 0x3e, 0x3b, 0xec, 0xb2, 0xea,
		0x53, 0xac, 0x7e, 0x65, 0x3a, 0x75, 0xfa, 0x30, 0x0b, 0x6f, 0xc9, 0x07, 0xbc, 0x04, 0xc1, 0x71,
		0x71, 0x56, 0x35, 0xe3, 0xe2, 0x52, 0x49, 0xd6, 0x58, 0xbd, 0x9a, 0x15, 0xa1, 0xb5, 0xfb, 0x01,
		0x00, 0x00, 0xff, 0xff, 0x01, 0x00, 0x00, 0xff, 0xff, 0x7a, 0xcb, 0xfb, 0x88, 0x54, 0x0b, 0x00,
		0x00,
	}
)

// ΛEnumTypes is a map, keyed by a YANG schema path, of the enumerated types that
// correspond with the leaf. The type is represented as a reflect.Type. The naming
// of the map ensures that there are no clashes with valid YANG identifiers.
func initΛEnumTypes() {
	ΛEnumTypes = map[string][]reflect.Type{}
}