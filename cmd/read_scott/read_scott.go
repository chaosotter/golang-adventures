// read_scott is a utility to read in and parse a Scott Adams adventure file in
// the TRS-80 format supported by the ScottFree interpreter.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"google.golang.org/protobuf/encoding/prototext"

	"github.com/chaosotter/golang-adventures/api/scottpb"
	"github.com/chaosotter/golang-adventures/internal/scott/stream"
)

var game = flag.String("game", "", "Path to the game file in ScottFree (TRS-80) format.")

func main() {
	flag.Parse()

	data, err := ioutil.ReadFile(*game)
	if err != nil {
		log.Fatalf("Could not read %q: %v", *game, err)
	}

	fmt.Println(*game)
	fmt.Println(len(data))

	s, err := stream.New(data)
	if err != nil {
		log.Fatalf("Could not parse %q: %v", *game, err)
	}
	fmt.Println(s.Done())

	pb := &scottpb.Game{}
	if err := LoadHeader(pb, s); err != nil {
		log.Fatalf("Error in header: %v", err)
	}

	fmt.Println(prototext.Format(pb))
}

func LoadHeader(pb *scottpb.Game, s *stream.Stream) error {
	h := &scottpb.Header{}
	for _, field := range []*int32{
		&h.Unknown0,
		&h.NumItems,
		&h.NumActions,
		&h.NumWords,
		&h.MaxInventory,
		&h.StartingRoom,
		&h.NumTreasures,
		&h.WordLength,
		&h.LightDuration,
		&h.NumMessages,
		&h.TreasureRoom,
		&h.Unknown12,
	} {
		val, err := s.NextInt()
		if err != nil {
			return err
		}
		*field = int32(val)
	}

	pb.Header = h
	return nil
}
