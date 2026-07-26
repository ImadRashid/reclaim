package update

import "testing"

func TestNewer(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"v0.2.0", "0.1.1", true},
		{"0.2.0", "v0.2.0", false},
		{"v0.1.1", "0.2.0", false},
		{"v1.0.0", "0.9.9", true},
		{"v0.2.10", "0.2.9", true},
		{"", "0.2.0", false},
		{"v0.3.0", "", false},
		{"dev", "0.2.0", false},
		{"v0.3.0", "dev", false},
		{"v0.2.0-rc1", "0.1.0", false}, // pre-release suffixes don't parse; stay quiet
	}
	for _, c := range cases {
		if got := newer(c.a, c.b); got != c.want {
			t.Errorf("newer(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}
