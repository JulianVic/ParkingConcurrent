package application

import (
	"math/rand"
	"sync"
	"time"
	"github.com/JulianVic/SimuladorEstacionamiento/domain"
)

type Simulator struct {
	parkingService *domain.ParkingService
}

func NewSimulator(parkingService *domain.ParkingService) *Simulator {
	return &Simulator{
		parkingService: parkingService,
	}
}

func (s *Simulator) SimulateVehicle(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	vehicle := &domain.Vehicle{ID: id}

	var space int
	var success bool
	for {
		space, success = s.parkingService.VehicleEnter(vehicle)
		if success {
			break
		}
		time.Sleep(time.Millisecond * 50)
	}

	parkingTime := 1 + rand.Intn(2)
	time.Sleep(time.Duration(parkingTime) * time.Second)

	for {
		if s.parkingService.VehicleExit(vehicle, space) {
			break
		}
		time.Sleep(time.Millisecond * 50)
	}
}
