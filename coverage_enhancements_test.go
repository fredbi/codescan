// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package codescan

import (
	"testing"

	"github.com/go-openapi/spec"
	"github.com/go-openapi/testify/v2/require"
)

// These tests target partially covered code paths in the baseline scanner.
// Each one scans a dedicated fixture under fixtures/enhancements/ and
// captures the resulting Swagger document as a golden file so that the
// refactored branch can be checked for behavioural drift.

func TestCoverage_EmbeddedTypes(t *testing.T) {
	doc, err := Run(&Options{
		Packages:   []string{"./enhancements/embedded-types/..."},
		WorkDir:    "fixtures",
		ScanModels: true,
	})
	require.NoError(t, err)
	require.NotNil(t, doc)

	compareOrDumpJSON(t, doc, "enhancements_embedded_types.json")
}

func TestCoverage_AllOfEdges(t *testing.T) {
	doc, err := Run(&Options{
		Packages:   []string{"./enhancements/allof-edges/..."},
		WorkDir:    "fixtures",
		ScanModels: true,
	})
	require.NoError(t, err)
	require.NotNil(t, doc)

	compareOrDumpJSON(t, doc, "enhancements_allof_edges.json")
}

func TestCoverage_StrfmtArrays(t *testing.T) {
	doc, err := Run(&Options{
		Packages:   []string{"./enhancements/strfmt-arrays/..."},
		WorkDir:    "fixtures",
		ScanModels: true,
	})
	require.NoError(t, err)
	require.NotNil(t, doc)

	compareOrDumpJSON(t, doc, "enhancements_strfmt_arrays.json")
}

func TestCoverage_DefaultsExamples(t *testing.T) {
	doc, err := Run(&Options{
		Packages:   []string{"./enhancements/defaults-examples/..."},
		WorkDir:    "fixtures",
		ScanModels: true,
	})
	require.NoError(t, err)
	require.NotNil(t, doc)

	compareOrDumpJSON(t, doc, "enhancements_defaults_examples.json")
}

func TestCoverage_InterfaceMethods(t *testing.T) {
	doc, err := Run(&Options{
		Packages:   []string{"./enhancements/interface-methods/..."},
		WorkDir:    "fixtures",
		ScanModels: true,
	})
	require.NoError(t, err)
	require.NotNil(t, doc)

	compareOrDumpJSON(t, doc, "enhancements_interface_methods.json")
}

// TestCoverage_AliasExpand scans the alias-expand fixture with default
// Options so that buildAlias / buildFieldAlias take the non-transparent
// expansion path: each alias resolves to the underlying type and the
// target is emitted inline rather than as a $ref.
func TestCoverage_AliasExpand(t *testing.T) {
	doc, err := Run(&Options{
		Packages:   []string{"./enhancements/alias-expand/..."},
		WorkDir:    "fixtures",
		ScanModels: true,
	})
	require.NoError(t, err)
	require.NotNil(t, doc)

	compareOrDumpJSON(t, doc, "enhancements_alias_expand.json")
}

// TestCoverage_AliasRef scans the alias-expand fixture with RefAliases=true
// so body-parameter and response aliases resolve to $ref via makeRef, and
// the alias-of-alias chain resolves through the non-transparent switch.
func TestCoverage_AliasRef(t *testing.T) {
	doc, err := Run(&Options{
		Packages:   []string{"./enhancements/alias-expand/..."},
		WorkDir:    "fixtures",
		ScanModels: true,
		RefAliases: true,
	})
	require.NoError(t, err)
	require.NotNil(t, doc)

	compareOrDumpJSON(t, doc, "enhancements_alias_ref.json")
}

func TestCoverage_TopLevelKinds(t *testing.T) {
	doc, err := Run(&Options{
		Packages:   []string{"./enhancements/top-level-kinds/..."},
		WorkDir:    "fixtures",
		ScanModels: true,
	})
	require.NoError(t, err)
	require.NotNil(t, doc)

	compareOrDumpJSON(t, doc, "enhancements_top_level_kinds.json")
}

func TestCoverage_NamedStructTags(t *testing.T) {
	doc, err := Run(&Options{
		Packages:   []string{"./enhancements/named-struct-tags/..."},
		WorkDir:    "fixtures",
		ScanModels: true,
	})
	require.NoError(t, err)
	require.NotNil(t, doc)

	compareOrDumpJSON(t, doc, "enhancements_named_struct_tags.json")
}

// TestCoverage_UnknownAnnotation asserts that scanning a file with an
// unknown swagger: annotation returns a classifier error. This exercises
// the default branch of typeIndex.detectNodes.
func TestCoverage_UnknownAnnotation(t *testing.T) {
	_, err := Run(&Options{
		Packages:   []string{"./enhancements/unknown-annotation/..."},
		WorkDir:    "fixtures",
		ScanModels: true,
	})
	require.Error(t, err)
}

