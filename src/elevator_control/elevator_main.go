package elevator_control

import (
	//"fmt"
	req "project/elevator_control/local_requests"
	"project/hardware/elevio"
	. "project/types"
	"time"
)

const DOOR_TIMEOUT_SEC = 3

var shared_state ElevatorSharedState_t

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

func RunElevatorControl(
	requests_chan <-chan [N_FLOORS][N_BUTTONS]bool,
	completed_request_chan chan<- elevio.ButtonEvent,
) {
	elevio.Init("localhost:15657", N_FLOORS)

	
	drv_Floors := make(chan int)
	drv_obstr := make(chan bool)
	
	go elevio.PollFloorSensor(drv_Floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	
	door_timeout := time.NewTimer(0)
	
	elevator := elevator_init(drv_Floors)
	updateElevatorState(elevator)
	setElevatorAvailability(true)

	for {
		select {
		case requests := <-requests_chan:
			elevator.Requests = requests

			switch elevator.Behaviour {
			case DOOR_OPEN:
				// we removed reset timer here
			case IDLE:
				elevator.Direction, elevator.Behaviour = req.ChooseNewDirectionAndBehavior(elevator)

				switch elevator.Behaviour {
				case DOOR_OPEN:
					elevio.SetDoorOpenLamp(true)
					if !elevio.IsObstruction() {
						timer_start(door_timeout)
					}

				case MOVING:
					elevio.SetMotorDirection(Direction_converter(elevator.Direction))
				}
			}
			updateElevatorState(elevator)
		case newFloor := <-drv_Floors:
			elevator.Floor = newFloor
			elevio.SetFloorIndicator(elevator.Floor)

			if elevator.Behaviour == MOVING && req.ShouldStop(elevator) {

				elevio.SetMotorDirection(elevio.MD_Stop)
				elevio.SetDoorOpenLamp(true)

				elevator.Behaviour = DOOR_OPEN

				if !elevio.IsObstruction() {
					timer_start(door_timeout)
				}

			}
			updateElevatorState(elevator)
		case <-door_timeout.C:
			if elevator.Behaviour == DOOR_OPEN {
				if req.ShouldClearCab(elevator) {
					elevator.Requests[elevator.Floor][elevio.BT_Cab] = false
					completed_request_chan <- elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_Cab}
				}

				if req.ShouldClearUp(elevator) {
					elevator.Requests[elevator.Floor][elevio.BT_HallUp] = false
					completed_request_chan <- elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_HallUp}
				} else if req.ShouldClearDown(elevator) {
					elevator.Requests[elevator.Floor][elevio.BT_HallDown] = false
					completed_request_chan <- elevio.ButtonEvent{Floor: elevator.Floor, Button: elevio.BT_HallDown}
				}

				elevator.Direction, elevator.Behaviour = req.ChooseNewDirectionAndBehavior(elevator)

				switch elevator.Behaviour {
				case DOOR_OPEN:
					if !elevio.IsObstruction() {
						timer_start(door_timeout)
					}
				case MOVING, IDLE:
					elevio.SetDoorOpenLamp(false)
					elevio.SetMotorDirection(Direction_converter(elevator.Direction))
				}
			}
			updateElevatorState(elevator)
		case isObstructed := <-drv_obstr:
			if elevator.Behaviour == DOOR_OPEN {
				if isObstructed {
					timer_kill(door_timeout)
				} else {
					timer_start(door_timeout)
				}
			}
		}
	}
}

func elevator_init(drv_Floors <-chan int) Elevator_t {
	elevio.SetDoorOpenLamp(false)

	for f := 0; f < N_FLOORS; f++ {
		for b := elevio.ButtonType(0); b < N_BUTTONS; b++ {
			elevio.SetButtonLamp(b, f, false)
		}
	}

	elevio.SetMotorDirection(elevio.MD_Down)
	current_Floor := <-drv_Floors
	elevio.SetMotorDirection(elevio.MD_Stop)

	elevio.SetFloorIndicator(current_Floor)

	return Elevator_t{Floor: current_Floor, Direction: DIR_STOP, Requests: [N_FLOORS][N_BUTTONS]bool{}, Behaviour: IDLE}
}

func timer_start(t *time.Timer) {
	t.Reset(DOOR_TIMEOUT_SEC * time.Second)
}

func timer_kill(t *time.Timer) {
	if !t.Stop() {
		<-t.C
	}
}

func Direction_converter(dir Direction_t) elevio.MotorDirection {
	switch dir {
	case DIR_UP:
		return elevio.MD_Up
	case DIR_DOWN:
		return elevio.MD_Down
	case DIR_STOP:
		return elevio.MD_Stop
	}
	return elevio.MD_Stop
}
