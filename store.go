package slim

import (
	"fmt"
	"strings"
)

type Store struct {
	root *node
}

func NewStore() *Store {
	return &Store{
		root: &node{
			static:    true,
			children:  make([]*node, 256),
			pchildren: []*node{},
			pindex:    -1,
		},
	}
}

func (s *Store) Add(key string, data interface{}) int {
	return s.root.add(key, data)
}

func (s *Store) String() string {
	return s.root.print(0)
}

type node struct {
	key       string
	static    bool
	data      interface{}
	children  []*node
	pchildren []*node
	pindex    int
	pnames    []string
}

func (n *node) add(key string, data interface{}) int {
	matched := 0
	for ; matched < len(key) && matched < len(n.key); matched++ {
		if key[matched] != n.key[matched] {
			break
		}
	}
	if matched == len(n.key) {
		if matched == len(key) {
			// key is the same as the n.key
			if n.data == nil {
				n.data = data
			}
			return n.pindex + 1
		}

		newKey := key[matched:]
		if child := n.children[newKey[0]]; child != nil {
			if pn := child.add(newKey, data); pn >= 0 {
				return pn
			}
		}

		for _, child := range n.pchildren {
			if pn := child.add(newKey, data); pn >= 0 {
				return pn
			}
		}

		return n.addChild(newKey, data)
	}

	if matched == 0 || !n.static {
		return -1
	}

	n1 := &node{
		static:    true,
		key:       n.key[matched:],
		data:      n.data,
		children:  n.children,
		pchildren: n.pchildren,
		pindex:    n.pindex,
		pnames:    n.pnames,
	}

	n.key = key[0:matched]
	n.data = nil
	n.children = make([]*node, 256)
	n.pchildren = []*node{}
	n.children[n1.key[0]] = n1
	return n.add(key, data)
}

func (n *node) addChild(key string, data interface{}) int {
	p0, p1 := -1, -1
	for i := 0; i < len(key); i++ {
		if p0 < 0 && key[i] == ':' {
			p0 = i
		}
		if p0 >= 0 {
			if key[i] == '/' {
				p1 = i
				break
			} else if i == len(key)-1 {
				p1 = i + 1
				break
			}
		}
	}

	if p0 != 0 {
		// key has static prefix, create a static node.
		child := &node{
			static:    true,
			key:       key,
			children:  make([]*node, 256),
			pchildren: []*node{},
			pindex:    n.pindex,
			pnames:    n.pnames,
		}
		n.children[key[0]] = child
		if p0 > 0 {
			child.key = key[:p0]
			n = child
		} else {
			child.data = data
			return child.pindex + 1
		}
	}

	child := &node{
		key:       key[p0:p1],
		children:  make([]*node, 256),
		pchildren: []*node{},
		pindex:    n.pindex,
		pnames:    n.pnames,
	}

	pname := key[p0+1 : p1]
	pnames := make([]string, len(n.pnames)+1)
	copy(pnames, n.pnames)
	pnames[len(n.pnames)] = pname
	child.pnames = pnames
	child.pindex = len(pnames) - 1
	n.pchildren = append(n.pchildren, child)

	if p1 == len(key) {
		child.data = data
		return child.pindex + 1
	}

	return child.addChild(key[p1+1:], data)
}

func (n *node) print(level int) string {
	r := fmt.Sprintf("%v{key: %v, data: %v, pindex: %v, pnames: %v}\n", strings.Repeat(" ", level<<2), n.key, n.data, n.pindex, n.pnames)
	for _, child := range n.children {
		if child != nil {
			r += child.print(level + 1)
		}
	}
	for _, child := range n.pchildren {
		r += child.print(level + 1)
	}
	return r
}
