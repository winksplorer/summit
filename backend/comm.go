package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/vmihailenco/msgpack/v5"
)

func commHandler(w http.ResponseWriter, r *http.Request) {
	// upgrade to ws
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("couldn't upgrade to websockets:", err)
		return
	}
	defer conn.Close()

	// handle auth
	if !authenticated(w, r) {
		data := map[string]interface{}{
			"t": "auth.status",
			"data": map[string]interface{}{
				"authed": false,
			},
		}

		if err := commSend(data, conn); err != nil {
			log.Println("couldn't send auth reject message:", err)
		}

		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	// send to frontend
	go func(ctx context.Context) {
		for {
			// ---- do the work first ----
			percentages, err := cpu.Percent(0, false)
			if err != nil {
				log.Println("couldn't get cpu info:", err)
			}

			virtualMem, err := mem.VirtualMemory()
			if err != nil {
				log.Println("couldn't get memory info:", err)
			}

			usageValue, usageUnit := humanReadableSplit(virtualMem.Used)

			stats := map[string]interface{}{
				"t": "stat.basic",
				"data": map[string]interface{}{
					"memTotal":     humanReadable(virtualMem.Total),
					"memUsage":     math.Round(usageValue),
					"memUsageUnit": usageUnit,
					"cpuUsage":     math.Round(percentages[0]),
				},
			}
			if err := commSend(stats, conn); err != nil {
				log.Println("couldn't send stats:", err)
			}

			// ---- wait for next round ----
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
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Println("comm: websocket closed, terminating now")
			}
			cancel()
			break
		}

		var decoded map[string]interface{}
		if err := msgpack.Unmarshal(msg, &decoded); err != nil {
			log.Println("could not read data")
			continue
		}

		data := map[string]interface{}{
			"t":  decoded["t"],
			"id": decoded["id"],
		}

		switch decoded["t"] {
		case "info.hostname":
			data["data"] = map[string]interface{}{
				"hostname": hostname,
			}
		case "info.buildString":
			data["data"] = map[string]interface{}{
				"buildString": fmt.Sprintf("summit v%s (built on %s)", Version, BuildDate),
			}
		case "info.pages":
			data["data"] = []string{
				"terminal", "logging", "storage", "networking",
				"containers", "services", "updates", "settings",
			}
		default:
			data["error"] = map[string]interface{}{
				"code": 404,
				"msg":  "unknown type",
			}
		}

		if err := commSend(data, conn); err != nil {
			log.Println("could not send data for", decoded["t"])
		}
	}
}

func commSend(data map[string]interface{}, connection *websocket.Conn) error {
	encodedData, err := msgpack.Marshal(data)
	if err != nil {
		log.Println("couldn't format with mspack:", err)
		return err
	}

	if err := connection.WriteMessage(websocket.BinaryMessage, encodedData); err != nil {
		log.Println("couldn't send to websockets:", err)
		return err
	}
	return nil
}
