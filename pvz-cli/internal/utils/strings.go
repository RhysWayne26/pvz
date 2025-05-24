package utils

import "strings"

func UniqueStrings(input []string) []string {
	seen := make(map[string]struct{}, len(input))
	var result []string
	for _, s := range input {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, dup := seen[s]; dup {
			continue
		}
		seen[s] = struct{}{}
		result = append(result, s)
	}
	return result
}
