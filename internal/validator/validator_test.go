package validator

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/protobuf/proto"
)

// TestCleanPath tests the CleanPath function
func TestCleanPath(t *testing.T) {
	tests := []struct {
		description string
		got         *gnmi.Path
		want        *gnmi.Path
	}{
		{
			description: "simple test, nothing to cleanup",
			got: &gnmi.Path{
				Elem: []*gnmi.PathElem{{Name: "Foo"}, {Name: "Bar"}},
			},
			want: &gnmi.Path{
				Elem: []*gnmi.PathElem{{Name: "Foo"}, {Name: "Bar"}},
			},
		},
		{
			description: "remove namespace from elem",
			got: &gnmi.Path{
				Elem: []*gnmi.PathElem{{Name: "Foo:bar"}, {Name: "bla:Bar"}},
			},
			want: &gnmi.Path{
				Elem: []*gnmi.PathElem{{Name: "bar"}, {Name: "Bar"}},
			},
		},
		{
			description: "remove namespace from Key and key value",
			got: &gnmi.Path{
				Elem: []*gnmi.PathElem{{Name: "bar"}, {Name: "Bar", Key: map[string]string{"foo:Key1": "foo:data"}}},
			},
			want: &gnmi.Path{
				Elem: []*gnmi.PathElem{{Name: "bar"}, {Name: "Bar", Key: map[string]string{"Key1": "data"}}},
			},
		},
		{
			description: "remove namespace from Key and key value",
			got: &gnmi.Path{
				Elem: []*gnmi.PathElem{{Name: "bar"}, {Name: "Bar", Key: map[string]string{"foo::Key1": "foo:data"}}},
			},
			want: &gnmi.Path{
				Elem: []*gnmi.PathElem{{Name: "bar"}, {Name: "Bar", Key: map[string]string{"Key1": "data"}}},
			},
		},
	}

	for _, test := range tests {
		result := cleanPath(test.got)
		if diff := cmp.Diff(test.want, result, cmp.Comparer(proto.Equal)); diff != "" {
			t.Errorf("CleanedPath (%s) returned differ (-want, +got):\n%s", test.description, diff)
		}
	}
}

// func TestValidateCreate(t *testing.T) {
// 	tests := []struct {
// 		description string
// 		got         struct {
// 			ce config.ConfigEntry
// 			x  interface{}
// 		}
// 		want struct {
// 			ygotValidatedStruct ygot.ValidatedGoStruct
// 			err                 error
// 		}
// 	}{
// 		{
// 			description: "test one",
// 			got: struct {
// 				ce config.ConfigEntry
// 				x  interface{}
// 			}{
// 				ce: config.NewConfigEntry(),
// 				x: func() interface{} {
// 					d := &ygotnddpstate.Device{}
// 					d.GetOrCreateStateEntry("Foo").Description = ygot.String("This is the description")
// 					d.GetOrCreateStateEntry("Foo").AdminState = ygotnddpstate.NddpCommon_AdminState_enable
// 					d.GetOrCreateStateEntry("Foo").Name = ygot.String("BLUBB")

// 					return d
// 				}(),
// 			},
// 			want: struct {
// 				ygotValidatedStruct ygot.ValidatedGoStruct
// 				err                 error
// 			}{
// 				ygotValidatedStruct: func() ygot.ValidatedGoStruct {
// 					d := &ygotnddpstate.Device{}
// 					// d.GetOrCreateStateEntry("Foo").Description = ygot.String("This is the description")
// 					// d.GetOrCreateStateEntry("Foo").AdminState = ygotnddpstate.NddpCommon_AdminState_enable
// 					// d.GetOrCreateStateEntry("Foo").Name = ygot.String("BLUBB")

// 					return d
// 				}(),
// 				err: nil,
// 			},
// 		},
// 	}
// 	for _, test := range tests {
// 		ygotvalgostr, err := ValidateCreate(test.got.ce, test.got.x)

// 		if err != test.want.err {
// 			t.Errorf("Expected error \"%s\" but was \"%s\"", test.want.err, err)
// 		}

// 		if diff := cmp.Diff(test.want.ygotValidatedStruct, ygotvalgostr, cmp.Comparer(proto.Equal)); diff != "" {
// 			t.Errorf("TestValidateCreate (%s) expected and actual results differ (-want, +got):\n%s", test.description, diff)
// 		}
// 	}
// }
