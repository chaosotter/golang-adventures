// read_scott is a utility to read in and parse a Scott Adams adventure file in
// the TRS-80 format supported by the ScottFree interpreter.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

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

	pb := &scottpb.Game{
		Header: &scottpb.Header{
			Unknown0:      int32(data[0]<<8 | data[1]),
			NumItems:      int32(data[2]<<8 | data[3]),
			NumActions:    int32(data[4]<<8 | data[5]),
			NumWords:      int32(data[6]<<8 | data[7]),
			MaxInventory:  int32(data[8]<<8 | data[9]),
			StartingRoom:  int32(data[10]<<8 | data[11]),
			NumTreasures:  int32(data[12]<<8 | data[13]),
			WordLength:    int32(data[14]<<8 | data[15]),
			LightDuration: int32(data[16]<<8 | data[17]),
			NumMessages:   int32(data[18]<<8 | data[19]),
			TreasureRoom:  int32(data[20]<<8 | data[21]),
			Unknown12:     int32(data[22]<<8 | data[23]),
		},
	}

	for i := 0; i < 28; i += 2 {
		fmt.Printf("%02x%02x ", data[i], data[i+1])
	}
	fmt.Println()

	fmt.Println(pb)
}
