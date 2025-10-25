package main

import (
	"fmt"
	"math/rand"
	"time"
)

// maximum amount of events that can be read at once
const L_MaxReadAmount = 200

// an event
type L_Event struct {
	Time    time.Time
	Source  string
	Message string
}

// reads `amount` events from `source`, starting from `offset`
func L_Read(source string, offset uint16, amount uint16) ([]L_Event, error) {
	if amount > L_MaxReadAmount {
		return nil, fmt.Errorf("amount too high (%d > %d)", amount, L_MaxReadAmount)
	} else if amount == 0 {
		return nil, nil
	}

	var events []L_Event = []L_Event{}

	switch source {
	case "test":
		for i := 0; i < int(amount); i++ {
			n := 5 + rand.Intn(146) // length 5–150
			b := make([]byte, n)
			for j := range b {
				b[j] = byte('a' + (i+j)%26)
			}
			events = append(events, L_Event{
				Time:    time.Unix(rand.Int63n(1<<31), 0),
				Source:  source,
				Message: string(b),
			})
		}
	}

	return events, nil
}
