package deepclone

import (
	"net/url"
	"testing"
)

var testGetFullPathCases = []struct {
	in   string
	kind Kind
	want string
}{
	{"sarisia.cc", Any, "sarisia.cc/index.html"},
	{"sarisia.cc/", Any, "sarisia.cc/index.html"},
	{"sarisia.cc/path", Any, "sarisia.cc/path/index.html"},
	{"sarisia.cc/path/", Any, "sarisia.cc/path/index.html"},
	{"sarisia.cc/path.dot", Any, "sarisia.cc/path.dot/index.html"},
	{"sarisia.cc/path.dot/", Any, "sarisia.cc/path.dot/index.html"},
}

func TestGetFullPath(t *testing.T) {
	for _, tc := range testGetFullPathCases {
		t.Run("URL="+tc.in, func(t *testing.T) {
			u, err := url.Parse(tc.in)
			if err != nil {
				t.Errorf("URL parse failed: %v\n", err)
				return
			}
			s := getFullPath(u, tc.kind)
			if s != tc.want {
				t.Errorf("want: %s, got: %s\n", tc.want, s)
			}
		})
	}
}
