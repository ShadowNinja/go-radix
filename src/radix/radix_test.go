package radix

import (
	"fmt"
	"testing"
)

func TestRadixContains(t *testing.T) {
	m := New()
	m.Insert("foo")
	m.Insert("foobar")
	m.Insert("fool")
	m.Insert("fountain")

	fmt.Println(m)

	if m.Contains("fo", false) {
		t.Error("fo is in tree")
	}
	if m.Contains("fount", false) {
		t.Error("fount is in tree")
	}
	if m.Contains("font", false) {
		t.Error("font is in tree")
	}
	if !m.Contains("foo", false) {
		t.Error("foo is not in tree")
	}
	if !m.Contains("fool", false) {
		t.Error("fool is not in tree")
	}

	// With prefix searching
	if !m.Contains("fo", true) {
		t.Error("fo prefix is not in tree")
	}
	if !m.Contains("foun", true) {
		t.Error("foun prefix is not in tree")
	}
	if m.Contains("font", true) {
		t.Error("font prefix is in tree")
	}
}

func TestRadixRemove(t *testing.T) {
	m := New()
	m.Insert("foobar")
	m.Insert("fool")

	if m.Remove("font", false) {
		t.Error("font was removed")
	}
	if m.Remove("foo", false) {
		t.Error("fo was removed (not final)")
	}
	if !m.Remove("fool", false) {
		t.Error("fool wasn't removed")
	}
	if m.Contains("fool", false) {
		t.Error("fool is in tree")
	}
	if !m.Remove("fo", true) {
		t.Error("fo prefix was not removed")
	}
	if m.Contains("foobar", false) {
		t.Error("foobar is in tree")
	}
}

func TestRadixMerge(t *testing.T) {
	m := New()
	m.Insert("ab")
	m.Insert("ac")
	m.Remove("ac", false)
	if m.root.children[0].label != "ab" {
		t.Error("'a' and 'b' nodes weren't merged")
	}
}
