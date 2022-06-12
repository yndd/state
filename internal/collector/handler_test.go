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

func TestXPathToSubject(t *testing.T) {
	type args struct {
		p string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "empty",
			args:    args{p: ""},
			want:    "",
			wantErr: false,
		},
		{
			name:    "single_elem",
			args:    args{p: "foo"},
			want:    "foo.>",
			wantErr: false,
		},
		{
			name:    "two_elems",
			args:    args{p: "/foo/bar"},
			want:    "foo.bar.>",
			wantErr: false,
		},
		{
			name:    "elem_with_key",
			args:    args{p: "foo[k=v]"},
			want:    "foo.{k=v}.>",
			wantErr: false,
		},
		{
			name:    "elem_with_wildcard_key_at_the_end",
			args:    args{p: "foo[k=*]"},
			want:    "foo.>",
			wantErr: false,
		},
		{
			name:    "two_elems_with_wildcard_key",
			args:    args{p: "/foo[k=*]/bar"},
			want:    "foo.*.bar.>",
			wantErr: false,
		},
		{
			name:    "elem_with_two_keys",
			args:    args{p: "foo[k2=v2][k1=v1]"},
			want:    "foo.{k1=v1}.{k2=v2}.>",
			wantErr: false,
		},
		{
			name:    "two_elems_with_two_keys",
			args:    args{p: "foo[k2=v2][k1=v1]/bar"},
			want:    "foo.{k1=v1}.{k2=v2}.bar.>",
			wantErr: false,
		},
		{
			name:    "two_elems_with_two_keys_each",
			args:    args{p: "foo[k2=v2][k1=v1]/bar[a=1][b=2]"},
			want:    "foo.{k1=v1}.{k2=v2}.bar.{a=1}.{b=2}.>",
			wantErr: false,
		},
		{
			name:    "two_elems_with_two_keys_one_wildcard",
			args:    args{p: "foo[k1=*][k2=v2]/bar"},
			want:    "foo.*.{k2=v2}.bar.>",
			wantErr: false,
		},
		{
			name:    "two_elems_with_two_keys_both_wildcards",
			args:    args{p: "foo[a=*][b=*]/bar"},
			want:    "foo.*.*.bar.>",
			wantErr: false,
		},
		{
			name:    "two_elems_with_two_keys_each_one_wildcard",
			args:    args{p: "foo[k2=v2][k1=*]/bar[a=1][b=*]"},
			want:    "foo.*.{k2=v2}.bar.{a=1}.>",
			wantErr: false,
		},
		{
			name:    "path_with_origin",
			args:    args{p: "origin:/foo"},
			want:    "origin.foo.>",
			wantErr: false,
		},
		{
			name:    "only_origin",
			args:    args{p: "origin:/"},
			want:    "origin.>",
			wantErr: false,
		},
		{
			name:    "only_origin_no_slash",
			args:    args{p: "origin:"},
			want:    "origin.>",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := XPathToSubject(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("XPathToSubject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("XPathToSubject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkXPathToSubject(b *testing.B) {
	for i := 0; i < b.N; i++ {
		XPathToSubject("origin:/foo[k2=v2][k1=*]/bar[a=1][b=*]")
	}
}
