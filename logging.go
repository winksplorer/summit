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
	Time    int64  `msgpack:"time"`
	Source  string `msgpack:"source"`
	Message string `msgpack:"msg"`
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
				Time:    rand.Int63n(1 << 31),
				Source:  source,
				Message: string(b),
			})
		}
	}

	return events, nil
}

func Comm_LogRead(data Comm_Message, keyCookie string) (any, error) {
	request, ok := data.Data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("data doesn't exist or isn't an object")
	}

	source := IT_Must(request, "source", "all")
	amount := IT_MustNumber(request, "amount", uint16(50))
	page := IT_MustNumber(request, "page", uint16(0))

	// actual read
	events, err := L_Read(source, page*amount, amount)
	if err != nil {
		return nil, err
	}

	thedata := map[string][]L_Event{}

	for _, e := range events {
		key := time.Unix(e.Time, 0).Format("2006-01-02")
		thedata[key] = append(thedata[key], e)
	}

	return thedata, nil
}
