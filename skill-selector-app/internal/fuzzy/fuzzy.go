// Package fuzzy provides a case-insensitive subsequence match used to filter
// lists as the user types. It is a pure utility with no UI or app knowledge.
package fuzzy

import (
	"strings"
	"unicode"
)

// Match reports whether every rune of query appears in candidate in order,
// ignoring case. Spaces in the query are ignored so "code rev" matches
// "code-review". An empty query matches everything.
func Match(query, candidate string) bool {
	q := []rune(strings.ToLower(query))
	c := []rune(strings.ToLower(candidate))
	qi := 0
	for qi < len(q) && unicode.IsSpace(q[qi]) {
		qi++
	}
	for _, r := range c {
		if qi >= len(q) {
			return true
		}
		if r == q[qi] {
			qi++
			for qi < len(q) && unicode.IsSpace(q[qi]) {
				qi++
			}
		}
	}
	return qi >= len(q)
}

// Predicate returns a function that reports whether a candidate matches
// query, for use with Filter.
func Predicate(query string) func(string) bool {
	return func(candidate string) bool { return Match(query, candidate) }
}

// Filter returns the items for which keep(item) is true, preserving order.
// The result is always a new slice.
func Filter[T any](items []T, keep func(T) bool) []T {
	out := make([]T, 0, len(items))
	for _, it := range items {
		if keep(it) {
			out = append(out, it)
		}
	}
	return out
}
