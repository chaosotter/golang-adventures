// Package game defines the basic runtime object for working with a Scott Adams
// adventure.
package game

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/chaosotter/golang-adventures/api/scottpb"
	"github.com/chaosotter/golang-adventures/internal/scott/parser"
)

const (
	NumFlags    = 32 // number of flags
	NumCounters = 16 // number of counters
)

const (
	LightItem    = 9  // constant across all adventures
	Inventory    = -1 // location corresponding to player inventory
	DarkFlag     = 15 // flag number for darkness
	LightOutFlag = 16 // flag number for light gone out
	UnknownWord  = -1 // value used to represent unknown words
)

const (
	AutoVerb = 0 // used for automatic actions
	GoVerb   = 1 // used for motion ("GO")
)

// Status is used as a return code for the attempted execution of a command.
type Status int

const (
	Success       = Status(iota) // command was processed successfully
	Unknown                      // command wasn't understood
	NoDirection                  // "GO" command with no direction
	BadDirection                 // "GO" command with an invalid direction
	DangerousDark                // a dangerous move in the dark (valid direction)
	DeadDark                     // a fatal move in the dark (invalid direction)
	Unsuccessful                 // command is valid but couldn't be fulfilled
)

// A Game encaspulates the current state of a Scott Adams adventure.
type Game struct {
	// Initial is a proto holding the initial state of the game in the form of
	// a proto initialized from the external data file.
	Initial *scottpb.Game

	// Current is a deep copy of Initial that contains the current game state.
	Current *scottpb.Game

	// DefaultCommand is the default "auto-execution" command for this game.
	// It is always based on verb 0 and noun 0, but the text might vary from
	// game to game.
	DefaultCommand *ParseData
}

// New initializes a fresh Game value from the raw bytes read from the external
// game file.
func New(data []byte) (*Game, error) {
	pb, err := parser.Parse(data)
	if err != nil {
		return nil, err
	}

	g := &Game{
		Initial: pb,
		DefaultCommand: &ParseData{
			Verb:      pb.Verbs[AutoVerb].Word,
			VerbIndex: AutoVerb,
			Noun:      pb.Nouns[0].Word,
		},
	}

	g.Restart()
	return g, nil
}

// MustLoadFromFile tries to initialize a fresh Game value from the given file
// or aborts the process.
func MustLoadFromFile(path string) *Game {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Could not read %q: %v", path, err)
	}

	g, err := New(data)
	if err != nil {
		log.Fatalf("Could not parse %q: %v", path, err)
	}

	return g
}

// LookData encapsulates all of the data that go into refreshing the room
// description.  We avoid direct output for the benefit of writing additional
// drivers.
type LookData struct {
	IsDark          bool     // true if it's too dark to see
	RoomDescription string   // the room description, made into a sentence
	Exits           []string // ordered list of obvious exits
	Items           []string // ordered list of items in the room
}

// Look prints the standard description information to the given output.
func (g *Game) Look() *LookData {
	ld := &LookData{}

	if g.IsDark() {
		ld.RoomDescription = "I can't see. It is too dark!"
		return ld
	}

	r := g.Current.Rooms[g.Current.State.Location]
	if r.Literal {
		ld.RoomDescription = r.Description
	} else {
		ld.RoomDescription = fmt.Sprintf("I'm in a %s", r.Description)
	}

	for i, dir := range []string{"North", "South", "East", "West", "Up", "Down"} {
		if r.Exits[i] != 0 {
			ld.Exits = append(ld.Exits, dir)
		}
	}

	for i := 0; i < int(g.Current.Header.NumItems); i++ {
		if it := g.Current.Items[i]; it.Location == g.Current.State.Location {
			ld.Items = append(ld.Items, it.Description)
		}
	}

	return ld
}

// ParseData encapsulates all of the data from parsing a command.
type ParseData struct {
	Verb      string // the (uppercase, truncated) text of the entered verb
	VerbIndex int    // the index of the verb, or UnknownWord
	Noun      string // the (uppercase, truncated) text of the entered noun
	NounIndex int    // the index of the noun, or UnknownWord
}

// Parse parses the given line of user input into a verb and noun.  It does not
// attempt to execute the command.
func (g *Game) Parse(input string) *ParseData {
	input = strings.ToUpper(strings.TrimSpace(input))

	var verb, noun string
	words := strings.SplitN(input, " ", 2)
	if len(words) > 0 {
		verb = words[0]
	}
	if len(words) > 1 {
		noun = words[1]
	}

	switch verb {
	case "N", "NORTH":
		verb, noun = "GO", "NORTH"
	case "S", "SOUTH":
		verb, noun = "GO", "SOUTH"
	case "W", "WEST":
		verb, noun = "GO", "WEST"
	case "E", "EAST":
		verb, noun = "GO", "EAST"
	case "U", "UP":
		verb, noun = "GO", "UP"
	case "D", "DOWN":
		verb, noun = "GO", "DOWN"
	case "I":
		verb = "INVENTORY"
	}

	return &ParseData{
		Verb:      verb,
		VerbIndex: g.findWord(g.Current.Verbs, verb),
		Noun:      noun,
		NounIndex: g.findWord(g.Current.Nouns, noun),
	}
}

// findWord finds the index of |w| within |ws|.  Synonyms are resolved down to
// the lowest matching index.
func (g *Game) findWord(ws []*scottpb.Word, w string) int {
	w = (w + "      ")[0:g.Current.Header.WordLength]
	for i := 0; i < len(ws); i++ {
		if (ws[i].Word + "      ")[0:g.Current.Header.WordLength] != w {
			continue
		}
		for j := i; j >= 0; j-- {
			if !ws[j].Synonym {
				return j
			}
		}
	}
	return UnknownWord
}

// Execute the given command.
func (g *Game) Execute(pd *ParseData) Status {
	// Try movement first as a special case.
	if pd.VerbIndex == GoVerb {
		switch {
		case pd.NounIndex == UnknownWord:
			return NoDirection
		case pd.NounIndex >= 1 && pd.NounIndex <= 6: // nouns 1..6 are always the directions
			dark := g.IsDark()
			dest := g.Current.Rooms[g.Current.State.Location].Exits[pd.NounIndex-1]
			switch {
			case dark && dest == 0:
				g.KillPlayer()
				return DeadDark
			case dark:
				g.Current.State.Location = dest
				return DangerousDark
			case dest == 0:
				return BadDirection
			default:
				g.Current.State.Location = dest
				return Success
			}
		}
	}

	return Success
}

// Execute the default commands (i.e., actions that represent the passage of
// time rather a reaction to user input).  This is always verb 0 (usually
// "AUTO") and noun 0 (usually "ANY").
func (g *Game) ExecuteDefault() Status {
	return g.Execute(g.DefaultCommand)
}

// IsDark checks if the player is currently in the dark.
func (g *Game) IsDark() bool {
	return g.Current.State.Flags[DarkFlag] &&
		(g.Current.Items[LightItem].Location != Inventory) &&
		(g.Current.Items[LightItem].Location != g.Current.State.Location)
}

// KillPlayer kills the player.  This turns darkness off and places them in the
// last room in the game (action DEATH).
func (g *Game) KillPlayer() {
	g.Current.State.Location = g.Current.Header.NumRooms - 1
	g.Current.State.Flags[DarkFlag] = false
}

// Restart the game.
func (g *Game) Restart() {
	g.Current = proto.Clone(g.Initial).(*scottpb.Game)
	g.Current.State = &scottpb.State{
		Location: g.Current.Header.StartingRoom,
		Flags:    make([]bool, NumFlags),
		Counters: make([]int32, NumCounters),
	}
}
