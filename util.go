package deepclone

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

type Kind int

const (
	// https://golang.org/ref/spec#Iota
	// iota always starts from zero
	// means zero value of int (Kind) is Unknown
	Unknown Kind = iota
	HTML
	CSS
	Any
)

var kindMap = map[string]Kind{
	"text/html": HTML,
	"text/css":  CSS,
}
var extMap = map[string]Kind{
	"css":  CSS,
	"html": HTML,
	"htm":  HTML,
}

var dir string

func getContentType(header *http.Header) (Kind, bool) {
	ct := header.Get("Content-Type")
	if ct == "" {
		return Unknown, false
	}

	for _, t := range strings.Split(ct, ";") {
		if k, ok := kindMap[strings.Trim(t, " ")]; ok {
			return k, true
		}
	}

	return Any, true
}

func getExtensionType(url string) Kind {
	ext := path.Ext(path.Clean(url))

	if k, ok := extMap[ext]; ok {
		return k
	}

	return Any
}

func getFullPath(u *url.URL, kind Kind) (full string) {
	full = path.Join(u.Hostname(), u.Path)

	// need to determine urls below as Site
	// hoge.com
	// hoge.com/
	// hoge.com/fuga
	// hoge.com/fuga/
	// hoge.com/yes.fuga <- rare case, seems replaced automatically by modern server
	// hoge.com/yes.fuga/ <- possible, but gaiji
	// -> which has suffix "/" is always Site
	// -> no Path or Path does not have ext is Site

	// omg apple.com
	// https://www.apple.com/wss/fonts?families=SF+Pro,v2|SF+Pro+Icons,v1
	// DO NOT have extension but it's CSS!
	if strings.HasSuffix(full, "/") || u.Path == "" || path.Ext(u.Path) == "" {
		switch kind {
		case CSS:
			if path.Base(u.Path) == "." || path.Base(u.Path) == "/" {
				full = path.Join(full, "stylesheet.css")
			} else {
				full = path.Join(path.Dir(full), path.Base(full)+".css")
			}
		default:
			full = path.Join(full, "index.html")
		}
	}

	d, f := path.Split(full)
	fext := path.Ext(f)
	fbase := strings.TrimSuffix(f, fext)
	sq := toFsSafeString(u.RawQuery)
	if sq != "" {
		sq = "-" + sq
	}
	full = path.Join(d, fbase+sq+fext)
	return
}

func toFsSafeString(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if _, err := b.WriteRune(r); err != nil {
				log.Printf("WriteRune failed: %v\n", err)
				return ""
			}
		} else {
			if _, err := b.WriteString("-"); err != nil {
				log.Printf("WriteString failed: %v\n", err)
				return ""
			}
		}
	}

	return b.String()
}

func getStylesheetResource(attrs []html.Attribute) (pos int, ok bool) {
	for i, a := range attrs {
		switch a.Key {
		case "rel":
			ok = a.Val == "stylesheet"
		case "href":
			pos = i
		}
	}
	return
}

func getReplacePath(base, target *url.URL, basekind, targkind Kind) (relative string) {
	// Absolute path means cross domain href
	// we need to replace abs to rel in order to make clone page
	// works if the cache location is changed

	// use filepath.Rel to generate relative reference

	targpath := filepath.FromSlash(getFullPath(target, targkind))
	// log.Printf("targpath %s\n", targpath)
	basepath := filepath.FromSlash(path.Dir(getFullPath(base, basekind)))
	// log.Printf("basepath %s\n", basepath)
	rel, err := filepath.Rel(basepath, targpath)
	if err != nil {
		log.Printf("Failed to get relative: %v\n", err)
		// fuck!
		relative = path.Join("../", getFullPath(target, targkind))
		return
	}

	relative = filepath.ToSlash(rel)
	// log.Printf("rel %s relpath %s\n", rel, relpath)

	return
}

func SetLoggerFlags() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Logger flag set")
}

func SetDirectory(d string) {
	dir = d
	log.Printf("content directory set to %s\n", dir)
}

func openFile(fullFilepath string) (*os.File, error) {
	full := filepath.Join(dir, fullFilepath)
	if err := os.MkdirAll(filepath.Dir(full), 0777); err != nil {
		return nil, errors.New("failed to create directory")
	}

	f, err := os.Create(full)
	if err != nil {
		return nil, errors.New("failed to create file")
	}

	return f, nil
}
