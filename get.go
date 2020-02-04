package deepclone

import (
	"net/http"
	"strings"
	"bytes"
	"io"
	"context"
	"log"
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

func GetResource(ctx context.Context, url, referer string) (buf *bytes.Buffer, kind Kind, err error) {
	limit<- struct{}{}
	defer func() {
		<-limit
	}()

	// TODO: handle when url is base64
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("%s: Failed to enclose Request: %v\n", url, err)
		return
	}

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

	kind, ok := GetContentType(&resp.Header)
	if !ok {
		kind = GetExtensionType(url)
	}

	return
}
