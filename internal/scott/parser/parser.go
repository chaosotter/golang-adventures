// Package parser contains routines for parsing a Scott Adams adventure game
// into the proto representation we use at runtime.
//
// See the file "Definition" in the ScottFree distribution for more information
// on the underlying file format.
package parser

import (
	"fmt"

	"github.com/chaosotter/golang-adventures/api/scottpb"
	"github.com/chaosotter/golang-adventures/internal/scott/stream"
)

// Parse attempts to initialize a new Game proto from the game file.
func Parse(data []byte) (*scottpb.Game, error) {
	s, err := stream.New(data)
	if err != nil {
		return nil, fmt.Errorf("Could not tokenize game data: %v", err)
	}

	pb := &scottpb.Game{}
	for _, step := range []struct {
		phase string
		fn    func(*scottpb.Game, *stream.Stream) error
	}{
		{"header", loadHeader},
	} {
		if err := step.fn(pb, s); err != nil {
			return nil, fmt.Errorf("Error parsing %s: %v", step.phase, err)
		}
	}

	return pb, nil
}

// loadHeader loads in the game header.
func loadHeader(pb *scottpb.Game, s *stream.Stream) error {
	h := &scottpb.Header{}
	for _, field := range []*int32{
		&h.Unknown0,
		&h.NumItems,
		&h.NumActions,
		&h.NumWords,
		&h.NumRooms,
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
