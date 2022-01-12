package mux

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
)

// empty represents the empty method set.
const empty = "EMPTY"

// tree is the default router implementation.
// tree also implements the Builder and Walker interfaces.
type tree struct {
	names   map[string]*Route
	methods map[string]*node
}

// Add adds r to the tree.
func (t *tree) Add(r *Route) error {
	name := r.Name()
	if name != "" {
		if t.names == nil {
			t.names = make(map[string]*Route)
		}
		_, ok := t.names[name]
		if ok {
			return fmt.Errorf("mux: duplicate named route '%s'", name)
		}
		t.names[name] = r
	}
	methods := r.Methods()
	if len(methods) == 0 {
		methods = append(methods, empty)
	}
	if t.methods == nil {
		t.methods = make(map[string]*node)
	}
	for _, method := range methods {
		_, ok := t.methods[method]
		if !ok {
			t.methods[method] = &node{}
		}
		err := t.methods[method].add(r)
		if err != nil {
			return err
		}
	}
	return nil
}

// Build returns the URL for the named route or an error if
// the named route does not exist or a parameter is missing.
//
// Build implements the Builder interface.
func (t *tree) Build(name string, params Params) (string, error) {
	r, ok := t.names[name]
	if !ok {
		return "", ErrBuild
	}
	var buf strings.Builder
	pattern := r.Pattern()
	for pattern != "" {
		b := pattern[0]
		switch b {
		case ':':
			var i int
			for i = 1; i < len(pattern); i++ {
				if isBreak(pattern[i]) {
					break
				}
			}
			key := pattern[1:i]
			v, ok := params[key]
			if !ok {
				return "", ErrBuild
			}
			buf.WriteString(v)
			pattern = pattern[i:]
		case '*':
			key := string(b)
			v, ok := params[key]
			if ok {
				buf.WriteString(v)
			}
			pattern = pattern[1:]
		default:
			var i int
			for i = 1; i < len(pattern); i++ {
				if isParam(pattern[i]) {
					break
				}
			}
			buf.WriteString(pattern[:i])
			pattern = pattern[i:]
		}
	}
	return buf.String(), nil
}

// Match returns the matching route and parameters.
func (t *tree) Match(req *http.Request) (*Route, Params, error) {
	path := req.URL.EscapedPath()
	root, ok := t.methods[req.Method]
	if ok {
		route, params, err := root.match(path)
		if err == nil {
			return route, params, nil
		}
	}
	root, ok = t.methods[empty]
	if ok {
		route, params, err := root.match(path)
		if err == nil {
			return route, params, nil
		}
	}
	allowed := make([]string, 0)
	for method := range t.methods {
		if method == req.Method || method == http.MethodOptions {
			continue
		}
		root = t.methods[method]
		_, _, err := root.match(path)
		if err == nil {
			allowed = append(allowed, method)
		}
	}
	if len(allowed) > 0 {
		allowed = append(allowed, http.MethodOptions)
		sort.Strings(allowed)
		return nil, nil, ErrMethodNotAllowed(allowed)
	}
	return nil, nil, ErrNotFound
}

// Walk yields named routes to fn sorting alphabetically.
//
// Walk implements the Walker interface.
func (t *tree) Walk(fn WalkFunc) error {
	names := make([]string, 0)
	for name := range t.names {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		err := fn(t.names[name])
		if err != nil {
			return err
		}
	}
	return nil
}
