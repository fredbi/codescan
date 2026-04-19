// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package codescan

import (
	"testing"

	"github.com/go-openapi/testify/v2/require"
)

// Bucket-B error-path tests. Each subfixture under fixtures/enhancements/
// malformed/ carries exactly one annotation that the scanner cannot
// reconcile, so Run() must return a non-nil error. No goldens are
// produced.

func TestMalformed_DefaultInt(t *testing.T) {
	_, err := Run(&Options{
		Packages: []string{"./enhancements/malformed/default-int/..."},
		WorkDir:  "fixtures",
	})
	require.Error(t, err)
}

func TestMalformed_ExampleInt(t *testing.T) {
	_, err := Run(&Options{
		Packages: []string{"./enhancements/malformed/example-int/..."},
		WorkDir:  "fixtures",
	})
	require.Error(t, err)
}

func TestMalformed_MetaBadExtensionKey(t *testing.T) {
	_, err := Run(&Options{
		Packages: []string{"./enhancements/malformed/meta-bad-ext-key/..."},
		WorkDir:  "fixtures",
	})
	require.Error(t, err)
}

func TestMalformed_InfoBadExtensionKey(t *testing.T) {
	_, err := Run(&Options{
		Packages: []string{"./enhancements/malformed/info-bad-ext-key/..."},
		WorkDir:  "fixtures",
	})
	require.Error(t, err)
}

func TestMalformed_BadContact(t *testing.T) {
	_, err := Run(&Options{
		Packages: []string{"./enhancements/malformed/bad-contact/..."},
		WorkDir:  "fixtures",
	})
	require.Error(t, err)
}

func TestMalformed_DuplicateBodyTag(t *testing.T) {
	_, err := Run(&Options{
		Packages: []string{"./enhancements/malformed/duplicate-body-tag/..."},
		WorkDir:  "fixtures",
	})
	require.Error(t, err)
}

func TestMalformed_BadResponseTag(t *testing.T) {
	_, err := Run(&Options{
		Packages: []string{"./enhancements/malformed/bad-response-tag/..."},
		WorkDir:  "fixtures",
	})
	require.Error(t, err)
}

func TestMalformed_BadSecurityDefinitions(t *testing.T) {
	_, err := Run(&Options{
		Packages: []string{"./enhancements/malformed/bad-sec-defs/..."},
		WorkDir:  "fixtures",
	})
	require.Error(t, err)
}
