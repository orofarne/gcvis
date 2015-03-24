// gzvis is a tool to assist you visualising the operation of
// the go runtime garbage collector.
//
// usage:
//
//     gcvis program [arguments]...
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

var listen = flag.String("listen", ":8083", "Listen address")

var gcvisGraph Graph

func indexHandler(w http.ResponseWriter, req *http.Request) {
	gcvisGraph.write(w)
}

func main() {
	var err error

	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "usage: [flags] %s command <args>...\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	listener, err := net.Listen("tcp4", *listen)
	if err != nil {
		log.Fatal(err)
	}

	pr, pw, _ := os.Pipe()
	gcChan := make(chan *gctrace, 1)
	scvgChan := make(chan *scvgtrace, 1)

	parser := Parser{
		reader:   pr,
		gcChan:   gcChan,
		scvgChan: scvgChan,
	}

	gcvisGraph = NewGraph(strings.Join(os.Args[1:], " "), GCVIS_TMPL)

	go startSubprocess(pw)
	go parser.Run()

	http.HandleFunc("/", indexHandler)

	go http.Serve(listener, nil)

	for {
		select {
		case gcTrace := <-gcChan:
			gcvisGraph.AddGCTraceGraphPoint(gcTrace)
		case scvgTrace := <-scvgChan:
			gcvisGraph.AddScavengerGraphPoint(scvgTrace)
		}
	}
}
