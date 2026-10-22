package fuzzy

import "testing"

func TestMatch(t *testing.T) {
	cases := []struct {
		q, c string
		want bool
	}{
		{"", "anything", true},
		{"cr", "code-review", true},
		{"CR", "code-review", true},
		{"code rev", "code-review", true},
		{"rc", "code-review", false},
		{"xyz", "code-review", false},
		{"pdf", "pdf-07", true},
		{"   ", "pdf", true},
	}
	for _, tc := range cases {
		if got := Match(tc.q, tc.c); got != tc.want {
			t.Errorf("Match(%q,%q)=%v want %v", tc.q, tc.c, got, tc.want)
		}
	}
}

func TestFilter(t *testing.T) {
	got := Filter([]string{"code-review", "pdf", "docx"}, Predicate("oc"))
	if len(got) != 1 || got[0] != "docx" {
		t.Fatalf("Filter = %v", got)
	}
}
