// convert_scott is a utility for converting a Scott Adams adventure file from
// the TRS-80 format supported by the ScottFree interpreter to the proto-based
// version used by this system.
package main

import (
	"flag"
	"io/ioutil"
	"log"

	"google.golang.org/protobuf/proto"

	"github.com/chaosotter/golang-adventures/internal/scott/game"
)

var (
	inPath  = flag.String("in", "", "Path to the input game file in ScottFree (TRS-80) format.")
	outPath = flag.String("out", "", "Path to the output game file in proto format.")
)

func main() {
	flag.Parse()
	g := game.MustLoadFromFile(*inPath)

	wire, err := proto.Marshal(g.Initial)
	if err != nil {
		log.Fatalf("Could not marshal proto: %v", err)
	}

	if err := ioutil.WriteFile(*outPath, wire, 0664); err != nil {
		log.Fatalf("Could not write %q: %v", *outPath, err)
	}
}
