// summit backend/comm.go - handles communication (server-side)

package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/vmihailenco/msgpack/v5"
)

const Comm_ReadLimit int64 = 2048

var Comm_ConnMu sync.Mutex

type (
	Comm_ErrorT struct {
		Code int    `msgpack:"code"`
		Msg  string `msgpack:"msg"`
	}

	Comm_Message struct {
		T     string      `msgpack:"t"`
		ID    uint32      `msgpack:"id"`
		Data  any         `msgpack:"data"`
		Error Comm_ErrorT `msgpack:"error"`
	}

	Comm_StatsMsg struct {
		MemTotal uint64 `msgpack:"mem_total"`
		MemUsed  uint64 `msgpack:"mem_used"`
		CpuUsage uint8  `msgpack:"cpu_usage"`
	}
)

var Comm_Handlers = map[string]func(Comm_Message, string) (any, error){
	"config.set":      Comm_ConfigSet,
	"log.read":        Comm_LogRead,
	"storage.getdevs": Comm_StorageGetdevs,
	"net.getnics":     Comm_NetGetnics,
}

var Comm_Events = map[string]func(ctx context.Context, conn *websocket.Conn, id uint32){
	"stat.basic": Comm_StatsTimer,
	"net.stats":  Comm_NetStats,
}

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

	conn.SetReadLimit(Comm_ReadLimit)
	conn.SetReadDeadline(time.Now().Add(A_SessionExpireTime))

	ctx, cancel := context.WithCancel(context.Background())

	// read from frontend
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("REST_Comm: Couldn't read from WebSocket:", err)
			cancel()
			break
		}

		// decode comm object
		var decoded Comm_Message
		if err := msgpack.Unmarshal(msg, &decoded); err != nil {
			log.Println("REST_Comm: Could not read data")
			continue
		}

		if decoded.T == "subscribe" {
			request, ok := decoded.Data.(map[string]any)
			if !ok {
				continue
			}

			requested := IT_Must(request, "t", "")

			if handler, ok := Comm_Events[requested]; !ok {
				log.Println("subscribe error: Unknown type \"" + requested + "\"")
			} else {
				go handler(ctx, conn, decoded.ID)
			}

			continue
		}

		// pre-assemble our response data
		data := Comm_Message{
			T:  decoded.T,
			ID: decoded.ID,
		}

		if handler, ok := Comm_Handlers[decoded.T]; !ok {
			Comm_Error(&data, http.StatusNotFound, "Unknown type")
		} else {
			answer, err := handler(decoded, sc.Value)
			if err != nil {
				Comm_ISE(&data, err.Error())
			} else {
				data.Data = answer
			}
		}

		if err := Comm_Send(data, conn); err != nil {
			log.Println("REST_Comm: Could not send data for", data.T)
		}
	}
}

// encodes and sends a comm message
func Comm_Send(data Comm_Message, connection *websocket.Conn) error {
	// encode
	encodedData, err := msgpack.Marshal(data)
	if err != nil {
		log.Printf("Comm_Send: %v: Couldn't format with MessagePack: %s", data.T, err)
		return fmt.Errorf("couldn't format with MessagePack: %s", err)
	}

	Comm_ConnMu.Lock()
	defer Comm_ConnMu.Unlock()

	// send
	if err := connection.WriteMessage(websocket.BinaryMessage, encodedData); err != nil {
		return fmt.Errorf("couldn't send to WebSocket: %s", err)
	}

	return nil
}

// prints log message and sets data["error"]
func Comm_Error(data *Comm_Message, code int, msg string) {
	data.Error = Comm_ErrorT{
		Code: code,
		Msg:  msg,
	}
	log.Printf("%s: error %d: %s.\n", data.T, code, msg)
}

// Comm_Error(data, http.StatusInternalServerError, msg)
func Comm_ISE(data *Comm_Message, msg string) {
	Comm_Error(data, http.StatusInternalServerError, msg)
}

// Comm_Error(data, http.StatusBadRequest, msg)
func Comm_BR(data *Comm_Message, msg string) {
	Comm_Error(data, http.StatusBadRequest, msg)
}

func Comm_StatsTimer(ctx context.Context, conn *websocket.Conn, id uint32) {
	// 0s
	if err := Comm_SendStats(conn, id); err != nil {
		log.Println("Comm_StatsTimer: Couldn't send stats:", err)
		return
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := Comm_SendStats(conn, id); err != nil {
				log.Println("Comm_StatsTimer: Couldn't send stats:", err)
				return
			}
		}
	}
}

func Comm_SendStats(conn *websocket.Conn, id uint32) error {
	percentages, err := cpu.Percent(0, false)
	if err != nil {
		return err
	}

	virtualMem, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	stats := Comm_Message{
		ID: id,
		T:  "stat.basic",
		Data: Comm_StatsMsg{
			MemTotal: virtualMem.Total,
			MemUsed:  virtualMem.Used,
			CpuUsage: uint8(math.Round(percentages[0])),
		},
	}

	return Comm_Send(stats, conn)
}
