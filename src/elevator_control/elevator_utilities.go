package elevator_control

import (
	elevio "project/hardware"
	. "project/types"
	"time"
)

var sharedInfo ElevatorSharedInfo_t

func GetElevatorInfo() ElevatorInfo_t {
	sharedInfo.Mutex.RLock()
	defer sharedInfo.Mutex.RUnlock()

	return ElevatorInfo_t{
		Available: sharedInfo.Available,
		Behaviour: sharedInfo.Behaviour,
		Direction: sharedInfo.Direction,
		Floor:     sharedInfo.Floor,
	}
}

func updateElevatorInfo(e Elevator_t) {
	sharedInfo.Mutex.Lock()
	defer sharedInfo.Mutex.Unlock()

	sharedInfo.Behaviour = e.Behaviour
	sharedInfo.Direction = e.Direction
	sharedInfo.Floor = e.Floor
}

func setElevatorAvailability(value bool) {
	sharedInfo.Mutex.Lock()
	defer sharedInfo.Mutex.Unlock()

	sharedInfo.Available = value
}

func elevatorInit(floorSensorCh <-chan int) Elevator_t {
	elevio.SetDoorOpenLamp(false)

	for f := 0; f < N_FLOORS; f++ {
		for b := ButtonType_t(0); b < N_BUTTONS; b++ {
			elevio.SetButtonLamp(b, f, false)
		}
	}

	elevio.SetMotorDirection(MD_Down)
	currentFloor := <-floorSensorCh
	elevio.SetMotorDirection(MD_Stop)

	elevio.SetFloorIndicator(currentFloor)

	return Elevator_t{Floor: currentFloor, Direction: DIR_STOP, Requests: [N_FLOORS][N_BUTTONS]bool{}, Behaviour: IDLE}
}

func timerRestart(timer *time.Timer, sec int) {
	timer.Reset(time.Duration(sec) * time.Second)

}

func timerKill(timer *time.Timer) {
	if !timer.Stop() {
		<-timer.C
	}
}

func directionConverter(dir Direction_t) MotorDirection_t {
	switch dir {
	case DIR_UP:
		return MD_Up
	case DIR_DOWN:
		return MD_Down
	case DIR_STOP:
		return MD_Stop
	}
	return MD_Stop
}
