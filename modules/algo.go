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

// Algorithm used - Modified Priority Round Robin (PRR) with Dynamic Quantum Times
// 1. Find the direction with the highest vehicle density - Priority {North, South, East, West}
// 2. Set the traffic light to green for the direction with the highest density, and red for the others
// 3. Allow vehicles to pass for a certain count of vehicles (MaxVehicleCount * Avg Time to cross the junction)
// 4. Exclude this direction from the next iteration
// 5. Repeat the process until all directions have been given a chance to pass (One cycle)
// 6. Repeat the process for the next cycle
// Quantum (Aka Threshold) - Dynamic based on Traffic History and Predictions

var (
	RandomGenMode = false
)

type Vehicle struct {
	Direction string `json:"direction"`
}

type Junction struct {
	Vehicles  map[string]int `json:"vehicles"`
	JnID      int            `json:"id"`
	Location  [2]float64     `json:"location"`
	Threshold int            `json:"threshold"`

	mutex sync.Mutex
}

func NewJunction(id int) *Junction {
	return &Junction{
		Vehicles:  map[string]int{"north": 0, "south": 0, "east": 0, "west": 0},
		Threshold: Threshold,
		JnID:      id,
	}
}

func (j *Junction) AddVehicle(direction string) {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	j.Vehicles[direction]++
}

func (j *Junction) FindMaxDensityDirection(excluded map[string]bool) string {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	highestDensity := 0
	highestDensityDirection := ""

	for direction, count := range j.Vehicles {
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

	if j.Vehicles[direction] > timeEffective {
		j.Vehicles[direction] -= timeEffective
	} else {
		j.Vehicles[direction] = 0
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
			if j.Vehicles[greenDirection] > MaxVehicleCount {
				vehiclesToActivate = MaxVehicleCount
				j.Vehicles[greenDirection] -= MaxVehicleCount
			} else {
				vehiclesToActivate = j.Vehicles[greenDirection]
				j.Vehicles[greenDirection] = 0
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
	jn        = NewJunction(0)
	Junctions = []*Junction{jn}
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
			jn.Vehicles[direction] = count
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

	for direction := range jn.Vehicles {
		jn.Vehicles[direction] = 0
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

// Dummy Junctions for testing
func init() {
	for i := 1; i <= 10; i++ {
		Junctions = append(Junctions, NewJunction(i))
	}

	for _, jn := range Junctions[1:] {
		for direction := range jn.Vehicles {
			jn.Vehicles[direction] = rand.Intn(20)
		}
	}
}
