package deepclone

import (
	"context"
	"log"
	"net/url"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

type Site struct {
	*Resource

	wait      sync.WaitGroup
	resources []*Resource
}

func NewSite(r *Resource) *Site {
	s := new(Site)
	s.Resource = r
	return s
}

func (s *Site) performSite(ctx context.Context, depth int) {
	if depth < 1 {
		// log.Println("Max depth exceed. Stop.")
		return
	}

	s.parseHTML()

	s.wait.Add(len(s.resources))
	for _, r := range s.resources {
		go func(rr *Resource) {
			rr.performResource(ctx, depth-1)
			s.wait.Done()
		}(r)
	}

	s.wait.Wait()
}

func (s *Site) parseHTML() {
	// log.Println("Parsing!")
	node, err := html.Parse(s.Body)
	if err != nil {
		log.Printf("Failed to parse HTML: %v\n", err)
		return
	}

	s.parseNode(node)
	s.render(node)
}

func (s *Site) parseNode(node *html.Node) {
	if node.Type == html.ElementNode {
		switch node.Data {
		// we need to prevent downloading unneeded pages if <link rel="alternate" />
		case "link":
			if i, ok := getStylesheetResource(node.Attr); ok {
				if rep := s.handleExternalResource(node.Attr[i].Val, CSS); rep != "" {
					node.Attr[i].Val = rep
				}
			}
		// TODO: just do it.
		default:
			for i, attr := range node.Attr {
				switch attr.Key {
				case "href", "src":
					// don't process data URI
					// https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/Data_URIs
					if strings.HasPrefix(attr.Val, "data:") {
						break
					}
					if rep := s.handleExternalResource(node.Attr[i].Val, Unknown); rep != "" {
						node.Attr[i].Val = rep
					}
				}
			}
		}
	}

	for n := node.FirstChild; n != nil; n = n.NextSibling {
		s.parseNode(n)
	}
}

func (s *Site) handleExternalResource(rawpath string, kind Kind) (replace string) {
	rawpath = strings.TrimSpace(rawpath)
	rawurl, err := url.Parse(rawpath)
	if err != nil {
		log.Printf("Failed to enclose URL: %v\n", err)
		return
	}
	fullurl := s.URL.ResolveReference(rawurl)
	// log.Printf("processing %s, raw: %s, resolved: %s\n", s.URL, rawpath, u)

	// record
	s.resources = append(s.resources, newResource(s.URL.String(), fullurl, kind))

	// return replace path
	return s.getReplace(fullurl, kind)
}

func (s *Site) getReplace(fullurl *url.URL, kind Kind) (replace string) {
	return getReplacePath(s.URL, fullurl, s.Kind, kind)
}

func (s *Site) render(node *html.Node) {
	full := filepath.FromSlash(getFullPath(s.URL, s.Kind))
	f, err := openFile(full)
	if err != nil {
		log.Printf("Failed to create %s: %v\n", full, err)
		return
	}
	defer f.Close()

	if err := html.Render(f, node); err != nil {
		log.Printf("Failed to render HTML: %v\n", err)
		return
	}
	log.Printf("Page rendered: %s\n", full)
}
