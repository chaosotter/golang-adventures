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
	in := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println()
		Look(g.Look())

		os.Stdout.Write([]byte("Tell me what to do ? "))
		if in.Scan() {
			pd := g.Parse(in.Text())
			fmt.Printf("I got this: %q<%d> %q<%d>\n", pd.Verb, pd.VerbIndex, pd.Noun, pd.NounIndex)
			switch g.Execute(pd) {
			case game.Unknown:
				fmt.Println("I don't understand your command.")
			case game.Unsuccessful:
				fmt.Println("I can't do that yet.")
			}
		}

		// TODO: Handle the light source ticking down.
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
