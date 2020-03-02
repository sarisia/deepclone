package deepclone

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// TODO: make configurable
var cli = &http.Client{
	Timeout: 30 * time.Second,
}
var limit = make(chan struct{}, 16)

func SetMaxConnsPerHost(limit int) {
	http.DefaultTransport.(*http.Transport).MaxConnsPerHost = limit
	log.Printf("MaxConnsPerHost set to %d\n", limit)
}

func getResource(ctx context.Context, url, referer string) (buf *bytes.Buffer, kind Kind, err error) {
	limit <- struct{}{}
	defer func() {
		<-limit
	}()

	// TODO: handle when url is base64
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("%s: Failed to enclose Request: %v\n", url, err)
		return
	}

	// User Agent!
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36")

	if referer != "" {
		if !strings.HasSuffix(referer, "/") {
			referer += "/"
		}
		req.Header.Add("Referer", referer)
	}

	resp, err := cli.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	log.Printf("%s: %s ref: %s\n", resp.Status, url, referer)

	buf = bytes.NewBuffer(nil)
	io.Copy(buf, resp.Body)

	kind, ok := getContentType(&resp.Header)
	if !ok {
		kind = getExtensionType(url)
	}

	return
}
