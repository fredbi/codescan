// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package scanner

import "github.com/go-openapi/spec"

type Options struct {
	Packages                []string
	InputSpec               *spec.Swagger
	ScanModels              bool
	WorkDir                 string
	BuildTags               string
	ExcludeDeps             bool
	Include                 []string
	Exclude                 []string
	IncludeTags             []string
	ExcludeTags             []string
	SetXNullableForPointers bool
	RefAliases              bool // aliases result in $ref, otherwise aliases are expanded
	TransparentAliases      bool // aliases are completely transparent, never creating definitions
	DescWithRef             bool // allow overloaded descriptions together with $ref, otherwise jsonschema draft4 $ref predates everything
	SkipExtensions          bool // skip generating x-go-* vendor extensions in the spec
	Debug                   bool // enable verbose debug logging during scanning

	// UseGrammarParser routes comment-group parsing through the v2
	// hand-rolled grammar parser at internal/parsers/grammar/ (plus
	// bridge-taggers that call the existing ValidationBuilder /
	// SwaggerTypable / … interfaces) instead of the legacy
	// regex-based taggers.
	//
	// Default false: the legacy path runs unchanged. The flag is
	// the dual-path coexistence seam used by the parity harness
	// during the P5 migration (one run per value of the flag, outputs
	// diffed). At P6 cutover the flag is removed and grammar-parser
	// becomes the only path. See .claude/plans/p5-builder-migrations.md
	// and grammar-parser-tasks.md P4.3 / P5 cross-cutting.
	UseGrammarParser bool
}
