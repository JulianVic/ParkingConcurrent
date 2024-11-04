package main

import (
	"math/rand"
	"time"
	"github.com/JulianVic/SimuladorEstacionamiento/application"
	"github.com/JulianVic/SimuladorEstacionamiento/domain"
	"github.com/JulianVic/SimuladorEstacionamiento/presentation"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	
	lot := &domain.ParkingLot{
		Entrance:        make(chan bool, 1),
		SpacesAvailable: make(chan bool, 20),
		Direction:       0,
		UpdateChan:      make(chan string, 100),
	}
	
	for i := 0; i < 20; i++ {
		lot.SpacesAvailable <- true
	}

	parkingService := domain.NewParkingService(lot)
	gui := presentation.CreateGUI()
	guiService := presentation.NewGUIService(gui)
	simulator := application.NewSimulator(parkingService)

	// Configurar el manejo de actualizaciones de la GUI
	go func() {
		for text := range lot.UpdateChan {
			gui.Stats.SetText(text)
			gui.Window.Canvas().Refresh(gui.Stats)
		}
	}()

	// Configurar las actualizaciones visuales del estacionamiento
	go func() {
		for {
			lot.Mu.Lock()
			for i, occupied := range lot.Spaces {
				guiService.UpdateParkingSpace(i, occupied, gui.CarImages[rand.Intn(len(gui.CarImages))])
			}
			guiService.UpdateEntranceColor(lot.Direction)
			lot.Mu.Unlock()
			time.Sleep(time.Millisecond * 100)
		}
	}()

	application.SetupStartButton(gui, simulator)
	gui.Window.ShowAndRun()
}