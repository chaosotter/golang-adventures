// read_scott is a utility to read in and parse a Scott Adams adventure file in
// the TRS-80 format supported by the ScottFree interpreter.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"google.golang.org/protobuf/encoding/prototext"

	"github.com/chaosotter/golang-adventures/internal/scott/game"
)

var gamePath = flag.String("game", "", "Path to the game file in ScottFree (TRS-80) format.")

func main() {
	flag.Parse()

	data, err := ioutil.ReadFile(*gamePath)
	if err != nil {
		log.Fatalf("Could not read %q: %v", *gamePath, err)
	}
	fmt.Printf("Read %d bytes from %q.\n", len(data), *gamePath)

	g, err := game.New(data)
	if err != nil {
		log.Fatalf("Could not parse %q: %v", *gamePath, err)
	}

	fmt.Println(prototext.Format(g.Initial))
}
