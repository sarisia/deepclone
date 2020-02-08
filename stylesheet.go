package deepclone

import (
	"context"
	"io/ioutil"
	"log"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"path/filepath"
)

var regURI *regexp.Regexp = regexp.MustCompile(`url\(['"]?(.+?)['"]?\)`)

type Stylesheet struct {
	*Resource

	wait      sync.WaitGroup
	resources []*Resource
}

func NewStylesheet(r *Resource) *Stylesheet {
	// TODO: use bufio.Buffer instead of http.Response.Body
	// to make inline css processable
	s := new(Stylesheet)
	s.Resource = r
	return s
}

func (s *Stylesheet) PerformStylesheet(ctx context.Context, depth int) {
	s.parseCSS()

	s.wait.Add(len(s.resources))
	for _, r := range s.resources {
		go func(rr *Resource) {
			rr.PerformResource(ctx, depth) // CSS don't consume depth
			s.wait.Done()
		}(r)
	}

	s.wait.Wait()
}

func (s *Stylesheet) parseCSS() {
	buf, err := ioutil.ReadAll(s.Body)
	if err != nil {
		log.Printf("Failed to read CSS: %v\n", err)
		return
	}
	bufstr := string(buf)
	// parser := scanner.New(bufstr)
	// this CSS library is shit.
	// TODO: rewrite CSS library that has html parser fashion API

	for _, u := range regURI.FindAllStringSubmatch(bufstr, -1) {
		if u != nil && u[1] != "" {
			// log.Printf("%q\n", u)
			s.recordResource(u[1])
			ur, err := url.Parse(u[1])
			if err != nil {
				log.Printf("Failed to parse url: %v\n", err)
				break
			}
			ur = s.URL.ResolveReference(ur)
			log.Println(ur.String())
			bufstr = strings.Replace(bufstr, u[1], GetReplacePath(s.URL, ur, s.Kind, Any), 1)
		}
	}

	// for tok := parser.Next(); tok.Type != scanner.TokenEOF; tok = parser.Next() {
	// 	if tok.Type == scanner.TokenURI {
	// 		// TODO: つらい
	// 		// https://www.w3.org/TR/css-syntax-3/#token-diagrams
	// 		// この <url-token> 間違えてない？ クォーテーションは？？？
	// 		if u := regURI.FindStringSubmatch(tok.Value); u != nil && u[1] != "" {
	// 			s.recordResource(u[1])
	// 			ur, err := url.Parse(u[1])
	// 			if err != nil {
	// 				log.Printf("Failed to parse url: %v\n", err)
	// 				break
	// 			}
	// 			ur = s.URL.ResolveReference(ur)
	// 			log.Println(ur.String())
	// 			bufstr = strings.Replace(bufstr, u[1], GetReplacePath(s.URL, ur, s.Kind, Any), 1)
	// 			// log.Printf("from CSS: %s\n", u[1])
	// 		}
	// 	}
	// }

	s.save([]byte(bufstr))
}

func (s *Stylesheet) recordResource(rawurl string) {
	// todo: unify this and Site.recordResource
	if strings.HasPrefix(rawurl, "data:") {
		return
	}

	rawurl = strings.TrimSpace(rawurl)
	ru, err := url.Parse(rawurl)
	if err != nil {
		log.Printf("Failed to enclose URL: %v\n", err)
		return
	}
	u := s.URL.ResolveReference(ru)
	// log.Printf("processing %s, raw: %s, resolved: %s\n", s.URL, rawurl, u)

	// record
	s.resources = append(s.resources, NewResource(s.Parent, u, 0))
}

func (s *Stylesheet) save(buf []byte) {
	full := filepath.FromSlash(getFullPath(s.URL, s.Kind))
	f, err := openFile(full)
	if err != nil {
		log.Printf("Failed to create %s: %v\n", full, err)
		return
	}
	defer f.Close()

	size, err := f.Write(buf)
	if err != nil {
		log.Printf("Failed to write buffer to file: %v\n", err)
		return
	}

	log.Printf("Rendered stylesheet %s: %d bytes\n", full, size)
}
