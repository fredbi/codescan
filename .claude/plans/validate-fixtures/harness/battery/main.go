package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
)

func unescape(t string) string { return strings.NewReplacer("~1", "/", "~0", "~").Replace(t) }

// resolves reports whether the pointer addresses a node the document actually contains.
func resolves(doc any, ptr string) bool {
	if ptr == "" {
		return true
	}
	node := doc
	for _, t := range strings.Split(strings.TrimPrefix(ptr, "/"), "/") {
		t = unescape(t)
		switch n := node.(type) {
		case map[string]any:
			v, ok := n[t]
			if !ok {
				return false
			}
			node = v
		case []any:
			i, err := strconv.Atoi(t)
			if err != nil || i < 0 || i >= len(n) {
				return false
			}
			node = n[i]
		default:
			return false
		}
	}
	return true
}

func main() {
	files, _ := filepath.Glob(filepath.Join(os.Args[1], "*.json"))
	sort.Strings(files)
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var doc any
		_ = json.Unmarshal(raw, &doc)
		name := strings.TrimSuffix(filepath.Base(f), ".json")

		an, err := loads.Analyzed(json.RawMessage(raw), "")
		if err != nil {
			fmt.Printf("%-26s LOAD ERROR: %v\n", name, err)
			continue
		}
		res, _ := validate.NewSpecValidator(an.Schema(), strfmt.Default).Validate(an)

		type item struct{ sev, ptr, msg string }
		var items []item
		for _, l := range res.LocatedErrors() {
			items = append(items, item{"E", l.Pointer, l.Err.Error()})
		}
		for _, l := range res.LocatedWarnings() {
			items = append(items, item{"W", l.Pointer, l.Err.Error()})
		}
		if len(items) == 0 {
			fmt.Printf("%-26s (no findings)\n", name)
			continue
		}
		for _, it := range items {
			mark := "ok "
			if !resolves(doc, it.ptr) {
				mark = "!! "
			}
			shown := it.ptr
			if shown == "" {
				shown = "«root»"
			}
			fmt.Printf("%-26s %s%s %-42s %s\n", name, mark, it.sev, shown, trunc(it.msg, 74))
		}
	}
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
