package deepclone

import (
	"context"
	"log"
	"net/url"
)

// Perform is a entrypoint of application
func Perform(ctx context.Context, rawurl string, depth int, done chan<- struct{}) {
	defer func() {
		done <- struct{}{}
	}()
	u, err := url.Parse(rawurl)
	if err != nil {
		log.Printf("Failed to enclose URL: %v\n", err)
		return
	}

	res := new(Resource)
	res.URL = u
	res.performResource(ctx, depth)
}
