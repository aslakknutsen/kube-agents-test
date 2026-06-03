package scenario

import "strings"

// rewriteExpectLegacy converts the documented expect block (assertion list items as
// siblings of timeout) into an explicit assertions key for standard YAML parsers.
func rewriteExpectLegacy(data []byte) []byte {
	lines := strings.Split(string(data), "\n")
	var out []string

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		indent := len(line) - len(strings.TrimLeft(line, " "))

		if trimmed != "expect:" && !strings.HasPrefix(trimmed, "expect:") {
			out = append(out, line)
			continue
		}

		expectIndent := indent
		childIndent := expectIndent + 2
		out = append(out, line)

		if i+1 >= len(lines) {
			continue
		}

		// Already canonical.
		nextTrim := strings.TrimSpace(lines[i+1])
		if strings.HasPrefix(nextTrim, "assertions:") {
			continue
		}

		if !strings.HasPrefix(nextTrim, "- ") {
			continue
		}

		// Collect assertion block lines until timeout at child indent.
		var block []string
		j := i + 1
		var timeoutLine string
		for j < len(lines) {
			l := lines[j]
			t := strings.TrimSpace(l)
			if t == "" {
				block = append(block, l)
				j++
				continue
			}
			in := len(l) - len(strings.TrimLeft(l, " "))
			if in == childIndent && strings.HasPrefix(t, "timeout:") {
				timeoutLine = l
				j++
				break
			}
			if in < childIndent {
				break
			}
			block = append(block, l)
			j++
		}

		out = append(out, strings.Repeat(" ", childIndent)+"assertions:")
		for _, bl := range block {
			if strings.TrimSpace(bl) == "" {
				out = append(out, bl)
				continue
			}
			out = append(out, strings.Repeat(" ", 2)+bl)
		}
		if timeoutLine != "" {
			out = append(out, timeoutLine)
		}
		i = j - 1
	}

	return []byte(strings.Join(out, "\n"))
}
