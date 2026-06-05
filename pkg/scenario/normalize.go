package scenario

import "strings"

// normalizeScenarioYAML rewrites documented hybrid expect blocks into strict YAML.
func normalizeScenarioYAML(src []byte) []byte {
	lines := strings.Split(string(src), "\n")
	out := make([]string, 0, len(lines)+2)

	for i := 0; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "expect:" {
			out = append(out, lines[i])
			continue
		}

		out = append(out, lines[i])
		i++

		var block []string
		for i < len(lines) && isExpectContinuation(lines[i]) {
			block = append(block, lines[i])
			i++
		}
		i--

		out = append(out, normalizeExpectBlock(block)...)
	}

	return []byte(strings.Join(out, "\n"))
}

func isExpectContinuation(line string) bool {
	if line == "" {
		return true
	}
	return strings.HasPrefix(line, "  ")
}

func normalizeExpectBlock(block []string) []string {
	hasSequence := false
	hasTimeout := false
	for _, line := range block {
		if strings.HasPrefix(line, "  - ") {
			hasSequence = true
		}
		if strings.HasPrefix(line, "  timeout:") {
			hasTimeout = true
		}
	}

	if !hasSequence || !hasTimeout {
		return block
	}

	out := []string{"  assertions:"}
	for _, line := range block {
		if strings.HasPrefix(line, "  timeout:") {
			out = append(out, line)
			continue
		}
		out = append(out, "  "+line)
	}
	return out
}
