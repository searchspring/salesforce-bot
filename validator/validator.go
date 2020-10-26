package validator

import "strings"

// FindBlankVals takes a map[string]string and returns a list of all keys whose values are blank
func FindBlankVals(m map[string]string) []string {
	var blanks []string
	for k, v := range m {
		if strings.TrimSpace(v) == "" {
			blanks = append(blanks, k)
		}
	}
	return blanks
}
