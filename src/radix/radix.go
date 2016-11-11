package radix

import (
	"fmt"
	"io"
)

// Tree is the type of the radix tree container.
type Tree struct {
	root subtree
}

// New createa a empty radix tree and returns it.
func New() *Tree {
	m := new(Tree)
	return m
}

// Insert inserts a string into the tree.
// It returns whether the tree was modified.
func (m *Tree) Insert(str string) bool {
	if len(str) == 0 {
		return false
	}

	curTree := &m.root

	for {
		found := false
		children := curTree.children

		for i := range children {
			e := &children[i]

			shared := sharedPrefix(str, e.label)

			// If the search string is equal to the shared string,
			// the string already exists in the tree.  We just have
			// to make sure it's final.
			if str == shared {
				e.isFinal = true
				return false
				// If the child's label is equal to the shared string, you have to
				// traverse another level down the tree, so substring the search
				// string, break the loop, and run it back
			} else if shared == e.label {
				curTree = &e.st
				str = str[len(shared):]
				found = true
				break
				// If the child's label and the search string share a partial prefix,
				// then both the label and the search string need to be substringed
				// and a new branch needs to be created
			} else if len(shared) > 0 {
				// Substring both the search string and the label from the shared prefix
				str = str[len(shared):]
				labelPostfix := e.label[len(shared):]

				e.label = shared

				// Create edge to hold all the values the values under the current edge
				newE := newEdge(labelPostfix, e.isFinal)

				// Edge is no longer final since the end is being moved to newE
				e.isFinal = false

				// Move all of the children of the current edge under the new edge
				newE.st.children = e.st.children

				// Resed the edge to have two children: the new child and an edge with the newly inserted string
				e.st.children = []edge{newE, newEdge(str, true)}

				return true
			}
		}

		// If none of the children share a prefix, you have to create a new child
		if !found {
			curTree.AddChild(str)
			return true
		}
	}
}

// Contains checks if the tree has a match for the string being searched for.
// If prefix is true, this will return true if any string with the passed
// string as a prefix exists in the tree.  Otherwise, if prefix is false,
// this requires an exact match for the string to be in the tree.
func (m *Tree) Contains(str string, prefix bool) bool {
	if len(str) == 0 {
		return true
	} else if len(m.root.children) == 0 {
		return false
	}

	curTree := &m.root
	for {
		found := false
		children := curTree.children

		for i := range children {
			e := &children[i]
			if str == e.label {
				return prefix || e.isFinal
			}

			shared := sharedPrefix(str, e.label)

			if len(shared) == 0 {
				continue
				// If the string is longer than the label, move into the label's subtree
			} else if shared == e.label {
				curTree = &e.st
				str = str[len(shared):]
				found = true
				break
			} else if prefix && shared == str {
				return true
			}
			return false
		}

		if !found {
			return false
		}
	}
}

// Remove removes a string from the tree.
// If prefix is true, the string is interpreted as a prefix, and the
// string and all strings with that prefix will be removed. Otherwise,
// if prefix is false, only the specific string passed will be removed.
func (m *Tree) Remove(str string, prefix bool) bool {
	if len(str) == 0 {
		if prefix {
			m.root = newSubtree()
		}
		return prefix
	} else if len(m.root.children) == 0 {
		return false
	}

	curTree := &m.root
	var curEdge *edge
	for {
		found := false
		children := curTree.children

		for i := range children {
			e := &children[i]
			if str == e.label {
				return m.remove(curEdge, curTree, e, i, prefix)
			}

			shared := sharedPrefix(str, e.label)

			// If the string is longer than the label, move into the label's subtree
			if shared == e.label {
				curEdge = e
				curTree = &e.st
				str = str[len(shared):]
				found = true
				break
				// If the label is longer than the string and we're matching of prefixes
			} else if prefix && shared == str {
				return m.remove(curEdge, curTree, e, i, prefix)
			}
		}

		if !found {
			return false
		}
	}
}

// This function does the actual removal once you've found an edge to remove.
func (m *Tree) remove(parent *edge, parentTree *subtree, e *edge, i int, prefix bool) bool {
	if !prefix && !e.isFinal {
		return false
	}
	e.isFinal = false
	if prefix || len(e.st.children) == 0 {
		children := parentTree.children
		// Move entries after the current edge back one, overwriting the current edge
		copy(children[i:], children[i+1:])
		// ...and remove the last (now duplicate) element
		parentTree.children = children[:len(children)-1]
		// Merge edges.  That is, if we have a tree like this:
		// Root [x]
		// `- a [ ]
		//    `- b [x]
		//    `- c [x]
		// If we remove "ac" then the tree can be compacted by
		// merging the "a" and "b" edges into an "ab" edge like so:
		// Root [x]
		// `- ab [x]
		// parent points to the "a" edge
		if parent != nil && len(parentTree.children) == 1 && !parent.isFinal {
			// This is the "b" edge from the diagram above.  Since
			// there's exactly one element in the slice we know we
			// can use the index 0.
			mergeEdge := parentTree.children[0]

			// Move children up
			parentTree.children = mergeEdge.st.children

			// Set the label to the concatenation of the two labels.
			parent.label += mergeEdge.label

			// Transfer finality
			parent.isFinal = mergeEdge.isFinal
		}
	}
	return true
}

// Format formats the Tree for display
func (m *Tree) Format(f fmt.State, c rune) {
	// Print the Root
	f.Write([]byte("Root\n"))
	// Print the children
	for i := range m.root.children {
		m.root.children[i].formatEdge(f, 1)
	}
}

type subtree struct {
	children []edge
}

func newSubtree() subtree {
	return subtree{
		children: []edge(nil),
	}
}

func (t *subtree) AddChild(label string) {
	t.children = append(t.children, newEdge(label, true))
}

type edge struct {
	// The string contents of this node
	label string

	// Final nodes: a final node is the last node in a series that makes up a
	// string in the tree.  For example, for a tree with the strings "foobar"
	// and "fooqux" you'll have: ([x] indicates final)
	// Root [x]
	//  `- foo [ ]
	//     |- bar [x]
	//     `- qux [x]
	// Now, if you add "foo" to the tree, the foo node will simply become final.
	// This allows you to tell if foo in the tree or just a node that had to be
	// split to add other elements
	isFinal bool

	st subtree
}

func newEdge(label string, isFinal bool) edge {
	return edge{
		label:   label,
		isFinal: isFinal,
		st:      newSubtree(),
	}
}

// Prints the tree for debugging/visualization purposes.
func (e *edge) formatEdge(w io.Writer, level int) {
	var err error
	// Tab over once up to the edge's level
	for i := 1; i <= level; i++ {
		if i == level {
			_, err = w.Write([]byte("`- "))
		} else {
			_, err = w.Write([]byte("   "))
		}
		check(err)
	}
	_, err = w.Write([]byte(e.label))
	check(err)
	if e.isFinal {
		_, err = w.Write([]byte(" [x]\n"))
	} else {
		_, err = w.Write([]byte(" [ ]\n"))
	}
	check(err)

	// Print the children
	for i := range e.st.children {
		e.st.children[i].formatEdge(w, level+1)
	}
}

// Returns the prefix that is shared between the two input strings
// i.e. sharedPrefix("court", "coral") -> "co"
func sharedPrefix(str1, str2 string) string {
	l := min(len(str1), len(str2))
	temp := make([]byte, 0, l)
	for i := 0; i < l; i++ {
		if str1[i] != str2[i] {
			break
		}
		temp = append(temp, str1[i])
	}
	return string(temp)
}

func min(i1, i2 int) int {
	if i1 < i2 {
		return i1
	}
	return i2
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
