package mock

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/mshindle/structures"
)

type Empty struct{}

const AlertFrequency = 1000

var GreekLetters = []string{
	"alpha", "beta", "gamma", "delta", "epsilon", "zeta", "eta", "theta", "iota", "kappa",
	"lambda", "mu", "nu", "xi", "omicron", "pi", "rho", "sigma", "tau", "upsilon",
	"phi", "chi", "psi", "omega",
}

type Simulation struct {
	URL           string
	TotalMessages int
	WorkerCount   int
	DroneCount    int
}

type drone struct {
	id              string
	uptime          int
	battery         int
	latitude        float32
	longitude       float32
	altitude        float32
	speed           float32
	heading         int
	eventCount      int
	alertEventCount int
}

func createDrones(droneCount int) []*drone {
	drones := make([]*drone, droneCount)
	usedIDs := structures.NewSet[string](droneCount)

	for i := range droneCount {
		var id string
		for {
			id = fmt.Sprintf("%s%02d", GreekLetters[rand.IntN(len(GreekLetters))], rand.IntN(100))
			// .AddIfUnique() handles the RLock check, the Lock escalation, and to double-check internally.
			if usedIDs.AddIfUnique(id) {
				break
			}
		}

		drones[i] = new(drone{id: id, battery: 100, uptime: 1})
	}
	return drones
}

func (s *Simulation) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	msgChan := make(chan any, s.WorkerCount)
	sentCount := new(int64)

	workerPool := sync.Pool{
		New: func() any {
			return &http.Client{}
		},
	}

	// Create workers
	for range s.WorkerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := workerPool.Get().(*http.Client)
			defer workerPool.Put(client)
			for {
				select {
				case <-ctx.Done():
					return
				case msg, ok := <-msgChan:
					if !ok {
						return
					}
					s.send(ctx, client, msg)
					atomic.AddInt64(sentCount, 1)
				}
			}
		}()
	}

	drones := createDrones(s.DroneCount)

	// Simulation loop
	go func() {
		defer close(msgChan)
		for {
			if atomic.LoadInt64(sentCount) >= int64(s.TotalMessages) {
				return
			}

			// Pick a random drone to report
			d := drones[rand.IntN(len(drones))]

			// Decide what to send
			// telemetry, position, or alert

			// Alert event should never be thrown more than 1 in 1000 events.
			// telemetry and position are sent as they move.

			var msg any
			r := rand.IntN(AlertFrequency)
			if r == 0 && d.alertEventCount*AlertFrequency < d.eventCount {
				msg = d.generateAlert()
				d.alertEventCount++
			} else if rand.IntN(2) == 0 {
				msg = d.generateTelemetry()
			} else {
				msg = d.generatePosition()
			}

			d.eventCount++

			select {
			case <-ctx.Done():
				return
			case msgChan <- msg:
			}
		}
	}()

	wg.Wait()
	return nil
}

func (d *drone) generateAlert() any {
	return struct {
		Type string `json:"-"`
		Data any
	}{
		Type: "alert",
		Data: struct {
			DroneID     string `json:"drone_id"`
			FaultCode   int    `json:"fault_code"`
			Description string `json:"description"`
		}{
			DroneID:     d.id,
			FaultCode:   rand.IntN(10),
			Description: "Simulated fault",
		},
	}
}

func (d *drone) generateTelemetry() any {
	d.uptime++
	if d.battery > 0 && rand.IntN(10) == 0 {
		d.battery--
	}
	return struct {
		Type string `json:"-"`
		Data any
	}{
		Type: "telemetry",
		Data: struct {
			DroneID          string `json:"drone_id"`
			RemainingBattery int    `json:"battery"`
			Uptime           int    `json:"uptime"`
			CoreTemp         int    `json:"core_temp"`
		}{
			DroneID:          d.id,
			RemainingBattery: d.battery,
			Uptime:           d.uptime,
			CoreTemp:         20 + rand.IntN(40),
		},
	}
}

func (d *drone) generatePosition() any {
	d.latitude += rand.Float32() * 0.001
	d.longitude += rand.Float32() * 0.001
	d.altitude = 100 + rand.Float32()*10
	d.speed = 10 + rand.Float32()*5
	d.heading = (d.heading + rand.IntN(10)) % 360

	return struct {
		Type string `json:"-"`
		Data any
	}{
		Type: "position",
		Data: struct {
			DroneID         string  `json:"drone_id"`
			Latitude        float32 `json:"latitude"`
			Longitude       float32 `json:"longitude"`
			Altitude        float32 `json:"altitude"`
			CurrentSpeed    float32 `json:"current_speed"`
			HeadingCardinal int     `json:"heading_cardinal"`
		}{
			DroneID:         d.id,
			Latitude:        d.latitude,
			Longitude:       d.longitude,
			Altitude:        d.altitude,
			CurrentSpeed:    d.speed,
			HeadingCardinal: d.heading,
		},
	}
}

func (s *Simulation) send(ctx context.Context, client *http.Client, msg any) {
	m := msg.(struct {
		Type string `json:"-"`
		Data any
	})

	endpoint := ""
	switch m.Type {
	case "alert":
		endpoint = "/api/cmds/alert"
	case "telemetry":
		endpoint = "/api/cmds/telemetry"
	case "position":
		endpoint = "/api/cmds/position"
	}

	body, _ := json.Marshal(m.Data)
	req, _ := http.NewRequestWithContext(ctx, "POST", s.URL+endpoint, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending %s: %v\n", m.Type, err)
		return
	}
	defer resp.Body.Close()
}
