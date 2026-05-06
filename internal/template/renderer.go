package template

import (
	"fmt"
	"regexp"
	"strings"
)

var varPattern = regexp.MustCompile(`\{\{(\s*\w+\s*)\}\}`)

// Render replaces {{variable}} placeholders in src with values from data.
// Returns an error listing any variables present in src but missing from data.
func Render(src string, data map[string]any) (string, error) {
	missing := missingVars(src, data)
	if len(missing) > 0 {
		return "", fmt.Errorf("missing required variables: %s", strings.Join(missing, ", "))
	}
	return varPattern.ReplaceAllStringFunc(src, func(match string) string {
		key := strings.TrimSpace(match[2 : len(match)-2])
		return fmt.Sprintf("%v", data[key])
	}), nil
}

// ExtractVars returns the set of variable names referenced in src.
func ExtractVars(src string) []string {
	seen := map[string]struct{}{}
	var vars []string
	for _, m := range varPattern.FindAllStringSubmatch(src, -1) {
		k := strings.TrimSpace(m[1])
		if _, ok := seen[k]; !ok {
			seen[k] = struct{}{}
			vars = append(vars, k)
		}
	}
	return vars
}

func missingVars(src string, data map[string]any) []string {
	var missing []string
	for _, v := range ExtractVars(src) {
		if _, ok := data[v]; !ok {
			missing = append(missing, v)
		}
	}
	return missing
}
