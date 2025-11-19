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

// comm websockets. handles /api/comm
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
	conn, err := G_WSUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("REST_Comm: Couldn't upgrade to websockets:", err)
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
				log.Println("REST_Comm: Couldn't get CPU usage:", err)
				return
			}

			virtualMem, err := mem.VirtualMemory()
			if err != nil {
				log.Println("REST_Comm: Couldn't get memory usage:", err)
				return
			}

			usageValue, usageUnit := H_HumanReadableSplit(virtualMem.Used)

			// assemble stats into a comm object
			stats := map[string]any{
				"t": "stat.basic",
				"data": map[string]any{
					"memTotal":     H_HumanReadable(virtualMem.Total),
					"memUsage":     math.Round(usageValue),
					"memUsageUnit": usageUnit,
					"cpuUsage":     math.Round(percentages[0]),
				},
			}
			if err := Comm_Send(stats, conn); err != nil {
				log.Println("REST_Comm: Couldn't send stats:", err)
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
			log.Println("REST_Comm: Couldn't read from WebSocket:", err)
			cancel()
			break
		}

		// decode comm object
		var decoded map[string]any
		if err := msgpack.Unmarshal(msg, &decoded); err != nil {
			log.Println("REST_Comm: Could not read data")
			continue
		}

		// pre-assemble our response data
		data := map[string]any{
			"t":  decoded["t"],
			"id": decoded["id"],
		}

		// choose data based on t
		switch decoded["t"] {
		case "config.set":
			// get data
			keys, ok := decoded["data"].(map[string]any)
			if !ok {
				Comm_BR(data, "Data doesn't exist or isn't an object")
				break
			}

			// loop through and set value
			for key, value := range keys {
				if err := C_SetValue(sc.Value, key, value); err != nil {
					Comm_ISE(data, err.Error())
					break
				}
			}

			// save json
			if err := C_Save(sc.Value); err != nil {
				Comm_ISE(data, err.Error())
				break
			}

			// return success
			data["data"] = map[string]any{}
		case "log.read":
			source := IT_Must(decoded, "data.source", "all")
			amount := IT_MustNumber(decoded, "data.amount", uint16(50))
			page := IT_MustNumber(decoded, "data.page", uint16(0))

			// actual read
			events, err := L_Read(source, page*amount, amount)
			if err != nil {
				Comm_ISE(data, err.Error())
				break
			}

			// lovecraftian computing
			thedata := map[string][]map[string]any{}

			for _, e := range events {
				key := e.Time.Format("2006-01-02")
				thedata[key] = append(thedata[key], map[string]any{
					"time":   e.Time.Unix(),
					"source": e.Source,
					"msg":    e.Message,
				})
			}

			data["data"] = thedata
		case "storage.getdevs":
			S_GetDevices()
			data["data"] = 0
		default:
			// if t is not recognized, then throw error
			Comm_Error(data, http.StatusNotFound, "Unknown type")
		}

		if err := Comm_Send(data, conn); err != nil {
			log.Println("REST_Comm: Could not send data for", decoded["t"])
		}
	}
}

// encodes and sends a comm message
func Comm_Send(data map[string]any, connection *websocket.Conn) error {
	// encode
	encodedData, err := msgpack.Marshal(data)
	if err != nil {
		log.Println("Comm_Send: Couldn't format with MessagePack:", err)
		return err
	}

	// send
	if err := connection.WriteMessage(websocket.BinaryMessage, encodedData); err != nil {
		log.Println("Comm_Send: Couldn't send to WebSocket:", err)
		return err
	}
	return nil
}

// prints log message and sets data["error"]
func Comm_Error(data map[string]any, code int, msg string) {
	data["error"] = map[string]any{"code": code, "msg": msg}
	log.Printf("%s: %s.\n", data["t"], msg)
}

// Comm_Error(data, http.StatusInternalServerError, msg)
func Comm_ISE(data map[string]any, msg string) {
	Comm_Error(data, http.StatusInternalServerError, msg)
}

// Comm_Error(data, http.StatusBadRequest, msg)
func Comm_BR(data map[string]any, msg string) {
	Comm_Error(data, http.StatusBadRequest, msg)
}
