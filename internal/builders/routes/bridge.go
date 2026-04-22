// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package routes

import (
	"strings"

	"github.com/go-openapi/codescan/internal/parsers"
	"github.com/go-openapi/codescan/internal/parsers/grammar"
	oaispec "github.com/go-openapi/spec"
)

// applyBlockToRoute is the grammar-path counterpart of Builder.Build's
// SectionedParser invocation. Parses route.Remaining, extracts
// summary/description, and dispatches each level-0 Property to the
// appropriate setter. Body parsing for the multi-line keywords
// (consumes, produces, security, parameters, responses, extensions)
// delegates to the existing v1 parser instances — each already
// handles its specific body shape (YAML lists, name:value mappings,
// nested extension bodies). The bridge contributes line-splitting /
// title-description / dispatch; the heavy lifting stays in the
// established parser code.
func (r *Builder) applyBlockToRoute(op *oaispec.Operation) error {
	block := grammar.NewParser(r.ctx.FileSet()).Parse(r.route.Remaining)

	title, desc := parsers.CollectScannerTitleDescription(block.ProseLines())
	op.Summary = parsers.JoinDropLast(title)
	op.Description = parsers.JoinDropLast(desc)

	for prop := range block.Properties() {
		if prop.ItemsDepth != 0 {
			continue
		}
		if err := r.dispatchRouteKeyword(prop, op); err != nil {
			return err
		}
	}
	return nil
}

// Keyword names reused from grammar's keyword table — kept as
// constants to avoid magic strings in the dispatch table.
const (
	kwSchemes    = "schemes"
	kwDeprecated = "deprecated"
	kwConsumes   = "consumes"
	kwProduces   = "produces"
	kwSecurity   = "security"
	kwParameters = "parameters"
	kwResponses  = "responses"
	kwExtensions = "extensions"
)

// dispatchRouteKeyword routes one grammar Property to the legacy
// body-parser that already knows how to parse that keyword's body
// shape. The body-parsers' Parse(lines []string) signature accepts
// grammar's Property.Body directly — comment markers are already
// stripped, YAML list markers survive, etc.
func (r *Builder) dispatchRouteKeyword(p grammar.Property, op *oaispec.Operation) error {
	switch p.Keyword.Name {
	case kwSchemes:
		r.applyRouteSchemes(p, op)
	case kwDeprecated:
		if p.Typed.Type == grammar.ValueBoolean {
			op.Deprecated = p.Typed.Boolean
		}
	case kwConsumes:
		return parsers.NewConsumesDropEmptyParser(opConsumesSetter(op)).Parse(p.Body)
	case kwProduces:
		return parsers.NewProducesDropEmptyParser(opProducesSetter(op)).Parse(p.Body)
	case kwSecurity:
		return parsers.NewSetSecurityScheme(opSecurityDefsSetter(op)).Parse(p.Body)
	case kwParameters:
		return parsers.NewSetParams(r.parameters, opParamSetter(op)).Parse(p.Body)
	case kwResponses:
		return parsers.NewSetResponses(r.definitions, r.responses, opResponsesSetter(op)).Parse(p.Body)
	case kwExtensions:
		return parsers.NewSetExtensions(opExtensionsSetter(op), r.ctx.Debug()).Parse(p.Body)
	}
	return nil
}

// applyRouteSchemes parses `schemes: http, https, ws, wss` — v1 uses
// a regex capture that isolates the post-colon comma-list; the
// grammar already hands us the trimmed value directly.
func (r *Builder) applyRouteSchemes(p grammar.Property, op *oaispec.Operation) {
	schemes := make([]string, 0)
	for s := range strings.SplitSeq(p.Value, ",") {
		if ts := strings.TrimSpace(s); ts != "" {
			schemes = append(schemes, ts)
		}
	}
	if len(schemes) > 0 {
		op.Schemes = schemes
	}
}
