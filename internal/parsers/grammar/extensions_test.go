// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package grammar

import (
	"testing"
)

// P2.2: `extensions:` and `infoExtensions:` block bodies are parsed
// into Extension{Name, Value, Pos} entries on the Block.

func TestExtensionsBasic(t *testing.T) {
	src := `package p

// swagger:meta
//
// extensions:
//   x-foo: bar
//   x-baz: 42
type Root struct{}
`
	cg, fset := parseCommentGroup(t, src)
	b := Parse(cg, fset)

	var exts []Extension
	for e := range b.Extensions() {
		exts = append(exts, e)
	}
	if len(exts) != 2 {
		t.Fatalf("want 2 extensions, got %d: %+v", len(exts), exts)
	}
	if exts[0].Name != "x-foo" || exts[0].Value != "bar" {
		t.Errorf("ext 0: got %+v", exts[0])
	}
	if exts[1].Name != "x-baz" || exts[1].Value != "42" {
		t.Errorf("ext 1: got %+v", exts[1])
	}
}

func TestExtensionsInfoBlock(t *testing.T) {
	// infoExtensions: uses the same collector path.
	src := `package p

// swagger:meta
//
// infoExtensions:
//   x-logo: https://example.com/logo.png
type Root struct{}
`
	cg, fset := parseCommentGroup(t, src)
	b := Parse(cg, fset)

	count := 0
	for e := range b.Extensions() {
		count++
		if e.Name != "x-logo" {
			t.Errorf("name: got %q", e.Name)
		}
		if e.Value != "https://example.com/logo.png" {
			t.Errorf("value: got %q", e.Value)
		}
	}
	if count != 1 {
		t.Errorf("want 1 extension, got %d", count)
	}
}

func TestExtensionsPositionsPerLine(t *testing.T) {
	// Each extension's Pos points to its own source line.
	src := `package p

// swagger:meta
//
// extensions:
//   x-first: one
//   x-second: two
type Root struct{}
`
	cg, fset := parseCommentGroup(t, src)
	b := Parse(cg, fset)

	var lines []int
	for e := range b.Extensions() {
		lines = append(lines, e.Pos.Line)
	}
	if len(lines) != 2 {
		t.Fatalf("want 2 extensions, got %d", len(lines))
	}
	if lines[1] <= lines[0] {
		t.Errorf("line positions must be monotonic: %v", lines)
	}
}

func TestExtensionsSurvivesAndBodyPreserved(t *testing.T) {
	// Extensions are extracted to Block.Extensions(), but the raw
	// Property.Body is still populated for any analyzer that wants it.
	src := `package p

// swagger:meta
//
// extensions:
//   x-foo: bar
type Root struct{}
`
	cg, fset := parseCommentGroup(t, src)
	b := Parse(cg, fset)

	var prop Property
	for p := range b.Properties() {
		prop = p
	}
	if prop.Keyword.Name != "extensions" {
		t.Fatalf("keyword: got %q", prop.Keyword.Name)
	}
	if len(prop.Body) != 1 || prop.Body[0] != "x-foo: bar" {
		t.Errorf("Body: got %q want [x-foo: bar]", prop.Body)
	}

	extCount := 0
	for range b.Extensions() {
		extCount++
	}
	if extCount != 1 {
		t.Errorf("want 1 extension in parallel, got %d", extCount)
	}
}

func TestExtensionsOnlyForExtensionsKeyword(t *testing.T) {
	// consumes: is a block-head keyword but NOT an extensions block —
	// its body lines must not be scraped into Extensions.
	src := `package p

// swagger:meta
//
// consumes:
//   application/json
type Root struct{}
`
	cg, fset := parseCommentGroup(t, src)
	b := Parse(cg, fset)

	for e := range b.Extensions() {
		t.Errorf("consumes: body must not produce Extensions, got %+v", e)
	}
}

func TestExtensionsMalformedLineIgnored(t *testing.T) {
	// A body line without a ':' can't form an extension; collected
	// into Body but not emitted as an Extension.
	src := `package p

// swagger:meta
//
// extensions:
//   x-good: one
//   not an extension line
//   x-also-good: two
type Root struct{}
`
	cg, fset := parseCommentGroup(t, src)
	b := Parse(cg, fset)

	var names []string
	for e := range b.Extensions() {
		names = append(names, e.Name)
	}
	if len(names) != 2 {
		t.Fatalf("want 2 extensions, got %d: %v", len(names), names)
	}
	if names[0] != "x-good" || names[1] != "x-also-good" {
		t.Errorf("extension names: got %v", names)
	}
}
