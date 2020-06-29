// Package writer contains routines for outputting Scott Adams adventure games
// in a variety of formats.
package writer

import (
	"fmt"
	"io"

	"github.com/chaosotter/golang-adventures/api/scottpb"
)

// WriteTRS80 writes out the game data in the "TRS-80 format" used by ScottFree.
func WriteTRS80(out io.Writer, pb *scottpb.Game) {
	w := &trs80{out, pb}
	w.writeHeader()
	w.writeActions()
	w.writeWords()
	w.writeRooms()
	w.writeMessages()
	w.writeItems()
	w.writeComments()
	w.writeFooter()
}

// trs80 escapulates the Game proto and output writer, just to make for simplier
// method signatures.
type trs80 struct {
	out io.Writer
	pb  *scottpb.Game
}

// writeHeader writes out the game header.
func (t *trs80) writeHeader() {
	h := t.pb.Header
	t.writeIntLn(h.Unknown0)
	t.writeIntLn(h.NumItems - 1)   // adjustment intended
	t.writeIntLn(h.NumActions - 1) // adjustment intended
	t.writeIntLn(h.NumWords - 1)   // adjustment intended
	t.writeIntLn(h.NumRooms - 1)   // adjustment intended
	t.writeIntLn(h.MaxInventory)
	t.writeIntLn(h.StartingRoom)
	t.writeIntLn(h.NumTreasures)
	t.writeIntLn(h.WordLength)
	t.writeIntLn(h.LightDuration)
	t.writeIntLn(h.NumMessages - 1) // adjustment intended
	t.writeIntLn(h.TreasureRoom)
}

// writeActions writes out the actions.
func (t *trs80) writeActions() {
	for i := 0; i < int(t.pb.Header.NumActions); i++ {
		a := t.pb.Actions[i]
		t.writeIntLn(a.VerbIndex*150 + a.NounIndex)
		for j := 0; j < 5; j++ {
			t.writeIntLn(int32(a.Conditions[j].Type) + 20*a.Conditions[j].Value)
		}
		t.writeIntLn(int32(a.Actions[0])*150 + int32(a.Actions[1]))
		t.writeIntLn(int32(a.Actions[2])*150 + int32(a.Actions[3]))
	}
}

// writeWords writes out the words.
func (t *trs80) writeWords() {
	for i := 0; i < int(t.pb.Header.NumWords); i++ {
		t.writeWord(t.pb.Verbs[i])
		t.writeWord(t.pb.Nouns[i])
	}
}

// writeWord writes out a single word.
func (t *trs80) writeWord(w *scottpb.Word) {
	if w.Synonym {
		fmt.Fprintf(t.out, "\"*%s\"\n", w.Word)
	} else {
		fmt.Fprintf(t.out, "\"%s\"\n", w.Word)
	}
}

// writeRooms writes out the rooms.
func (t *trs80) writeRooms() {
	for i := 0; i < int(t.pb.Header.NumRooms); i++ {
		r := t.pb.Rooms[i]
		for j := 0; j < 6; j++ { // north, south, east, west, up, down
			t.writeIntLn(r.Exits[j])
		}
		t.writeRoomDescription(r)
	}
}

// writeRoomDescription writes out a single room description.
func (t *trs80) writeRoomDescription(r *scottpb.Room) {
	if r.Literal {
		fmt.Fprintf(t.out, "\"*%s\"\n", r.Description)
	} else {
		fmt.Fprintf(t.out, "\"%s\"\n", r.Description)
	}
}

// writeMessages writes out the messages.
func (t *trs80) writeMessages() {
	for i := 0; i < int(t.pb.Header.NumMessages); i++ {
		fmt.Fprintf(t.out, "\"%s\"\n", t.pb.Messages[i])
	}
}

// writeItems writes out the items.
func (t *trs80) writeItems() {
	for i := 0; i < int(t.pb.Header.NumItems); i++ {
		it := t.pb.Items[i]
		if it.Autograb != "" {
			fmt.Fprintf(t.out, "\"%s/%s/\" %d \n", it.Description, it.Autograb, it.Location)
		} else {
			fmt.Fprintf(t.out, "\"%s\" %d \n", it.Description, it.Location)
		}
	}
}

// writeComments writes out the action comments.
func (t *trs80) writeComments() {
	for i := 0; i < int(t.pb.Header.NumActions); i++ {
		fmt.Fprintf(t.out, "\"%s\"\n", t.pb.Actions[i].Comment)
	}
}

// writeFooter writes out the game footer.
func (t *trs80) writeFooter() {
	f := t.pb.Footer
	t.writeIntLn(f.Version)
	t.writeIntLn(f.Adventure)
	t.writeIntLn(f.Magic)
}

// writeIntLn writes out an integer value, with a trailing newline.
func (t *trs80) writeIntLn(v int32) {
	if v >= 0 {
		fmt.Fprintf(t.out, " %d \n", v)
	} else {
		fmt.Fprintf(t.out, "%d \n", v)
	}
}
