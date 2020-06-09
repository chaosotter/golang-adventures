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
		{"words", loadWords},
		{"rooms", loadRooms},
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

// loadWords loads in the verbs and nouns, which are an interleaved array of
// strings.  An initial "*" indicates a synonym.
func loadWords(pb *scottpb.Game, s *stream.Stream) error {
	for i := 0; i < int(pb.Header.NumWords); i++ {
		val, err := s.NextString()
		if err != nil {
			return fmt.Errorf("Verb %d: %v", i, err)
		}
		pb.Verbs = append(pb.Verbs, makeWord(val))

		val, err = s.NextString()
		if err != nil {
			return fmt.Errorf("Noun %d: %v", i, err)
		}
		pb.Nouns = append(pb.Nouns, makeWord(val))
	}

	return nil
}

// makeWord takes care of making a Word proto from an input string, parsing out
// the synonym flag if necessary.
func makeWord(raw string) *scottpb.Word {
	if len(raw) > 0 && raw[0] == '*' {
		return &scottpb.Word{Word: raw[1:len(raw)], Synonym: true}
	}
	return &scottpb.Word{Word: raw}
}

// loadRooms loads in the rooms, of which consists of six directions followed by
// a description.  The description starts with "*" to indicate that it stands
// alone, with no "I'm in a" prefix.
func loadRooms(pb *scottpb.Game, s *stream.Stream) error {
	for i := 0; i < int(pb.Header.NumRooms); i++ {
		r := &scottpb.Room{}
		for _, field := range []*int32{
			&r.North,
			&r.South,
			&r.East,
			&r.West,
			&r.Up,
			&r.Down,
		} {
			val, err := s.NextInt()
			if err != nil {
				return fmt.Errorf("Room %d, directions: %v", i, err)
			}
			*field = int32(val)
		}

		desc, err := s.NextString()
		if err != nil {
			return fmt.Errorf("Room %d, description: %v", i, err)
		}
		if len(desc) > 0 && desc[0] == '*' {
			r.Description = desc[1:len(desc)]
			r.Literal = true
		} else {
			r.Description = desc
		}

		pb.Rooms = append(pb.Rooms, r)
	}

	return nil
}
