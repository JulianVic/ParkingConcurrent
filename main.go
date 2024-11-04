// main.go
package main

import (
    "fmt"
    "image/color"
    "math/rand"
    "sync"
    "time"
    
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/canvas"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/layout"
    "fyne.io/fyne/v2/widget"
)

type ParkingLot struct {
    spaces          [20]bool
    entrance        chan bool
    spacesAvailable chan bool
    mu             sync.Mutex
    direction      int
    gui            *ParkingGUI
    updateChan     chan string
    vehiclesExited int
}

type ParkingGUI struct {
    spaces    [20]*canvas.Rectangle
    entrance  *canvas.Rectangle
    stats     *widget.Label
    window    fyne.Window
}

type Vehicle struct {
    id int
}

func NewParkingLot() *ParkingLot {
    return &ParkingLot{
        entrance:        make(chan bool, 1),
        spacesAvailable: make(chan bool, 20),
        direction:       0,
        updateChan:      make(chan string, 100),
    }
}

func createGUI(parking *ParkingLot) *ParkingGUI {
    myApp := app.New()
    window := myApp.NewWindow("Simulador de Estacionamiento")
    
    // Crear contenedor principal
    content := container.NewGridWithRows(3)
    
    // Panel superior para estadísticas
    stats := widget.NewLabel("Vehículos en espera: 0")
    statsContainer := container.NewHBox(layout.NewSpacer(), stats, layout.NewSpacer())
    
    // Panel central para el estacionamiento
    var fixedParkingSpaces [20]*canvas.Rectangle
    spacesContainer := container.NewGridWithColumns(10)
    
    for i := 0; i < 20; i++ {
        space := canvas.NewRectangle(color.RGBA{100, 100, 100, 255})
        space.Resize(fyne.NewSize(50, 50))
        fixedParkingSpaces[i] = space
        spacesContainer.Add(container.NewPadded(space))
    }
    
    // Entrada/Salida
    entrance := canvas.NewRectangle(color.RGBA{0, 255, 0, 255})
    entrance.Resize(fyne.NewSize(100, 20))
    entranceContainer := container.NewHBox(layout.NewSpacer(), entrance, layout.NewSpacer())
    
    content.Add(statsContainer)
    content.Add(spacesContainer)
    content.Add(entranceContainer)
    
    window.SetContent(content)
    window.Resize(fyne.NewSize(800, 400))
    
    gui := &ParkingGUI{
        spaces:   fixedParkingSpaces,
        entrance: entrance,
        stats:    stats,
        window:   window,
    }
    
    go func() {
        for text := range parking.updateChan {
            textCopy := text
            gui.window.Canvas().Refresh(gui.stats)
            gui.stats.SetText(textCopy)
        }
    }()
    
    parking.gui = gui
    return gui
}

func (p *ParkingLot) updateGUI() {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    // Actualizar espacios
    for i, occupied := range p.spaces {
        if occupied {
            p.gui.spaces[i].FillColor = color.RGBA{255, 0, 0, 255}
        } else {
            p.gui.spaces[i].FillColor = color.RGBA{100, 100, 100, 255}
        }
        p.gui.spaces[i].Refresh()
    }
    
    // Actualizar entrada
    switch p.direction {
    case 1: // Entrando
        p.gui.entrance.FillColor = color.RGBA{0, 0, 255, 255}
    case -1: // Saliendo
        p.gui.entrance.FillColor = color.RGBA{255, 165, 0, 255}
    default: // Libre
        p.gui.entrance.FillColor = color.RGBA{0, 255, 0, 255}
    }
    p.gui.entrance.Refresh()
}

func (p *ParkingLot) findAvailableSpace() int {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    for i := range p.spaces {
        if !p.spaces[i] {
            return i
        }
    }
    return -1
}

func (p *ParkingLot) occupySpace(space int) {
    p.mu.Lock()
    p.spaces[space] = true
    p.mu.Unlock()
    p.updateGUI()
}

func (p *ParkingLot) releaseSpace(space int) {
    p.mu.Lock()
    p.spaces[space] = false
    p.mu.Unlock()
    p.updateGUI()
}

func (p *ParkingLot) hasAvailableSpaces() bool {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    for _, occupied := range p.spaces {
        if !occupied {
            return true
        }
    }
    return false
}

func (p *ParkingLot) updateStats(text string) {
    p.updateChan <- text
}

func (v *Vehicle) Enter(p *ParkingLot) (int, bool) {
    if !p.hasAvailableSpaces() {
        p.updateStats(fmt.Sprintf("Vehículo %d esperando espacio", v.id))
        return -1, false
    }

    select {
    case p.entrance <- true:
        if p.direction == 0 || p.direction == 1 {
            p.direction = 1
            p.updateGUI()
            space := p.findAvailableSpace()
            if space != -1 {
                p.occupySpace(space)
                p.updateStats(fmt.Sprintf("Vehículo %d entrando al espacio %d", v.id, space))
                time.Sleep(time.Second)
                <-p.entrance
                p.direction = 0
                p.updateGUI()
                return space, true
            }
        }
        <-p.entrance
    default:
        if p.direction == -1 {
            p.updateStats(fmt.Sprintf("Vehículo %d esperando entrada", v.id))
            return -1, false
        }
    }
    return -1, false
}

func (v *Vehicle) Exit(p *ParkingLot, space int) bool {
    select {
    case p.entrance <- true:
        if p.direction == 0 || p.direction == -1 {
            p.direction = -1
            p.updateGUI()
            p.releaseSpace(space)
            
            // Incrementar contador y mostrar mensaje
            p.mu.Lock()
            p.vehiclesExited++
            exitCount := p.vehiclesExited
            p.mu.Unlock()
            
            mensaje := fmt.Sprintf("Vehículo %d salió del espacio %d (Total de salidas: %d)", v.id, space, exitCount)
            fmt.Println(mensaje)
            p.updateStats(mensaje)
            
            time.Sleep(time.Second)
            <-p.entrance
            p.direction = 0
            p.updateGUI()
            return true
        }
        <-p.entrance
    default:
        if p.direction == 1 {
            p.updateStats(fmt.Sprintf("Vehículo %d esperando para salir", v.id))
            return false
        }
    }
    return false
}

func simulateVehicle(id int, p *ParkingLot, wg *sync.WaitGroup) {
    defer wg.Done()
    vehicle := &Vehicle{id: id}
    
    var space int
    var success bool
    for {
        space, success = vehicle.Enter(p)
        if success {
            break
        }
        time.Sleep(time.Millisecond * 100)
    }
    
    parkingTime := 3 + rand.Intn(3)
    time.Sleep(time.Duration(parkingTime) * time.Second)
    
    for {
        if vehicle.Exit(p, space) {
            break
        }
        time.Sleep(time.Millisecond * 100)
    }
}

func main() {
    rand.Seed(time.Now().UnixNano())
    parking := NewParkingLot()
    gui := createGUI(parking)
    
    // Botón para iniciar simulación
    startBtn := widget.NewButton("Iniciar Simulación", func() {
        var wg sync.WaitGroup
        
        for i := 0; i < 100; i++ {
            wg.Add(1)
            go simulateVehicle(i+1, parking, &wg)
            time.Sleep(time.Duration(rand.ExpFloat64()) * time.Second)
        }
        
        go func() {
            wg.Wait()
            parking.updateStats("Simulación completada")
        }()
    })
    
    // Agregar botón al contenedor principal
    content := gui.window.Content().(*fyne.Container)
    content.Add(container.NewHBox(layout.NewSpacer(), startBtn, layout.NewSpacer()))
    
    gui.window.ShowAndRun()
}