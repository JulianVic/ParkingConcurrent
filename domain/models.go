package domain

import (
	"sync"
)

type Vehicle struct {
	ID int
}

type ParkingLot struct {
	Spaces          [20]bool
	Entrance        chan bool
	SpacesAvailable chan bool
	Mu              sync.Mutex
	Direction       int
	UpdateChan      chan string
	VehiclesExited  int
}