package config

import (
	"strings"
)

// Event is a configuration event.
type Event int

const (
	// GET Get key event
	GET Event = iota

	// SET Set key event
	SET

	// TOINT Convert to int event
	TOINT

	// TOSTRING Convert to string event
	TOSTRING

	// TOBOOL Convert to bool event
	TOBOOL

	// TOTIME Convert to time event
	TOTIME

	// SWAP Swap handler event
	SWAP

	// INIT Initialize handler event
	INIT

	// UNKNOWN Unknown event
	UNKNOWN
)

// tostr maps events to their string representation.
var tostr = map[Event]string{
	GET:      "GET",
	SET:      "SET",
	TOINT:    "TOINT",
	TOSTRING: "TOSTRING",
	TOBOOL:   "TOBOOL",
	TOTIME:   "TOTIME",
	SWAP:     "SWAP",
	INIT:     "INIT",
}

// String turns an event into a string.
func (e Event) String() string {
	if val, found := tostr[e]; found {
		return val
	}

	return "UNKNOWN"
}

// toevent maps strings to events
var toevent = map[string]Event{
	"GET":      GET,
	"SET":      SET,
	"TOINT":    TOINT,
	"TOSTRING": TOSTRING,
	"TOBOOL":   TOBOOL,
	"TOTIME":   TOTIME,
	"SWAP":     SWAP,
	"INIT":     INIT,
	"UNKNOWN":  UNKNOWN,
}

// ToEvent maps a string to an event. The match on string is not
// case sensitive.
func ToEvent(str string) Event {
	if val, found := toevent[strings.ToUpper(str)]; found {
		return val
	}

	return UNKNOWN
}
