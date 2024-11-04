package domain

import (
	"fmt"
	"time"
)

type ParkingService struct {
	lot *ParkingLot
}

func NewParkingService(lot *ParkingLot) *ParkingService {
	return &ParkingService{lot: lot}
}

func (s *ParkingService) FindAvailableSpace() int {
	s.lot.Mu.Lock()
	defer s.lot.Mu.Unlock()

	for i := range s.lot.Spaces {
		if !s.lot.Spaces[i] {
			return i
		}
	}
	return -1
}

func (s *ParkingService) OccupySpace(space int) {
	s.lot.Mu.Lock()
	s.lot.Spaces[space] = true
	s.lot.Mu.Unlock()
}

func (s *ParkingService) ReleaseSpace(space int) {
	s.lot.Mu.Lock()
	s.lot.Spaces[space] = false
	s.lot.Mu.Unlock()
}

func (s *ParkingService) UpdateStats(text string) {
	s.lot.UpdateChan <- text
}

func (s *ParkingService) VehicleEnter(v *Vehicle) (int, bool) {
	select {
	case <-s.lot.SpacesAvailable:
		select {
		case s.lot.Entrance <- true:
			if s.lot.Direction == 0 || s.lot.Direction == 1 {
				s.lot.Direction = 1
				space := s.FindAvailableSpace()
				if space != -1 {
					s.OccupySpace(space)
					s.UpdateStats(fmt.Sprintf("Vehículo %d entrando al espacio %d", v.ID, space))
					time.Sleep(time.Millisecond * 500)
					<-s.lot.Entrance
					s.lot.Direction = 0
					return space, true
				}
			}
			<-s.lot.Entrance
		default:
			s.lot.SpacesAvailable <- true
			if s.lot.Direction == -1 {
				s.UpdateStats(fmt.Sprintf("Vehículo %d esperando entrada", v.ID))
				return -1, false
			}
		}
	default:
		s.UpdateStats(fmt.Sprintf("Vehículo %d esperando espacio", v.ID))
		return -1, false
	}
	return -1, false
}

func (s *ParkingService) VehicleExit(v *Vehicle, space int) bool {
	select {
	case s.lot.Entrance <- true:
		if s.lot.Direction == 0 || s.lot.Direction == -1 {
			s.lot.Direction = -1
			s.ReleaseSpace(space)
			s.lot.SpacesAvailable <- true

			s.lot.Mu.Lock()
			s.lot.VehiclesExited++
			exitCount := s.lot.VehiclesExited
			s.lot.Mu.Unlock()

			mensaje := fmt.Sprintf("Vehículo %d salió del espacio %d (Total de salidas: %d)", v.ID, space, exitCount)
			s.UpdateStats(mensaje)

			time.Sleep(time.Millisecond * 500)
			<-s.lot.Entrance
			s.lot.Direction = 0
			return true
		}
		<-s.lot.Entrance
	default:
		if s.lot.Direction == 1 {
			s.UpdateStats(fmt.Sprintf("Vehículo %d esperando para salir", v.ID))
			return false
		}
	}
	return false
}