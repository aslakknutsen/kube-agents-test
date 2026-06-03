package scenario

import (
	"strings"
	"testing"
)

func TestRewriteExpectLegacy(t *testing.T) {
	in := `expect:
  - resource:
      apiVersion: v1
    conditions:
      - path: .x
        value: 1
  timeout: 10s
`
	out := string(rewriteExpectLegacy([]byte(in)))
	if !strings.Contains(out, "assertions:") {
		t.Fatalf("missing assertions key:\n%s", out)
	}
	if strings.Contains(out, "did not find") {
		t.Fatal(out)
	}
}
