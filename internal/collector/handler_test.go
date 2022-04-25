package collector

import (
	"testing"

	"github.com/openconfig/gnmi/proto/gnmi"
)

func Test_gNMIPathToSubject(t *testing.T) {
	type args struct {
		p *gnmi.Path
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			// /foo
			name: "simple_path",
			args: args{
				p: &gnmi.Path{
					Origin: "",
					Elem: []*gnmi.PathElem{{
						Name: "foo",
					}},
				},
			},
			want: "foo",
		},
		{
			// /foo/bar
			name: "path_with_two_elems",
			args: args{
				p: &gnmi.Path{
					Origin: "",
					Elem: []*gnmi.PathElem{
						{
							Name: "foo",
						},
						{
							Name: "bar",
						},
					},
				},
			},
			want: "foo.bar",
		},
		{
			// origin:/foo/bar
			name: "path_with_origin",
			args: args{
				p: &gnmi.Path{
					Origin: "org1",
					Elem: []*gnmi.PathElem{
						{
							Name: "foo",
						},
						{
							Name: "bar",
						},
					},
				},
			},
			want: "org1.foo.bar",
		},
		{
			// /foo[k=v]
			name: "path_with_key",
			args: args{
				p: &gnmi.Path{
					Origin: "",
					Elem: []*gnmi.PathElem{{
						Name: "foo",
						Key: map[string]string{
							"k": "v",
						},
					}},
				},
			},
			want: "foo.{k=v}",
		},
		{
			// /foo[k1=v1][k2=v2]
			name: "path_with_two_keys",
			args: args{
				p: &gnmi.Path{
					Origin: "",
					Elem: []*gnmi.PathElem{{
						Name: "foo",
						Key: map[string]string{
							"k1": "v1",
							"k2": "v2",
						},
					}},
				},
			},
			want: "foo.{k1=v1}.{k2=v2}",
		},
		{
			// /foo[k1=v1][k2=v2]/bar
			name: "path_with_two_elems_two_keys",
			args: args{
				p: &gnmi.Path{
					Origin: "",
					Elem: []*gnmi.PathElem{
						{
							Name: "foo",
							Key: map[string]string{
								"k1": "v1",
								"k2": "v2",
							},
						}, {
							Name: "bar",
						},
					},
				},
			},
			want: "foo.{k1=v1}.{k2=v2}.bar",
		},
		{
			// /foo[k1=v1][k2=v2]/bar[k3=v3][k4=v3]
			name: "path_with_two_elems_two_keys_each",
			args: args{
				p: &gnmi.Path{
					Origin: "",
					Elem: []*gnmi.PathElem{
						{
							Name: "foo",
							Key: map[string]string{
								"k1": "v1",
								"k2": "v2",
							},
						}, {
							Name: "bar",
							Key: map[string]string{
								"k3": "v3",
								"k4": "v4",
							},
						},
					},
				},
			},
			want: "foo.{k1=v1}.{k2=v2}.bar.{k3=v3}.{k4=v4}",
		},
		{
			// /foo[k1=1.1.1.1]
			name: "path_with_elem_key_containing_dot",
			args: args{
				p: &gnmi.Path{
					Origin: "",
					Elem: []*gnmi.PathElem{
						{
							Name: "foo",
							Key: map[string]string{
								"k1": "1.1.1.1",
							},
						},
					},
				},
			},
			want: "foo.{k1=1^1^1^1}",
		},
		{
			// /foo[k1=1..1.1.1]
			name: "path_with_elem_key_containing_consecutive_dots",
			args: args{
				p: &gnmi.Path{
					Origin: "",
					Elem: []*gnmi.PathElem{
						{
							Name: "foo",
							Key: map[string]string{
								"k1": "1..1.1.1",
							},
						},
					},
				},
			},
			want: "foo.{k1=1^^1^1^1}",
		},
		{
			// /foo[k1=v 1]
			name: "path_with_elem_key_containing_space",
			args: args{
				p: &gnmi.Path{
					Origin: "",
					Elem: []*gnmi.PathElem{
						{
							Name: "foo",
							Key: map[string]string{
								"k1": "v 1",
							},
						},
					},
				},
			},
			want: "foo.{k1=v~1}",
		},
		{
			// /foo[k1=v  1]
			name: "path_with_elem_key_containing_consecutive_spaces",
			args: args{
				p: &gnmi.Path{
					Origin: "",
					Elem: []*gnmi.PathElem{
						{
							Name: "foo",
							Key: map[string]string{
								"k1": "v  1",
							},
						},
					},
				},
			},
			want: "foo.{k1=v~~1}",
		},
		{
			// /foo[k1=1.1.1.1][k2=2.2.2.2]
			name: "path_with_elem_two_keys_containing_dot",
			args: args{
				p: &gnmi.Path{
					Origin: "",
					Elem: []*gnmi.PathElem{
						{
							Name: "foo",
							Key: map[string]string{
								"k1": "1.1.1.1",
								"k2": "2.2.2.2",
							},
						},
					},
				},
			},
			want: "foo.{k1=1^1^1^1}.{k2=2^2^2^2}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := gNMIPathToSubject(tt.args.p); got != tt.want {
				t.Errorf("gNMIPathToSubject() = %v, want %v", got, tt.want)
			}
		})
	}
}
