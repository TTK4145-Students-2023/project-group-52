package elevator_control

import (
	"project/hardware"
	. "project/types"
	"time"
)

func elevatorInit(drv_Floors <-chan int) Elevator_t {
	elevio.SetDoorOpenLamp(false)

	for f := 0; f < N_FLOORS; f++ {
		for b := ButtonType_t(0); b < N_BUTTONS; b++ {
			elevio.SetButtonLamp(b, f, false)
		}
	}

	elevio.SetMotorDirection(MD_Down)
	current_Floor := <-drv_Floors
	elevio.SetMotorDirection(MD_Stop)

	elevio.SetFloorIndicator(current_Floor)

	return Elevator_t{Floor: current_Floor, Direction: DIR_STOP, Requests: [N_FLOORS][N_BUTTONS]bool{}, Behaviour: IDLE}
}

func GetElevatorState() (bool, Behaviour_t, Direction_t, int) {
	shared_state.Mutex.RLock()
	defer shared_state.Mutex.RUnlock()

	return shared_state.Available, shared_state.Behaviour, shared_state.Direction, shared_state.Floor
}

func updateElevatorState(e Elevator_t) {
	shared_state.Mutex.Lock()
	defer shared_state.Mutex.Unlock()

	shared_state.Behaviour = e.Behaviour
	shared_state.Direction = e.Direction
	shared_state.Floor = e.Floor
}

func setElevatorAvailability(value bool) {
	shared_state.Mutex.Lock()
	defer shared_state.Mutex.Unlock()

	shared_state.Available = value
}

func handleObstruction(door_timeout *time.Timer) {
	if elevio.IsObstruction() {
		setElevatorAvailability(false)
	} else {
		timerRestart(door_timeout, DOOR_TIMEOUT_SEC)
	}
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
