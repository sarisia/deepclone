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
	{"https://www.apple.com/wss/fonts?families=SF+Pro,v2|SF+Pro+Icons,v1", CSS,
		"www.apple.com/wss/fonts-families-SF-Pro-v2-SF-Pro-Icons-v1.css"},
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

var testToFsSafeStringCases = []struct {
	in   string
	want string
}{
	{"query=yes", "query-yes"},
	{"query=yes&query2=no", "query-yes-query2-no"},
	{"empty=&query=yes", "empty--query-yes"},
	{"families=SF+Pro,v2|SF+Pro+Icons,v1", "families-SF-Pro-v2-SF-Pro-Icons-v1"},
}

func TestToFsSafeString(t *testing.T) {
	for _, tc := range testToFsSafeStringCases {
		t.Run("In="+tc.in, func(t *testing.T) {
			ss := toFsSafeString(tc.in)
			if ss != tc.want {
				t.Errorf("want: %s, got: %s\n", tc.want, ss)
			}
		})
	}
}
