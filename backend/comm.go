package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
			encodedStats, err := msgpack.Marshal(stats)
			if err != nil {
				log.Println("couldn't format with mspack:", err)
				return
			}

			if err := conn.WriteMessage(websocket.BinaryMessage, encodedStats); err != nil {
				log.Println("couldn't send to websockets:", err)
				return
			}
			time.Sleep(time.Second * 5)
		}
	}()

	// read from frontend
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("couldn't read from websockets:", err)
			break
		}
		log.Println("websocket message:", msg)
		var decoded map[string]interface{}
		_ = msgpack.Unmarshal(msg, &decoded)
		jsonBytes, _ := json.Marshal(decoded)
		fmt.Println(string(jsonBytes))
	}
}
