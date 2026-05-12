package server

import "strings"

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

// extractID pulls the first path segment after the given prefix.
func extractID(path, prefix string) string {
	s := strings.TrimPrefix(path, prefix)
	idx := strings.Index(s, "/")
	if idx == -1 {
		return s
	}
	return s[:idx]
}
