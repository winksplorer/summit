package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/vmihailenco/msgpack/v5"
)

func commHandler(w http.ResponseWriter, r *http.Request) {
	if !authenticated(w, r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// upgrade to ws
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("couldn't upgrade to websockets:", err)
		return
	}
	defer conn.Close()

	// send to frontend
	go func() {
		socketErrCount := 1
		for {
			// stats
			percentages, err := cpu.Percent(0, false)
			if err != nil {
				log.Println("couldn't get cpu info:", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			virtualMem, err := mem.VirtualMemory()
			if err != nil {
				log.Println("couldn't get memory info:", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			var usageValue float64
			var usageUnit string
			const (
				kb = 1024
				mb = kb * 1024
				gb = mb * 1024
			)

			switch {
			case virtualMem.Used >= gb:
				usageValue = float64(virtualMem.Used) / float64(gb)
				usageUnit = "g"
			case virtualMem.Used >= mb:
				usageValue = float64(virtualMem.Used) / float64(mb)
				usageUnit = "m"
			case virtualMem.Used >= kb:
				usageValue = float64(virtualMem.Used) / float64(kb)
				usageUnit = "k"
			default:
				usageValue = float64(virtualMem.Used)
				usageUnit = ""
			}

			stats := map[string]interface{}{
				"t": "stat.basic",
				"data": map[string]interface{}{
					"memTotal":     humanReadable(virtualMem.Total),
					"memUsage":     fmt.Sprintf("%.1f", usageValue),
					"memUsageUnit": usageUnit,
					"cpuUsage":     float64(int(percentages[0]*100)) / 100,
				},
			}
			if err := commSend(stats, conn); err != nil {
				log.Printf("comm: %d/2 websocket send errors until terminating\n", socketErrCount)
				if strings.Contains(err.Error(), "websocket:") {
					socketErrCount++
				}
			}

			if socketErrCount > 2 {
				break
			}
			time.Sleep(time.Second * 2)
		}
	}()

	// read from frontend
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("couldn't read from websockets:", err)
			break
		}

		var decoded map[string]interface{}
		_ = msgpack.Unmarshal(msg, &decoded)

		data := map[string]interface{}{
			"t":    decoded["t"],
			"data": nil,
		}

		switch decoded["t"] {
		case "info.hostname":
			data["data"] = map[string]interface{}{
				"hostname": hostname,
			}
		}

		if err := commSend(data, conn); err != nil {
			log.Panicln("could not send data for", decoded["t"])
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
