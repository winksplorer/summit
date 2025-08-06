// summit backend/comm.go - handles communication (server-side)

package main

import (
	"context"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/vmihailenco/msgpack/v5"
)

func REST_Comm(w http.ResponseWriter, r *http.Request) {
	// if not authed, then close connection
	if !A_Authenticated(w, r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sc, err := r.Cookie("s")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// upgrade to ws
	conn, err := WS_Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("couldn't upgrade to websockets:", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())

	// send to frontend
	go func(ctx context.Context) {
		for {
			// calculate stats
			percentages, err := cpu.Percent(0, false)
			if err != nil {
				log.Println("couldn't get cpu info:", err)
				return
			}

			virtualMem, err := mem.VirtualMemory()
			if err != nil {
				log.Println("couldn't get memory info:", err)
				return
			}

			usageValue, usageUnit := H_HumanReadableSplit(virtualMem.Used)

			// assemble stats into a comm object
			stats := map[string]interface{}{
				"t": "stat.basic",
				"data": map[string]interface{}{
					"memTotal":     H_HumanReadable(virtualMem.Total),
					"memUsage":     math.Round(usageValue),
					"memUsageUnit": usageUnit,
					"cpuUsage":     math.Round(percentages[0]),
				},
			}
			if err := Comm_Send(stats, conn); err != nil {
				log.Println("couldn't send stats:", err)
				return
			}

			// wait for next round
			delay := time.NewTimer(5 * time.Second)

			select {
			case <-ctx.Done():
				if !delay.Stop() {
					<-delay.C
				}
				return

			case <-delay.C:
				// loop back and do again
			}
		}
	}(ctx)

	// read from frontend
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("couldn't read from websockets:", err)
			cancel()
			break
		}

		// decode comm object
		var decoded map[string]interface{}
		if err := msgpack.Unmarshal(msg, &decoded); err != nil {
			log.Println("could not read data")
			continue
		}

		// pre-assemble our response data
		data := map[string]interface{}{
			"t":  decoded["t"],
			"id": decoded["id"],
		}

		// choose data based on t
		switch decoded["t"] {
		case "config.set":
			// get data
			keys, ok := decoded["data"].(map[string]interface{})
			if !ok {
				Comm_Error(data, "config.set", http.StatusBadRequest, "data doesn't exist or isn't an object")
				break
			}

			// loop through and set value
			for key, value := range keys {
				if err := C_SetValue(sc.Value, key, value); err != nil {
					Comm_Error(data, "config.set", http.StatusInternalServerError, err.Error())
					break
				}
			}

			// save json
			if err := C_Save(sc.Value); err != nil {
				Comm_Error(data, "config.set", http.StatusInternalServerError, err.Error())
				break
			}

			// return success
			data["data"] = map[string]interface{}{}
		default:
			// if t is not recognized, then throw error
			data["error"] = map[string]interface{}{"code": http.StatusNotFound, "msg": "unknown type"}
		}

		if err := Comm_Send(data, conn); err != nil {
			log.Println("could not send data for", decoded["t"])
		}
	}
}

func Comm_Send(data map[string]interface{}, connection *websocket.Conn) error {
	// encode
	encodedData, err := msgpack.Marshal(data)
	if err != nil {
		log.Println("couldn't format with msgpack:", err)
		return err
	}

	// send
	if err := connection.WriteMessage(websocket.BinaryMessage, encodedData); err != nil {
		log.Println("couldn't send to websockets:", err)
		return err
	}
	return nil
}

func Comm_Error(data map[string]interface{}, t string, code int, msg string) {
	data["error"] = map[string]interface{}{"code": code, "msg": msg}
	log.Printf("%s failed: %s\n", t, msg)
}
