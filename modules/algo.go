package modules

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	MaxVehicleCount  = 20
	MaxEffectiveTime = 10
	Threshold        = 10 // TODO
)

var (
	RandomGenMode = false
)

type Vehicle struct {
	Direction string `json:"direction"`
}

type Junction struct {
	vehicles map[string]int
	mutex    sync.Mutex
}

func NewJunction() *Junction {
	return &Junction{
		vehicles: map[string]int{"north": 0, "south": 0, "east": 0, "west": 0},
	}
}

func (j *Junction) AddVehicle(direction string) {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	j.vehicles[direction]++
}

func (j *Junction) FindMaxDensityDirection(excluded map[string]bool) string {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	highestDensity := 0
	highestDensityDirection := ""

	for direction, count := range j.vehicles {
		if !excluded[direction] && count > highestDensity {
			highestDensity = count
			highestDensityDirection = direction
		}
	}

	if highestDensity == 0 {
		return ""
	}

	return highestDensityDirection
}

func (j *Junction) RemoveVehicles(direction string, timeEffective int) {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	if j.vehicles[direction] > timeEffective {
		j.vehicles[direction] -= timeEffective
	} else {
		j.vehicles[direction] = 0
	}
}

func (j *Junction) SwitchLights() {
	for {
		excluded := make(map[string]bool)
		for i := 0; i < 4; i++ {
			greenDirection := j.FindMaxDensityDirection(excluded)
			if greenDirection == "" {
				time.Sleep(time.Second * 1)
				break
			}

			excluded[greenDirection] = true
			j.mutex.Lock()

			var vehiclesToActivate int
			if j.vehicles[greenDirection] > MaxVehicleCount {
				vehiclesToActivate = MaxVehicleCount
				j.vehicles[greenDirection] -= MaxVehicleCount
			} else {
				vehiclesToActivate = j.vehicles[greenDirection]
				j.vehicles[greenDirection] = 0
			}

			effectiveTime := 3
			avgSpeed := 3.0

			effectiveTime += int(float64(vehiclesToActivate) / avgSpeed)
			if effectiveTime > MaxEffectiveTime {
				effectiveTime = MaxEffectiveTime
			}
			j.mutex.Unlock()

			event := TrafficEvent{
				Type:          "lightChange",
				Direction:     greenDirection,
				Color:         "green",
				TimeEffective: int32(effectiveTime),
				CountAllowed:  vehiclesToActivate,
			}
			broadcast <- event

			time.Sleep(time.Duration(effectiveTime) * time.Second)
			broadcast <- TrafficEvent{
				Type:      "lightChange",
				Direction: greenDirection,
				Color:     "red",
			}
		}
	}
}

type TrafficEvent struct {
	Type          string `json:"type"`            // "vehicleUpdate" or "lightChange"
	Direction     string `json:"direction"`       // "north", "south", "east", "west"
	Color         string `json:"color,omitempty"` // "red", "green"
	TimeEffective int32  `json:"timeEffective,omitempty"`
	CountAllowed  int    `json:"countAllowed,omitempty"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan TrafficEvent)
	mutex     sync.Mutex
	jn        = NewJunction()
)

func HandleWSConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket Upgrade Error:", err)
		return
	}
	defer conn.Close()

	mutex.Lock()
	clients[conn] = true
	mutex.Unlock()

	log.Println("New WebSocket client connected")

	var resp struct {
		Phases []int `json:"phases"`
	}

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket disconnected:", err)
			mutex.Lock()
			delete(clients, conn)
			mutex.Unlock()
			break
		}
		err = json.Unmarshal(p, &resp)
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		jn.mutex.Lock()
		for i, count := range resp.Phases {
			direction := []string{"north", "south", "east", "west"}[i]
			jn.vehicles[direction] = count
		}
		jn.mutex.Unlock()
	}
}

func BroadcastTrafficUpdates() {
	for {
		event := <-broadcast
		message, err := json.Marshal(event)
		if err != nil {
			log.Println("JSON Encoding Error:", err)
			continue
		}

		mutex.Lock()
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println("WebSocket Write Error:", err)
				client.Close()
				delete(clients, client)
			}
		}
		mutex.Unlock()
	}
}

func StartTrafficSimulation() {
	go jn.SwitchLights()
	go func() {
		for {
			if !RandomGenMode {
				time.Sleep(time.Second * 3)
				continue
			}
			vh := genRandomVehicle()
			broadcast <- TrafficEvent{Type: "vehicleUpdate", Direction: vh.Direction}
			jn.AddVehicle(vh.Direction)
			time.Sleep(time.Millisecond * 400)
		}
	}()
}

func genRandomVehicle() Vehicle {
	directions := []string{"north", "south", "east", "west"}
	return Vehicle{Direction: directions[rand.Intn(4)]}
}

func ClearVehiclesHandler(w http.ResponseWriter, r *http.Request) {
	jn.mutex.Lock()
	defer jn.mutex.Unlock()

	for direction := range jn.vehicles {
		jn.vehicles[direction] = 0
	}
}

func HandleRandomToggle(w http.ResponseWriter, r *http.Request) {
	enabled := r.FormValue("enabled")
	if enabled == "true" {
		RandomGenMode = true
	} else {
		RandomGenMode = false
	}
}

func HandleAutoToggle(w http.ResponseWriter, r *http.Request) {
	enabled := r.FormValue("enabled")
	fmt.Println("TODO: Implement auto toggle", enabled)
}

func SpawnRequestHandler(w http.ResponseWriter, r *http.Request) {
	var req []int

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for i, count := range req {
		direction := []string{"north", "south", "east", "west"}[i]
		for j := 0; j < count; j++ {
			jn.AddVehicle(direction)
			broadcast <- TrafficEvent{Type: "vehicleUpdate", Direction: direction}
		}
	}
}
