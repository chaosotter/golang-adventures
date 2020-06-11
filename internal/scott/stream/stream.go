// Package stream provides support for processing the data in a Scott Adams
// adventure file in the ScottFree (TRS-80) format as a stream of tokens.
//
// Such a file consists of a sequence of ASCII-formatted integers (possibly with
// surrounding whitespace) and quote-delimited strings (possibly with internal
// newlines).
//
// We don't pay the slightest bit of attention to Unicode or processing the data
// as runes, since this file format is from the 8-bit days.
package stream

import (
	"bytes"
	"fmt"
	"strconv"
)

// typ represents the type of a token, either Int or String
type typ int

const (
	typeInt = typ(iota)
	typeStr
)

// A token is the basic unit of data we parse from the input.  We do not expose
// this type externally because the file format always allows us to know in
// advance whether the next token should be an integer or a string.
type token struct {
	typ   typ
	value interface{}
}

// Stream is the actual type exposed through this package and contains a fully
// parsed sequences of tokens and a current-position marker.
type Stream struct {
	tokens []token
	next   int
}

// These states are used by the FSM in New for parsing the input data.  The
// individual states are documented inline.
const (
	stateInit = iota
	stateSign
	stateNum
	stateQuote
	stateEscape
)

// New initializes a new Stream from the given game data.  The files are small
// and we never read them partially, so we do all of the parsing up front.
func New(data []byte) (*Stream, error) {
	s := &Stream{}
	st := stateInit
	b := &bytes.Buffer{}

	for o := 0; o < len(data); o++ {
		ch := data[o]
		switch st {
		// Init state: Not currently reading any token.
		case stateInit:
			switch {
			case isSpace(ch):
				// pass
			case ch == '-':
				b.WriteByte(ch)
				st = stateSign
			case isDigit(ch):
				b.WriteByte(ch)
				st = stateNum
			case ch == '"':
				st = stateQuote
			default:
				return nil, fmt.Errorf("Unexpected character '%c' at offset %d (state Init)", ch, o)
			}

		// Sign state: Read the initial '-' of a negative integer.
		case stateSign:
			switch {
			case isDigit(ch):
				b.WriteByte(ch)
				st = stateNum
			default:
				return nil, fmt.Errorf("Unexpected character '%c' at offset %d (state Sign)", ch, o)
			}

		// Num state: Now reading an integer.
		case stateNum:
			switch {
			case isSpace(ch):
				val, err := strconv.Atoi(b.String())
				if err != nil {
					return nil, fmt.Errorf("Internal error converting '%s' to int at offset %d (state Num)", b.String(), o)
				}
				s.tokens = append(s.tokens, token{typeInt, val})
				b.Reset()
				st = stateInit
			case isDigit(ch):
				b.WriteByte(ch)
			default:
				return nil, fmt.Errorf("Unexpected character '%c' at offset %d (state Num)", ch, o)
			}

		// Quote state: Read the initial '"' of a string.
		case stateQuote:
			switch {
			//case ch == '\\':
			//	st = stateEscape
			case ch == '"':
				s.tokens = append(s.tokens, token{typeStr, b.String()})
				b.Reset()
				st = stateInit
			default:
				b.WriteByte(ch)
			}

		// Escape state: Read the next character in a string unconditionally.
		case stateEscape:
			b.WriteByte(ch)

		default:
			return nil, fmt.Errorf("Internal error: unknown state %d", st)
		}
	}

	return s, nil
}

// isDigit checks if this byte is a digit in the traditional ASCII code.
func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

// isSpace checks if this byte is whitespace in the traditional ASCII code.  We
// ignore '\v' because seriously, nobody ever used that.
func isSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

// Done checks if we're at the end of the stream.
func (s *Stream) Done() bool {
	return s.next >= len(s.tokens)
}

// NextInt returns the next integer in the stream.
func (s *Stream) NextInt() (int, error) {
	switch t, err := s.nextToken(); {
	case err != nil:
		return 0, err
	case t.typ != typeInt:
		return 0, fmt.Errorf("Next token was a string, expected an integer")
	default:
		return t.value.(int), nil
	}
}

// NextString returns the next string in the stream.
func (s *Stream) NextString() (string, error) {
	switch t, err := s.nextToken(); {
	case err != nil:
		return "", err
	case t.typ != typeStr:
		return "", fmt.Errorf("Next token was an integer, expected a string")
	default:
		return t.value.(string), nil
	}
}

// nextToken returns the next token (for internal use).
func (s *Stream) nextToken() (token, error) {
	if s.Done() {
		return token{}, fmt.Errorf("Premature end of stream")
	}
	t := s.tokens[s.next]
	s.next++
	return t, nil
}
