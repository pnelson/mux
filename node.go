package mux

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type node struct {
	param bool
	label string
	route *Route
	edges []*node
}

func (n *node) add(r *Route) error {
	pattern := r.Pattern()
	return n.addNode(r, pattern)
}

func (n *node) addNode(r *Route, pattern string) error {
	if pattern == "" {
		if n.route != nil {
			return fmt.Errorf("mux: duplicate route pattern '%s'", r.pattern)
		}
		n.route = r
		return nil
	}
	switch pattern[0] {
	case ':':
		return n.addParam(r, pattern)
	case '*':
		return n.addSplat(r, pattern)
	}
	return n.addStatic(r, pattern)
}

func (n *node) addEdge(e *node) *node {
	n.edges = append(n.edges, e)
	sort.Slice(n.edges, func(i, j int) bool {
		if n.edges[i].param {
			if n.edges[j].param {
				return n.edges[j].label == "*"
			}
			return false
		}
		if n.edges[j].param {
			return true
		}
		return n.edges[i].label < n.edges[j].label
	})
	return e
}

func (n *node) addStatic(r *Route, pattern string) error {
	i := prefixIndex(n.label, pattern)
	j := 0
	for j < len(pattern) {
		if isParam(pattern[j]) && isBreak(pattern[j-1]) {
			break
		}
		j++
	}
	if n.label == "" {
		n.label = pattern[:j]
		return n.addNode(r, pattern[j:])
	}
	if n.label == "/" && pattern == "/" {
		return n.addNode(r, "")
	}
	if i == 0 || i == len(n.label) {
		prefix := pattern[i]
		for _, e := range n.edges {
			if e.param || e.label[0] > prefix {
				break
			}
			if prefixIndex(e.label, pattern[i:]) > 0 {
				return e.addNode(r, pattern[i:])
			}
		}
	} else {
		s := n.addEdge(&node{label: n.label[i:], route: n.route, edges: append(n.edges[:0], n.edges...)})
		n.label = pattern[:i]
		n.route = nil
		n.edges = []*node{s}
	}
	if j > i {
		n = n.addEdge(&node{label: pattern[i:j]})
	}
	return n.addNode(r, pattern[j:])
}

func (n *node) addParam(r *Route, pattern string) error {
	i := len(pattern)
	for j := 1; j < len(pattern); j++ {
		if isParam(pattern[j]) {
			return fmt.Errorf("mux: invalid param '%s'", pattern[:j])
		}
		if isBreak(pattern[j]) {
			i = j
			break
		}
	}
	for _, e := range n.edges {
		if e.param {
			return e.addNode(r, pattern[i:])
		}
	}
	n = n.addEdge(&node{param: true, label: pattern[1:i]})
	return n.addNode(r, pattern[i:])
}

func (n *node) addSplat(r *Route, pattern string) error {
	if pattern != "*" {
		return errors.New("mux: splat only supported at terminus")
	}
	n = n.addEdge(&node{param: true, label: pattern[:1]})
	return n.addNode(r, pattern[1:])
}

func (n *node) match(path string) (*Route, Params, error) {
	var params Params
	m := n.search(path)
	if m == nil {
		return nil, nil, ErrNotFound
	}
	pattern := m.route.pattern
	for {
		i := prefixIndex(pattern, path)
		path = path[i:]
		pattern = pattern[i:]
		if params == nil {
			params = make(Params)
		}
		if pattern == "*" {
			params[pattern] = path
			break
		}
		var bc byte
		var j, k int
		if path == "" {
			break
		}
		for ; j < len(pattern); j++ {
			if isBreak(pattern[j]) {
				bc = pattern[j]
				break
			}
		}
		for ; k < len(path); k++ {
			if path[k] == bc {
				break
			}
		}
		key := pattern[1:j]
		params[key] = path[:k]
		path = path[k:]
		pattern = pattern[j:]
	}
	return m.route, params, nil
}

func (n *node) search(path string) *node {
	if n.param && n.label == "*" {
		return n
	}
	if n.param {
		i := 0
		bc := make(map[byte]struct{})
		for _, e := range n.edges {
			bc[e.label[0]] = struct{}{}
		}
		if len(bc) == 0 {
			bc['/'] = struct{}{}
		}
	loop:
		for ; i < len(path); i++ {
			for c := range bc {
				if path[i] == c {
					break loop
				}
			}
		}
		if i == 0 {
			return nil
		}
		path = path[i:]
	} else {
		if !strings.HasPrefix(path, n.label) {
			return nil
		}
		path = path[len(n.label):]
		if path == "" && n.route != nil {
			return n
		}
	}
	for _, e := range n.edges {
		p := e.search(path)
		if p != nil {
			return p
		}
	}
	if path != "" || n.route == nil {
		return nil
	}
	return n
}

func (n *node) walk(fn WalkFunc) error {
	if n.route != nil {
		err := fn(n.route)
		if err != nil {
			return err
		}
	}
	for _, e := range n.edges {
		err := e.walk(fn)
		if err != nil {
			return err
		}
	}
	return nil
}

func isParam(b byte) bool {
	switch b {
	case ':', '*':
		return true
	}
	return false
}

func isBreak(b byte) bool {
	switch b {
	case '+', ',', '-', '.', '/', ';', '_', '~':
		return true
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func prefixIndex(a, b string) int {
	n := min(len(a), len(b))
	for i := 0; i < n; i++ {
		switch {
		case a[i] != b[i]:
			return i
		case i > 0 && isParam(a[i]) && isBreak(a[i-1]):
			return i
		}
	}
	return n
}
