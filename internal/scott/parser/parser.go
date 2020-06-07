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
		{"actions", loadActions},
	} {
		if err := step.fn(pb, s); err != nil {
			return nil, fmt.Errorf("Error parsing %s: %v", step.phase, err)
		}
	}

	return pb, nil
}

// loadHeader loads in the game header, which consists of 14 integer values.
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
		//&h.Unknown12,
	} {
		val, err := s.NextInt()
		if err != nil {
			return err
		}
		*field = int32(val)
	}

	// Adjust the Num* fields to relect our modern understanding of arrays.
	h.NumItems++
	h.NumActions++
	h.NumWords++
	h.NumRooms++
	h.NumMessages++

	pb.Header = h
	return nil
}

// loadActions loads in the actions.  Each action has the following form:
//   (150 * verb index) + noun index
//   5x conditions, expressed as condition type + (20 * value)
//   (150 * action0 type) + action1 type
//   (150 * action2 type) + action3 type
func loadActions(pb *scottpb.Game, s *stream.Stream) error {
	for i := 0; i < int(pb.Header.NumActions); i++ {
		a := &scottpb.Action{}

		val, err := s.NextInt()
		if err != nil {
			return fmt.Errorf("Action %d: %v", i, err)
		}
		a.VerbIndex = int32(val / 150)
		a.NounIndex = int32(val % 150)

		for j := 0; j < 5; j++ {
			val, err := s.NextInt()
			if err != nil {
				return fmt.Errorf("Action %d, condition %d: %v", i, j, err)
			}
			a.Conditions = append(a.Conditions, &scottpb.Condition{
				Type:  (scottpb.ConditionType)(val % 20),
				Value: int32(val / 20),
			})
		}

		for j := 0; j < 2; j++ {
			val, err := s.NextInt()
			if err != nil {
				return fmt.Errorf("Action %d, action value %d: %v", i, j, err)
			}
			a.Actions = append(a.Actions, (scottpb.ActionType)(val/150))
			a.Actions = append(a.Actions, (scottpb.ActionType)(val%150))
		}

		pb.Actions = append(pb.Actions, a)
	}

	return nil
}
