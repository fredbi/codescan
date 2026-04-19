// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package codescan

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

// goldenDir is relative to the test CWD (repo root for this monolithic package).
const goldenDir = "fixtures/integration/golden"

// compareOrDumpJSON marshals got to stable JSON and either writes it to
// fixtures/integration/golden/<name> (when UPDATE_GOLDEN=1) or asserts it
// JSON-equals the stored golden.
func compareOrDumpJSON(t *testing.T, got any, name string) {
	t.Helper()

	data, err := json.MarshalIndent(got, "", "  ")
	require.NoError(t, err)

	path := filepath.Join(goldenDir, name)

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, data, 0o644))
		t.Logf("wrote golden %s", name)
		return
	}

	want, err := os.ReadFile(path)
	require.NoError(t, err, "missing golden %s — run with UPDATE_GOLDEN=1 to create", name)
	assert.JSONEqT(t, string(want), string(data))
}
