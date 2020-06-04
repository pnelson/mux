package mux

import (
	"reflect"
	"testing"
)

func TestNodeAddStatic(t *testing.T) {
	var tests = []string{
		"/romane",
		"/romanus",
		"/romulus",
		"/rubens",
		"/ruber",
		"/rubicon",
		"/rubicundus",
	}
	n := &node{}
	for _, pattern := range tests {
		r := NewRoute(pattern, nil)
		err := n.add(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	assertNode(t, n, "/r")
	assertEdges(t, n, []string{"om", "ub"})
	assertEdges(t, n.edges[0], []string{"an", "ulus"})
	assertEdges(t, n.edges[0].edges[0], []string{"e", "us"})
	assertRoute(t, n.edges[0].edges[0].edges[0], "e", "/romane")
	assertRoute(t, n.edges[0].edges[0].edges[1], "us", "/romanus")
	assertEdges(t, n.edges[0].edges[0].edges[0], nil)
	assertEdges(t, n.edges[0].edges[0].edges[1], nil)
	assertRoute(t, n.edges[0].edges[1], "ulus", "/romulus")
	assertEdges(t, n.edges[1], []string{"e", "ic"})
	assertEdges(t, n.edges[1].edges[0], []string{"ns", "r"})
	assertRoute(t, n.edges[1].edges[0].edges[0], "ns", "/rubens")
	assertRoute(t, n.edges[1].edges[0].edges[1], "r", "/ruber")
	assertEdges(t, n.edges[1].edges[0].edges[0], nil)
	assertEdges(t, n.edges[1].edges[0].edges[1], nil)
	assertEdges(t, n.edges[1].edges[1], []string{"on", "undus"})
	assertRoute(t, n.edges[1].edges[1].edges[0], "on", "/rubicon")
	assertRoute(t, n.edges[1].edges[1].edges[1], "undus", "/rubicundus")
	assertEdges(t, n.edges[1].edges[1].edges[0], nil)
	assertEdges(t, n.edges[1].edges[1].edges[1], nil)
}

func TestNodeAddStaticNested(t *testing.T) {
	var tests = []string{
		"/a",
		"/a/b/",
		"/a/b/c",
		"/a/b/c/d",
	}
	n := &node{}
	for _, pattern := range tests {
		r := NewRoute(pattern, nil)
		err := n.add(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	assertRoute(t, n, "/a", "/a")
	assertEdges(t, n, []string{"/b/"})
	assertRoute(t, n.edges[0], "/b/", "/a/b/")
	assertEdges(t, n.edges[0], []string{"c"})
	assertRoute(t, n.edges[0].edges[0], "c", "/a/b/c")
	assertEdges(t, n.edges[0].edges[0], []string{"/d"})
	assertRoute(t, n.edges[0].edges[0].edges[0], "/d", "/a/b/c/d")
	assertEdges(t, n.edges[0].edges[0].edges[0], nil)
}

func TestNodeAddParam(t *testing.T) {
	var tests = []string{
		"/z",     // /z
		"/:a",    // /a
		"/:a/",   // /a/
		"/:a/z",  // /a/z
		"/:a/:b", // /a/b
		"/:a/*",  // /a/b/c
		"/",      // /
	}
	n := &node{}
	for _, pattern := range tests {
		r := NewRoute(pattern, nil)
		err := n.add(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	assertRoute(t, n, "/", "/")
	assertEdges(t, n, []string{"z", ":a"})
	assertRoute(t, n.edges[0], "z", "/z")
	assertEdges(t, n.edges[0], nil)
	assertRoute(t, n.edges[1], ":a", "/:a")
	assertEdges(t, n.edges[1], []string{"/"})
	assertRoute(t, n.edges[1].edges[0], "/", "/:a/")
	assertEdges(t, n.edges[1].edges[0], []string{"z", ":b", "*"})
	assertRoute(t, n.edges[1].edges[0].edges[0], "z", "/:a/z")
	assertEdges(t, n.edges[1].edges[0].edges[0], nil)
	assertRoute(t, n.edges[1].edges[0].edges[1], ":b", "/:a/:b")
	assertEdges(t, n.edges[1].edges[0].edges[1], nil)
	assertRoute(t, n.edges[1].edges[0].edges[2], "*", "/:a/*")
	assertEdges(t, n.edges[1].edges[0].edges[2], nil)
}

func TestNodeAddDuplicate(t *testing.T) {
	r := NewRoute("/", nil)
	n := &node{}
	err := n.add(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = n.add(r)
	if err == nil {
		t.Fatalf("should not add duplicate pattern route")
	}
}

// Patterns borrowed from Goji as a reference.
// https://github.com/goji/goji/blob/0d89ff54b2c18c9c4ba530e32496aef902d3c6cd/pat/pat_test.go
func TestNodeMatch(t *testing.T) {
	var tests = []struct {
		pattern string
		path    string
		match   bool
		params  Params
	}{
		{"/", "/", true, nil},
		{"/", "/hello", false, nil},
		{"/hello", "/hello", true, nil},

		{"/:name", "/carl", true, Params{"name": "carl"}},
		{"/:name", "/carl/", false, nil},
		{"/:name", "/", false, nil},
		{"/:name/", "/carl/", true, Params{"name": "carl"}},
		{"/:name/", "/carl/no", false, nil},
		{"/:name/hi", "/carl/hi", true, Params{"name": "carl"}},
		{"/:name/:color", "/carl/red", true, Params{"name": "carl", "color": "red"}},
		{"/:name/:color", "/carl/", false, nil},
		{"/:name/:color", "/carl.red", false, nil},

		{"/:file.:ext", "/data.json", true, Params{"file": "data", "ext": "json"}},
		{"/:file.:ext", "/data.tar.gz", true, Params{"file": "data", "ext": "tar.gz"}},
		{"/:file.:ext", "/data", false, nil},
		{"/:file.:ext", "/data.", false, nil},
		{"/:file.:ext", "/.gitconfig", false, nil},
		{"/:file.:ext", "/data.json/", false, nil},
		{"/:file.:ext", "/data/json", false, nil},
		{"/:file.:ext", "/data;json", false, nil},
		{"/hello.:ext", "/hello.json", true, Params{"ext": "json"}},
		{"/:file.json", "/hello.json", true, Params{"file": "hello"}},
		{"/:file.json", "/hello.world.json", false, nil},
		{"/file;:version", "/file;1", true, Params{"version": "1"}},
		{"/file;:version", "/file,1", false, nil},
		{"/file,:version", "/file,1", true, Params{"version": "1"}},
		{"/file,:version", "/file;1", false, nil},

		{"/*", "/", true, Params{"*": ""}},
		{"/*", "/hello", true, Params{"*": "hello"}},
		{"/users/*", "/", false, nil},
		{"/users/*", "/users", false, nil},
		{"/users/*", "/users/", true, Params{"*": ""}},
		{"/users/*", "/users/carl", true, Params{"*": "carl"}},
		{"/users/*", "/profile/carl", false, nil},
		{"/:name/*", "/carl", false, nil},
		{"/:name/*", "/carl/", true, Params{"name": "carl", "*": ""}},
		{"/:name/*", "/carl/photos", true, Params{"name": "carl", "*": "photos"}},
		{"/:name/*", "/carl/photos%2f2015", true, Params{"name": "carl", "*": "photos%2f2015"}},
	}
	for _, tt := range tests {
		n := &node{}
		r := NewRoute(tt.pattern, nil)
		err := n.add(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var want error
		if !tt.match {
			want = ErrNotFound
		}
		_, params, err := n.match(tt.path)
		if err != want {
			t.Fatalf("match\npattern '%s'\npath '%s'\nhave %#v\nwant %#v", tt.pattern, tt.path, err, want)
		}
		if !reflect.DeepEqual(params, tt.params) {
			if tt.params == nil && reflect.DeepEqual(params, Params{}) {
				continue
			}
			t.Fatalf("match\nin '%s'\npath '%s'\nhave %v\nwant %v", tt.pattern, tt.path, params, tt.params)
		}
	}
}

func assertLabel(t *testing.T, n *node, label string) {
	have := n.label
	if n.param && n.label != "*" {
		have = ":" + have
	}
	assertString(t, "label", have, label)
}

func assertNode(t *testing.T, n *node, label string) {
	if n.route != nil {
		t.Fatalf("should be nil route\nhave '%s' => '%s'", label, n.route.pattern)
	}
	assertLabel(t, n, label)
}

func assertRoute(t *testing.T, n *node, label, pattern string) {
	if n.route == nil {
		t.Fatalf("should be non-nil route for '%s'", pattern)
	}
	assertLabel(t, n, label)
	assertString(t, "route", n.route.pattern, pattern)
}

func assertEdges(t *testing.T, n *node, edges []string) {
	var have []string
	for _, e := range n.edges {
		label := e.label
		if e.param && e.label != "*" {
			label = ":" + label
		}
		have = append(have, label)
	}
	assertDeepEqual(t, "edges", have, edges)
}
