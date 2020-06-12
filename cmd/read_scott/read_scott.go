// read_scott is a utility to read in and parse a Scott Adams adventure file in
// the TRS-80 format supported by the ScottFree interpreter.
package main

import (
	"flag"
	"fmt"

	"google.golang.org/protobuf/encoding/prototext"

	"github.com/chaosotter/golang-adventures/internal/scott/game"
)

var gamePath = flag.String("game", "", "Path to the game file in ScottFree (TRS-80) format.")

func main() {
	flag.Parse()
	g := game.MustLoadFromFile(*gamePath)

	fmt.Println(prototext.Format(g.Initial))
}
