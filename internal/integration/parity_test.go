// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package integration_test

import (
	"encoding/json"
	"testing"

	"github.com/go-openapi/codescan"
	"github.com/go-openapi/codescan/internal/scantest"
	"github.com/go-openapi/testify/v2/require"

	oaispec "github.com/go-openapi/spec"
)

// TestParity is the migration safety net for P5 builder migrations:
// every fixture in parityFixtures is scanned twice — once with the
// legacy regex-based pipeline, once with the v2 grammar-parser +
// bridge-taggers pipeline — and the resulting `*spec.Swagger`
// values are JSON-compared.
//
// Design rationale: spec-level compare measures the user-observable
// contract. See `.claude/plans/p5-builder-migrations.md` §5 for the
// full design discussion (why this over a view-level v1 adapter).
//
// Lifetime: **tactical**, not permanent. At P6 cutover the
// `Options.UseGrammarParser` flag is removed and this file is
// deleted in the same commit (P6.4). With only one path left, the
// test has no dual runs to diff — it becomes pure CI burden.
//
// Adding fixtures: append a new entry to parityFixtures below.
// Ordering doesn't matter — each case is independent and t.Run
// parallelises them.
func TestParity(t *testing.T) {
	for _, tc := range parityFixtures {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			docV1 := runFixture(t, tc.Opts, false)
			docV2 := runFixture(t, tc.Opts, true)
			assertSpecsEqual(t, docV1, docV2)
		})
	}
}

// parityFixture names one fixture scan that must produce the same
// spec under both pipeline values. Opts is cloned in runFixture
// before UseGrammarParser is set, so the template stays immutable.
type parityFixture struct {
	Name string
	Opts codescan.Options
}

//nolint:gochecknoglobals // migration-scoped test table; removed at P6 cutover alongside the flag.
var parityFixtures = []parityFixture{
	// fixtures/enhancements/ — each entry mirrors a TestCoverage_* test in
	// coverage_enhancements_test.go. Entries that intentionally exercise
	// error paths (e.g. UnknownAnnotation, malformed/*) are NOT included —
	// they don't produce comparable spec output.
	{"EmbeddedTypes", codescan.Options{Packages: pkgs("./enhancements/embedded-types/..."), ScanModels: true}},
	{"AllOfEdges", codescan.Options{Packages: pkgs("./enhancements/allof-edges/..."), ScanModels: true}},
	{"StrfmtArrays", codescan.Options{Packages: pkgs("./enhancements/strfmt-arrays/..."), ScanModels: true}},
	{"DefaultsExamples", codescan.Options{Packages: pkgs("./enhancements/defaults-examples/..."), ScanModels: true}},
	{"InterfaceMethods", codescan.Options{Packages: pkgs("./enhancements/interface-methods/..."), ScanModels: true}},
	{"InterfaceMethodsXNullable", codescan.Options{Packages: pkgs("./enhancements/interface-methods/..."), ScanModels: true, SetXNullableForPointers: true}},
	{"AliasExpand", codescan.Options{Packages: pkgs("./enhancements/alias-expand/..."), ScanModels: true}},
	{"AliasRef", codescan.Options{Packages: pkgs("./enhancements/alias-expand/..."), ScanModels: true, RefAliases: true}},
	{"AliasResponseRef", codescan.Options{Packages: pkgs("./enhancements/alias-response/..."), ScanModels: true, RefAliases: true}},
	{"ResponseEdges", codescan.Options{Packages: pkgs("./enhancements/response-edges/..."), ScanModels: true}},
	{"NamedBasic", codescan.Options{Packages: pkgs("./enhancements/named-basic/..."), ScanModels: true}},
	{"SwaggerTypeArray", codescan.Options{Packages: pkgs("./enhancements/swagger-type-array/..."), ScanModels: true}},
	{"RefAliasChain", codescan.Options{Packages: pkgs("./enhancements/ref-alias-chain/..."), ScanModels: true, RefAliases: true}},
	{"EnumDocs", codescan.Options{Packages: pkgs("./enhancements/enum-docs/..."), ScanModels: true}},
	{"EnumOverrides", codescan.Options{Packages: pkgs("./enhancements/enum-overrides/..."), ScanModels: true}},
	{"TextMarshal", codescan.Options{Packages: pkgs("./enhancements/text-marshal/..."), ScanModels: true}},
	{"AllHTTPMethods", codescan.Options{Packages: pkgs("./enhancements/all-http-methods/...")}},
	{"NamedStructTags", codescan.Options{Packages: pkgs("./enhancements/named-struct-tags/..."), ScanModels: true}},
	{"NamedStructTagsRef", codescan.Options{Packages: pkgs("./enhancements/named-struct-tags-ref/..."), ScanModels: true}},
	{"TopLevelKinds", codescan.Options{Packages: pkgs("./enhancements/top-level-kinds/..."), ScanModels: true}},
	// fixtures/goparsing/
	{"Petstore", codescan.Options{Packages: pkgs("./goparsing/petstore/...")}},
	{"Bookings", codescan.Options{Packages: pkgs("./goparsing/bookings/..."), ScanModels: true}},
	// Exercises swagger:operation (YAML-bodied operation spec).
	{"ClassificationOpAnnotation", codescan.Options{Packages: pkgs("./goparsing/classification/operations_annotation/...")}},
	// Exercises swagger:route with rich bodies: Consumes / Produces / Schemes /
	// Security / Parameters / Responses / Extensions.
	{"ClassificationRoutes", codescan.Options{Packages: pkgs("./goparsing/classification/...")}},
}

// pkgs is a tiny alias for []string — it makes the fixture table
// readable at a glance (the Packages field dominates the line
// otherwise).
func pkgs(p ...string) []string { return p }

// runFixture scans tc.Opts with UseGrammarParser=useGrammar and
// returns the resulting spec. The template Options is cloned so
// the table stays unmodified; WorkDir is injected once here rather
// than duplicated in every table row.
func runFixture(t *testing.T, template codescan.Options, useGrammar bool) *oaispec.Swagger {
	t.Helper()
	opts := template // value copy
	opts.WorkDir = scantest.FixturesDir()
	opts.UseGrammarParser = useGrammar
	doc, err := codescan.Run(&opts)
	require.NoError(t, err)
	require.NotNil(t, doc)
	return doc
}

// assertSpecsEqual marshals both specs to indented JSON and
// diffs as strings. This is stricter than reflect.DeepEqual (it
// catches field-order differences in slices) and produces a
// human-readable diff path on failure.
func assertSpecsEqual(t *testing.T, v1, v2 *oaispec.Swagger) {
	t.Helper()
	v1JSON, err := json.MarshalIndent(v1, "", "  ")
	require.NoError(t, err)
	v2JSON, err := json.MarshalIndent(v2, "", "  ")
	require.NoError(t, err)
	if string(v1JSON) != string(v2JSON) {
		t.Fatalf(
			"parity mismatch — v1 (legacy) vs v2 (grammar) differ:\n"+
				"--- v1 (%d bytes) ---\n%s\n"+
				"--- v2 (%d bytes) ---\n%s\n",
			len(v1JSON), v1JSON, len(v2JSON), v2JSON,
		)
	}
}
