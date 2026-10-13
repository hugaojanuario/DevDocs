package extraction

import (
	"regexp"
	"strings"
)

var ipythonInput = regexp.MustCompile(`^In \[\d+\]: ?`)
var ipythonContinuation = regexp.MustCompile(`^\s+\.\.\.: ?`)
var ipythonOutput = regexp.MustCompile(`^Out\[\d+\]: ?`)

func pythonCommand(line string) (string, bool) {
	if strings.HasPrefix(line, ">>> ") {
		return strings.TrimPrefix(line, ">>> "), true
	}
	if loc := ipythonInput.FindStringIndex(line); loc != nil {
		return line[loc[1]:], true
	}
	return "", false
}

// NormalizeTranscript recognizes explicit Python and shell prompts only. Plain
// source files keep their whitespace; arbitrary output-looking lines are not
// guessed away. SourceCode remains available as the literal evidence.
func NormalizeTranscript(raw, language string) (code, output string) {
	lines := strings.Split(strings.Trim(raw, "\r\n"), "\n")
	python := language == "python" || language == "pycon" || language == "python3"
	shell := language == "shell" || language == "bash" || language == "console"
	interactive := false
	for _, line := range lines {
		_, isPython := pythonCommand(line)
		if python && isPython || shell && strings.HasPrefix(line, "$ ") {
			interactive = true
			break
		}
	}
	if !interactive {
		return strings.Trim(raw, "\r\n"), ""
	}
	commands, results := []string{}, []string{}
	shellContinued := false
	for _, line := range lines {
		command, isPython := pythonCommand(line)
		switch {
		case python && isPython:
			commands = append(commands, command)
		case python && ipythonContinuation.MatchString(line):
			commands = append(commands, ipythonContinuation.ReplaceAllString(line, ""))
		case python && strings.HasPrefix(line, "... "):
			commands = append(commands, strings.TrimPrefix(line, "... "))
		case python && line == "...":
			commands = append(commands, "")
		case shell && strings.HasPrefix(line, "$ "):
			command := strings.TrimPrefix(line, "$ ")
			commands = append(commands, command)
			shellContinued = strings.HasSuffix(command, `\`)
		case shell && shellContinued:
			commands = append(commands, line)
			shellContinued = strings.HasSuffix(line, `\`)
		default:
			if python {
				line = ipythonOutput.ReplaceAllString(line, "")
			}
			results = append(results, line)
		}
	}
	return strings.Trim(strings.Join(commands, "\n"), "\n"), strings.Trim(strings.Join(results, "\n"), "\n")
}
