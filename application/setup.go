// application/setup.go
package application

import (
	"math/rand"
	"sync"
	"time"
	
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	
	"github.com/JulianVic/SimuladorEstacionamiento/presentation"
)

func SetupStartButton(gui *presentation.ParkingGUI, simulator *Simulator) {
	buttonContainer := gui.Window.Content().(*fyne.Container).Objects[0].(*fyne.Container).Objects[7].(*fyne.Container)
	startBtn := buttonContainer.Objects[1].(*widget.Button)
	
	startBtn.OnTapped = func() {
		startBtn.Disable()
		startBtn.SetText("Simulación en curso...")

		var wg sync.WaitGroup

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go simulator.SimulateVehicle(i+1, &wg)
			time.Sleep(time.Duration(rand.ExpFloat64() * float64(time.Millisecond) * 500))
		}

		go func() {
			wg.Wait()
			startBtn.Enable()
			startBtn.SetText("Iniciar Simulación")
			simulator.parkingService.UpdateStats("Simulación completada")
		}()
	}
}