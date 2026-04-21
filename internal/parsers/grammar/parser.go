// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

// Package grammar implements the v2 annotation parser.
//
// It replaces the regexp-based parser at internal/parsers/*.go with a
// hand-rolled recursive-descent parser producing a typed Block family
// (see ast.go).
//
// See .claude/plans/grammar-parser-architecture.md for the "why" and
// .claude/plans/grammar-parser-tasks.md for the "how".
package grammar

// TODO: P1 — recursive-descent envelope parser + Parser interface.
