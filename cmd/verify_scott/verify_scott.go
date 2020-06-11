// verify_scott is a utility for verifying the parsing routines for loading in
// Scott Adams adventure files in the TRS-80 format supported by the ScottFree
// interpreter.  It does this by loading the game, writing it back out in the
// same format, and doing a diff on the results.
package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"strings"

	"github.com/kylelemons/godebug/pretty"

	"github.com/chaosotter/golang-adventures/internal/scott/game"
	"github.com/chaosotter/golang-adventures/internal/scott/writer"
)

var gamePath = flag.String("game", "", "Path to the game file in ScottFree (TRS-80) format.")

func main() {
	flag.Parse()

	data, err := ioutil.ReadFile(*gamePath)
	if err != nil {
		log.Fatalf("Could not read %q: %v", *gamePath, err)
	}

	g, err := game.New(data)
	if err != nil {
		log.Fatalf("Could not parse %q: %v", *gamePath, err)
	}

	b := &bytes.Buffer{}
	writer.WriteTRS80(b, g.Initial)

	got := strings.Split(b.String(), "\n")
	want := strings.Split(string(data), "\n")
	if diff := pretty.Compare(got, want); diff != "" {
		log.Fatalf("Diff detected:\n%s", diff)
	}
}
