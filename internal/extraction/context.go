package extraction

import (
	"regexp"
	"strings"
)

var importLine = regexp.MustCompile(`^(?:import ([A-Za-z_][\w.]*)|from [A-Za-z_][\w.]* import ([A-Za-z_]\w*))(?: as ([A-Za-z_]\w*))?$`)

// Only documented, unambiguous single imports are reused. We do not invent
// imports, install commands, file contents or missing variable definitions.
func ImportContext(blocks []string, code string) string {
	// A module mentioned only in an explanatory comment is not a dependency.
	usage := []string{}
	for _, line := range strings.Split(code, "\n") {
		usage = append(usage, strings.SplitN(line, "#", 2)[0])
	}
	usedCode := strings.Join(usage, "\n")
	order := []string{}
	byName := map[string]string{}
	conflicts := map[string]bool{}
	seen := map[string]bool{}
	for _, block := range blocks {
		for _, line := range strings.Split(block, "\n") {
			if command, ok := pythonCommand(line); ok {
				line = command
			}
			m := importLine.FindStringSubmatch(line)
			if m == nil || strings.Contains(code, line) || seen[line] {
				continue
			}
			name := m[3]
			if name == "" {
				name = m[2]
			}
			if name == "" {
				name = strings.Split(m[1], ".")[0]
			}
			if regexp.MustCompile(`\b` + regexp.QuoteMeta(name) + `\s*[.(]`).MatchString(usedCode) {
				if previous, ok := byName[name]; ok && previous != line {
					conflicts[name] = true
				} else if !ok {
					order = append(order, name)
					byName[name] = line
				}
				seen[line] = true
			}
		}
	}
	imports := []string{}
	for _, name := range order {
		if !conflicts[name] {
			imports = append(imports, byName[name])
		}
	}
	return strings.Join(imports, "\n")
}
