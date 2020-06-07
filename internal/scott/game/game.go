// Package game defines the basic runtime object for working with a Scott Adams
// adventure.
package game

import (
	"github.com/chaosotter/golang-adventures/api/scottpb"
	"github.com/chaosotter/golang-adventures/internal/scott/parser"
)

// A Game encaspulates the current state of a Scott Adams adventure.
type Game struct {
	// Initial is a proto holding the initial state of the game in the form of
	// a proto initialized from the external data file.
	Initial *scottpb.Game
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
	}
	return g, nil
}
