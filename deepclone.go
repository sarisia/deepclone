package deepclone

import (
	"context"
	"net/url"
	"log"
)

// Perform is a entrypoint of application
func Perform(ctx context.Context, rawurl string, depth int, done chan<- struct{}) {
	u, err := url.Parse(rawurl)
	if err != nil {
		log.Printf("Failed to enclose URL: %v\n", err)
		done<- struct{}{}
		return
	}

	res := new(Resource)
	res.URL = u
	res.PerformResource(ctx, depth)
	done<- struct{}{}
}
