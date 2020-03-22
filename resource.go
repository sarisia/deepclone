package deepclone

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/url"
	"path"
	"path/filepath"
	"sync"
)

var cached = map[string]struct{}{}
var cachedLock = sync.RWMutex{}

type Resource struct {
	Parent string
	URL    *url.URL
	Kind   Kind
	Body   *bytes.Buffer
}

func newResource(parent string, u *url.URL, kind Kind) *Resource {
	return &Resource{parent, u, kind, nil}
}

func (r *Resource) performResource(ctx context.Context, depth int) {
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
		s.performSite(ctx, depth)
		return
	case CSS:
		ss := NewStylesheet(r)
		ss.performStylesheet(ctx, depth)
		return
	}

	r.save()
}

func (r *Resource) get(ctx context.Context) error {
	// set referer if Resource has Parent Site
	buf, kind, err := getResource(ctx, r.URL.String(), r.Parent)
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

func (r *Resource) save() {
	// make sure to close Response.Body here
	full := filepath.FromSlash(getFullPath(r.URL, r.Kind))
	f, err := openFile(full)
	if err != nil {
		log.Printf("Failed to create %s: %v\n", full, err)
		return
	}
	defer f.Close()

	size, err := io.Copy(f, r.Body)
	if err != nil {
		log.Printf("Failed to write to file: %v\n", err)
		return
	}
	log.Printf("Created %s: %d bytes\n", full, size)
}
