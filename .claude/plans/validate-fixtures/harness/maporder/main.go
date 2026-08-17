package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
)

// Runs the same document N times and reports how the reported set varies.
func main() {
	raw, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	continueOnErrors := len(os.Args) > 2 && os.Args[2] == "continue"

	counts := map[string]int{}
	const n = 50
	for range n {
		doc, err := loads.Analyzed(json.RawMessage(raw), "")
		if err != nil {
			panic(err)
		}
		v := validate.NewSpecValidator(doc.Schema(), strfmt.Default)
		v.Options.ContinueOnErrors = continueOnErrors
		res, _ := v.Validate(doc)

		var got []string
		for _, l := range res.LocatedErrors() {
			got = append(got, "ERR "+l.Pointer)
		}
		for _, l := range res.LocatedWarnings() {
			got = append(got, "WARN "+l.Pointer)
		}
		sort.Strings(got)
		counts[fmt.Sprint(got)]++
	}

	fmt.Printf("ContinueOnErrors=%v, %d runs:\n", continueOnErrors, n)
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("  %3d runs  %s\n", counts[k], k)
	}
	fmt.Printf("  => %d distinct outcomes\n", len(counts))
}
