// presentation/gui.go
package presentation

import (
	"fmt"
	"os"
	"path/filepath"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// ParkingGUI representa la interfaz gráfica del estacionamiento
type ParkingGUI struct {
	Spaces    [20]*fyne.Container
	Entrance  *canvas.Rectangle
	Stats     *widget.Label
	Legend    *widget.Label
	Window    fyne.Window
	CarImages []string
}

// GUIService maneja las operaciones de la interfaz gráfica
type GUIService struct {
	gui *ParkingGUI
}

// NewGUIService crea una nueva instancia del servicio GUI
func NewGUIService(gui *ParkingGUI) *GUIService {
	return &GUIService{gui: gui}
}

// UpdateParkingSpace actualiza el estado visual de un espacio de estacionamiento
func (s *GUIService) UpdateParkingSpace(index int, occupied bool, carImage string) {
	container := s.gui.Spaces[index]
	if occupied {
		if len(container.Objects) <= 2 {
			carImg := canvas.NewImageFromFile(carImage)
			carImg.Resize(fyne.NewSize(50, 50))
			carImg.FillMode = canvas.ImageFillContain
			container.Add(carImg)
		}
	} else {
		if len(container.Objects) > 2 {
			container.Objects = container.Objects[:2]
		}
	}
	container.Refresh()
}

// UpdateEntranceColor actualiza el color de la entrada según el estado
func (s *GUIService) UpdateEntranceColor(direction int) {
	switch direction {
	case 1:
		s.gui.Entrance.FillColor = color.RGBA{0, 123, 255, 255}
	case -1:
		s.gui.Entrance.FillColor = color.RGBA{255, 123, 0, 255}
	default:
		s.gui.Entrance.FillColor = color.RGBA{40, 167, 69, 255}
	}
	s.gui.Entrance.Refresh()
}

// CreateGUI crea y configura la interfaz gráfica del estacionamiento
func CreateGUI() *ParkingGUI {
	myApp := app.New()
	window := myApp.NewWindow("Simulador de Estacionamiento")

	// Crear contenedor principal con padding
	mainContainer := container.NewPadded()
	content := container.NewVBox()

	// Configurar estadísticas
	stats := widget.NewLabelWithStyle("Vehículos en espera: 0", 
		fyne.TextAlignCenter, 
		fyne.TextStyle{Bold: true})
	statsContainer := container.NewHBox(layout.NewSpacer(), stats, layout.NewSpacer())

	// Configurar leyenda
	legendCard := widget.NewCard("Leyenda", "", nil)
	legendText := "🟢 Verde: Entrada libre\n" +
		"🔵 Azul: Vehículo entrando\n" +
		"🟠 Naranja: Vehículo saliendo\n" +
		"⚪ Gris: Espacio libre\n"
	legend := widget.NewLabelWithStyle(legendText, 
		fyne.TextAlignLeading, 
		fyne.TextStyle{Monospace: true})
	legendCard.SetContent(legend)

	// Configurar espacios de estacionamiento
	var fixedParkingSpaces [20]*fyne.Container
	spacesContainer := container.NewGridWithColumns(10)

	// Cargar imágenes de vehículos
	var carImages []string
	files, _ := os.ReadDir("assets")
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".png" {
			carImages = append(carImages, filepath.Join("assets", file.Name()))
		}
	}

	// Crear espacios de estacionamiento
	for i := 0; i < 20; i++ {
		space := canvas.NewRectangle(color.RGBA{200, 200, 200, 255})
		space.SetMinSize(fyne.NewSize(60, 60))
		space.Resize(fyne.NewSize(60, 60))
		
		spaceNumber := canvas.NewText(fmt.Sprintf("%d", i+1), color.Black)
		spaceNumber.TextSize = 12
		spaceNumber.Alignment = fyne.TextAlignCenter
		
		spaceContainer := container.NewStack(space, spaceNumber)
		fixedParkingSpaces[i] = spaceContainer
		spacesContainer.Add(container.NewPadded(spaceContainer))
	}

	// Configurar entrada/salida
	entranceCard := widget.NewCard("Entrada/Salida", "", nil)
	entrance := canvas.NewRectangle(color.RGBA{0, 255, 0, 255})
	entrance.SetMinSize(fyne.NewSize(300, 40))
	entranceContainer := container.NewHBox(layout.NewSpacer(), entrance, layout.NewSpacer())
	entranceCard.SetContent(entranceContainer)

	// Configurar botón de inicio
	startBtn := widget.NewButton("Iniciar Simulación", nil)
	startBtn.Importance = widget.HighImportance
	startBtn.Resize(fyne.NewSize(200, 40))
	buttonContainer := container.NewHBox(layout.NewSpacer(), startBtn, layout.NewSpacer())

	// Organizar elementos en la interfaz
	content.Add(statsContainer)
	content.Add(widget.NewSeparator())
	content.Add(legendCard)
	content.Add(widget.NewSeparator())
	content.Add(spacesContainer)
	content.Add(widget.NewSeparator())
	content.Add(entranceCard)
	content.Add(buttonContainer)

	mainContainer.Add(content)
	window.SetContent(mainContainer)
	window.Resize(fyne.NewSize(900, 600))

	return &ParkingGUI{
		Spaces:    fixedParkingSpaces,
		Entrance:  entrance,
		Stats:     stats,
		Legend:    legend,
		Window:    window,
		CarImages: carImages,
	}
}