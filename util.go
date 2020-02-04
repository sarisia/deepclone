package deepclone

import (
	"net/http"
	"log"
	"strings"
	"path"
	"golang.org/x/net/html"
	"net/url"
	"path/filepath"
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
	"text/css": CSS,
}
var extMap = map[string]Kind{
	"css": CSS,
	"html": HTML,
	"htm": HTML,
}

func GetContentType(header *http.Header) (Kind, bool) {
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

func GetExtensionType(url string) Kind {
	ext := path.Ext(path.Clean(url))
	
	if k, ok := extMap[ext]; ok {
		return k
	} else {
		return Any
	}
}

func getFullPath(u *url.URL, kind Kind) (full string) {
	full = path.Join(u.Hostname(), u.Path)

	// need to determine urls below as Site
	// fuck.com
	// fuck.com/
	// fuck.com/shit
	// fuck.com/shit/
	// fuck.com/yes.shit <- rare case, seems replaced automatically by modern server
	// fuck.com/yes.shit/ <- possible, but gaiji
	// -> which has suffix "/" is always Site
	// -> no Path or Path does not have ext is Site
	
	// fuck apple.com
	// https://www.apple.com/wss/fonts?families=SF+Pro,v2|SF+Pro+Icons,v1
	// DO NOT have extension but it's CSS!
	if strings.HasSuffix(full, "/") || u.Path == "" || path.Ext(u.Path) == "" {
		switch (kind) {
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
	return
}

func getStylesheetResource(attrs []html.Attribute) (pos int, ok bool) {
	for i, a := range attrs {
		switch (a.Key) {
		case "rel":
			ok = a.Val == "stylesheet"
		case "href":
			pos = i
		}
	}
	return
}

func GetReplacePath(base, target *url.URL, basekind, targkind Kind) (relative string) {
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
	log.SetFlags(log.Ldate|log.Ltime|log.Lshortfile)
	log.Println("Logger flag set")
}
