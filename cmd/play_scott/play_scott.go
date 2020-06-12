// play_scott is an actual (single-player) game driver for Scott Adams adventure
// games.  Credit is owed to the ScottFree driver for understanding of the
// underlying file format and semantics.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/chaosotter/golang-adventures/internal/scott/game"
)

var gamePath = flag.String("game", "", "Path to the game file in ScottFree (TRS-80) format.")

func main() {
	flag.Parse()
	g := game.MustLoadFromFile(*gamePath)

	fmt.Println("Welcome to the play_scott driver for Scott Adams adventures.")
	fmt.Println("This is a single-player driver in Go based very loosely on the")
	fmt.Println("C-language ScottFree interpreter.")
	fmt.Println()
	fmt.Println("There aren't many bells and whistles here, as the main aim of the")
	fmt.Println("project is to provide multiplayer (MUD-like) support.")
	fmt.Println()
	fmt.Printf("Loaded Version %d.%02d of Adventure #%d.\n\n",
		g.Initial.Footer.Version/100, g.Initial.Footer.Version%100, g.Initial.Footer.Adventure)

	g.Restart()
	Look(g.Look())

	in := bufio.NewScanner(os.Stdin)
	os.Stdout.Write([]byte("Tell me what to do ? "))
	if in.Scan() {
		pd := g.Parse(in.Text())
		fmt.Printf("I got this: %v\n", pd)
	}
}

func Look(ld *game.LookData) {
	fmt.Println(ld.RoomDescription)

	fmt.Printf("Obvious exits: ")
	if len(ld.Exits) > 0 {
		fmt.Printf("%s\n", strings.Join(ld.Exits, ", "))
	} else {
		fmt.Println("None\n")
	}

	// TODO: Handle line-wrapping.
	if len(ld.Items) > 0 {
		fmt.Printf("\nI can also see: %s\n", strings.Join(ld.Items, " - "))
	}

	fmt.Println()
}
