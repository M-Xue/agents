package markdown

import (
	"strings"
	"testing"
)

const sample = "---\nname: grilling\ndescription: Grill the user.\n---\n\n# Title\n\nSome **bold** text.\n\n```\ncode\n```\n"

func TestSplitFrontmatter(t *testing.T) {
	fm, body := SplitFrontmatter(sample)
	if fm.Name != "grilling" || fm.Description != "Grill the user." {
		t.Fatalf("frontmatter = %+v", fm)
	}
	if !strings.HasPrefix(body, "# Title") {
		t.Fatalf("body = %q", body)
	}
	fm, body = SplitFrontmatter("plain text")
	if fm.Name != "" || body != "plain text" {
		t.Fatalf("no-frontmatter case: %+v %q", fm, body)
	}
}

func TestRenderStripsMarkersAndKeepsText(t *testing.T) {
	out := Render(sample, 40)
	for _, want := range []string{"grilling", "Grill the user.", "Title", "bold", "code"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "**") || strings.Contains(out, "name:") {
		t.Errorf("markers not stripped:\n%s", out)
	}
}
