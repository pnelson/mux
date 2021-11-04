package mux

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var (
	_ Router  = &tree{}
	_ Builder = &tree{}
	_ Walker  = &tree{}
)

func TestTreeAddDuplicateName(t *testing.T) {
	r1 := NewRoute("/foo", nil, WithName("test"))
	r2 := NewRoute("/bar", nil, WithName("test"))
	tree := &tree{}
	err := tree.Add(r1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = tree.Add(r2)
	if err == nil {
		t.Fatalf("should not add duplicate named route")
	}
}

func TestTreeAddDuplicateNodeParam(t *testing.T) {
	r1 := NewRoute("/:foo", nil)
	r2 := NewRoute("/:bar", nil)
	tree := &tree{}
	err := tree.Add(r1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = tree.Add(r2)
	if err == nil {
		t.Fatalf("should not add duplicate param route")
	}
}

func TestTreeBuild(t *testing.T) {
	var tests = []struct {
		name    string
		pattern string
		params  Params
		want    string
	}{
		{"index", "/", nil, "/"},
		{"param", "/posts/:slug", Params{"slug": "test"}, "/posts/test"},
		{"wildcard", "/files/*", Params{"*": "path/to/file.txt"}, "/files/path/to/file.txt"},
		{"param+param", "/:a/:b", Params{"a": "a", "b": "b"}, "/a/b"},
		{"param+wildcard", "/:a/:b/*", Params{"a": "a", "b": "b", "*": "c/d"}, "/a/b/c/d"},
	}
	tree := &tree{}
	for _, tt := range tests {
		r := NewRoute(tt.pattern, nil, WithName(tt.name))
		err := tree.Add(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	for _, tt := range tests {
		have, err := tree.Build(tt.name, tt.params)
		if err != nil {
			t.Fatalf("unexpected error: %v\n for '%s'", err, tt.name)
		}
		if have != tt.want {
			t.Fatalf("Build\npattern '%s'\nparams %#v\nhave '%s'\nwant '%s'", tt.pattern, tt.params, have, tt.want)
		}
	}
}

func TestTreeBuildNilWildcard(t *testing.T) {
	tree := &tree{}
	r := NewRoute("/*", nil, WithName("index"))
	err := tree.Add(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = tree.Build("index", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTreeMatch(t *testing.T) {
	var tests = []struct {
		pattern string
		path    string
		params  Params
	}{
		{"/romane", "/romane", nil},
		{"/romanus", "/romanus", nil},
		{"/romulus", "/romulus", nil},
		{"/rubens", "/rubens", nil},
		{"/ruber", "/ruber", nil},
		{"/rubicon", "/rubicon", nil},
		{"/rubicundus", "/rubicundus", nil},
		{"/s", "/s", nil},
		{"/s/a/", "/s/a/", nil},
		{"/s/a/b", "/s/a/b", nil},
		{"/p/:a", "/p/a", Params{"a": "a"}},
		{"/p/:a/", "/p/a/", Params{"a": "a"}},
		{"/p/:a/z", "/p/a/z", Params{"a": "a"}},
		{"/p/:a/:b", "/p/a/b", Params{"a": "a", "b": "b"}},
		{"/p/:a/*", "/p/a/b/c", Params{"a": "a", "*": "b/c"}},
		{"/", "/", nil},
	}
	tree := &tree{}
	for _, tt := range tests {
		r := NewRoute(tt.pattern, nil)
		err := tree.Add(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		r, params, err := tree.Match(req)
		if err != nil {
			t.Fatalf("unexpected error: %v\n for '%s'", err, tt.path)
		}
		have := r.Pattern()
		if have != tt.pattern {
			t.Fatalf("Match\npath '%s'\nparams %#v\nhave '%s'\nwant '%s'", tt.path, params, have, tt.pattern)
		}
	}
}

func TestTreeMatchNotFound(t *testing.T) {
	r := NewRoute("/romulus", nil)
	tree := &tree{}
	err := tree.Add(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/remus", nil)
	_, _, err = tree.Match(req)
	if err != ErrNotFound {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTreeWalk(t *testing.T) {
	patterns := []string{"/rubicon", "/ruber", "/rubicundus", "/romanus", "/romane", "/romulus", "/rubens"}
	tree := &tree{}
	for _, pattern := range patterns {
		r := NewRoute(pattern, nil, WithName(pattern))
		err := tree.Add(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	have := make([]string, 0)
	want := []string{"/romane", "/romanus", "/romulus", "/rubens", "/ruber", "/rubicon", "/rubicundus"}
	fn := func(r *Route) error {
		name := r.Name()
		have = append(have, name)
		return nil
	}
	err := tree.Walk(fn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(have, want) {
		t.Fatalf("Walk\nhave %v\nwant %v", have, want)
	}
}