func TestCoverage_AllHTTPMethods(t *testing.T) {
	doc, err := Run(&Options{
		Packages: []string{"./enhancements/all-http-methods/..."},
		WorkDir:  "fixtures",
	})
	require.NoError(t, err)
	require.NotNil(t, doc)

	compareOrDumpJSON(t, doc, "enhancements_all_http_methods.json")
}

func TestCoverage_TextMarshal(t *testing.T) {
	doc, err := Run(&Options{
		Packages:   []string{"./enhancements/text-marshal/..."},
		WorkDir:    "fixtures",
		ScanModels: true,
	})
	require.NoError(t, err)
	require.NotNil(t, doc)

	compareOrDumpJSON(t, doc, "enhancements_text_marshal.json")
}

func TestCoverage_EnumDocs(t *testing.T) {
	doc, err := Run(&Options{
		Packages:   []string{"./enhancements/enum-docs/..."},
		WorkDir:    "fixtures",
		ScanModels: true,
	})
	require.NoError(t, err)
	require.NotNil(t, doc)

	compareOrDumpJSON(t, doc, "enhancements_enum_docs.json")
}

func TestCoverage_RefAliasChain(t *testing.T) {
	doc, err := Run(&Options{
		Packages:   []string{"./enhancements/ref-alias-chain/..."},
		WorkDir:    "fixtures",
		ScanModels: true,
		RefAliases: true,
	})
	require.NoError(t, err)
	require.NotNil(t, doc)

	compareOrDumpJSON(t, doc, "enhancements_ref_alias_chain.json")
}

func TestCoverage_NamedBasic(t *testing.T) {
	doc, err := Run(&Options{
		Packages:   []string{"./enhancements/named-basic/..."},
		WorkDir:    "fixtures",
		ScanModels: true,
	})
	require.NoError(t, err)
	require.NotNil(t, doc)

	compareOrDumpJSON(t, doc, "enhancements_named_basic.json")
}

func TestCoverage_ResponseEdges(t *testing.T) {
	doc, err := Run(&Options{
		Packages:   []string{"./enhancements/response-edges/..."},
		WorkDir:    "fixtures",
		ScanModels: true,
	})
	require.NoError(t, err)
	require.NotNil(t, doc)

	compareOrDumpJSON(t, doc, "enhancements_response_edges.json")
}

// TestCoverage_AliasResponseRef scans a fixture where the swagger:response
// annotation is itself on an alias declaration. Under RefAliases=true the
// scanner takes the responseBuilder.buildAlias refAliases switch, which
// is not covered by any other test.
func TestCoverage_AliasResponseRef(t *testing.T) {
	doc, err := Run(&Options{
		Packages:   []string{"./enhancements/alias-response/..."},
		WorkDir:    "fixtures",
		ScanModels: true,
		RefAliases: true,
	})
	require.NoError(t, err)
	require.NotNil(t, doc)

	compareOrDumpJSON(t, doc, "enhancements_alias_response_ref.json")
}

func TestCoverage_InterfaceMethods_XNullable(t *testing.T) {
	doc, err := Run(&Options{
		Packages:                []string{"./enhancements/interface-methods/..."},
		WorkDir:                 "fixtures",
		ScanModels:              true,
		SetXNullableForPointers: true,
	})
	require.NoError(t, err)
	require.NotNil(t, doc)

	compareOrDumpJSON(t, doc, "enhancements_interface_methods_xnullable.json")
}

// TestCoverage_InputOverlay feeds an InputSpec carrying paths with every
// HTTP verb so that collectOperationsFromInput indexes all operations
// before the scanner merges its own discoveries.
func TestCoverage_InputOverlay(t *testing.T) {
	input := &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Swagger: "2.0",
			Info: &spec.Info{
				InfoProps: spec.InfoProps{
					Title:   "Overlay",
					Version: "0.0.1",
				},
			},
			Paths: &spec.Paths{
				Paths: map[string]spec.PathItem{
					"/items": {
						PathItemProps: spec.PathItemProps{
							Get:    &spec.Operation{OperationProps: spec.OperationProps{ID: "listItems"}},
							Post:   &spec.Operation{OperationProps: spec.OperationProps{ID: "createItem"}},
							Put:    &spec.Operation{OperationProps: spec.OperationProps{ID: "replaceItem"}},
							Patch:  &spec.Operation{OperationProps: spec.OperationProps{ID: "patchItem"}},
							Delete: &spec.Operation{OperationProps: spec.OperationProps{ID: "deleteItem"}},
							Head:   &spec.Operation{OperationProps: spec.OperationProps{ID: "checkItem"}},
							Options: &spec.Operation{OperationProps: spec.OperationProps{ID: "optionsItem"}},
						},
					},
				},
			},
		},
	}

	doc, err := Run(&Options{
		Packages:  []string{"./enhancements/embedded-types/..."},
		WorkDir:   "fixtures",
		InputSpec: input,
	})
	require.NoError(t, err)
	require.NotNil(t, doc)

	compareOrDumpJSON(t, doc, "enhancements_input_overlay.json")
}
