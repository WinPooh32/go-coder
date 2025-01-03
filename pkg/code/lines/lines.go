package lines

import (
	"fmt"
	"strings"
)

// AddNumbers adds number to the every line of the s.
// Format: "<line_number>:<line_content>".
// Example: "1:package main\n".
func AddNumbers(s string) string {
	if s == "" {
		return ""
	}

	ss := strings.Split(s, "\n")

	for i, s := range ss {
		ss[i] = fmt.Sprintf("%d:%s\n", i+1, s)
	}

	return strings.Join(ss, "")
}
