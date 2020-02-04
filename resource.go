package deepclone

import (
	"path"
	"bytes"
	"context"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sync"
)

var cached = map[string]struct{}{}
var cachedLock = sync.RWMutex{}

type Resource struct {
	Parent *Site
	URL    *url.URL
	Kind   Kind
	Body   *bytes.Buffer
}

func NewResource(parent *Site, u *url.URL, kind Kind) *Resource {
	return &Resource{parent, u, kind, nil}
}

func (r *Resource) PerformResource(ctx context.Context, depth int) {
	// ctx, cancel := context.WithTimeout(parent, 30*time.Second)
	// defer cancel()

	full := path.Join(r.URL.Hostname(), r.URL.Path)
	cachedLock.RLock()
	_, ok := cached[full]
	cachedLock.RUnlock()
	if ok {
		log.Printf("Already cached. Abort: %s\n", full)
		return
	}

	cachedLock.Lock()
	cached[full] = struct{}{}
	cachedLock.Unlock()

	if err := r.get(ctx); err != nil {
		log.Printf("Failed to get: %v\n", err)
		return
	}

	// switch to Site if Kind is HTML
	switch r.Kind {
	case HTML:
		// switch to Site if Kind is HTML
		s := NewSite(r)
		s.PerformSite(ctx, depth)
		return
	case CSS:
		ss := NewStylesheet(r)
		ss.PerformStylesheet(ctx, depth)
		return
	}

	r.save(nil)
}

func (r *Resource) get(ctx context.Context) error {
	// set referer if Resource has Parent Site
	var ref string
	if r.Parent != nil {
		ref = r.Parent.URL.String()
	}

	buf, kind, err := GetResource(ctx, r.URL.String(), ref)
	if err != nil {
		// log.Printf("Failed to get resource: %v\n", err)
		return err
	}
	r.Body = buf

	if r.Kind == Unknown {
		r.Kind = kind
	}

	return nil
}

func (r *Resource) save(buf []byte) {
	// make sure to close Response.Body here
	full := filepath.FromSlash(getFullPath(r.URL, r.Kind))

	// create base dir
	if err := os.MkdirAll(filepath.Dir(full), 0777); err != nil {
		log.Printf("Failed to create directory: %v\n", err)
		return
	}

	f, err := os.Create(full)
	if err != nil {
		log.Printf("Failed to create file: %v\n", err)
		return
	}
	defer f.Close()

	if buf != nil {
		_, err = f.Write(buf)
		if err != nil {
			log.Printf("Failed to write buffer to file: %v\n", err)
			return
		}
	} else {
		_, err = io.Copy(f, r.Body)
		if err != nil {
			log.Printf("Failed to write to file: %v\n", err)
			return
		}
	}
	log.Printf("Created %s\n", full)
}
