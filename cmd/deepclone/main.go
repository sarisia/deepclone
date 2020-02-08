package main

import (
	"context"
	"flag"
	"github.com/sarisia/deepclone"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"time"
)

func main() {
	go debugRoutine()
	depth := flag.Int("depth", 1, "fetch depth")
	conns := flag.Int("conn", 16, "max concurrent connections")
	dir := flag.String("dir", "", "directory to save contents")
	flag.Parse()
	url := flag.Arg(0)
	if url == "" {
		log.Println("No URL is provided.")
		return
	}

	start := time.Now()
	deepclone.SetLoggerFlags()

	// this will fuck the app and it is golang's bug
	// https://github.com/golang/go/issues/34941
	// so comment out this
	deepclone.SetMaxConnsPerHost(*conns)

	deepclone.SetDirectory(*dir)

	log.Println("Starting...")
	log.Printf("Max concurrent connections: %d\n", *conns)
	log.Printf("Fetch depth: %d\n", *depth)

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go deepclone.Perform(ctx, url, *depth, done)
	// 虚無じゃん
	// https://qiita.com/cia_rana/items/a2c3e1609bd25a5c5596
	// https://golang.org/ref/spec#Break_statements
Finish:
	for {
		select {
		case <-sig:
			log.Printf("Interrupted...")
			cancel()
		case <-done:
			log.Println("Done")
			break Finish
		}
	}
	log.Printf("Finish: %s\n", time.Since(start))
}

func debugRoutine() {
	log.Println(http.ListenAndServe("localhost:6611", nil))
}
