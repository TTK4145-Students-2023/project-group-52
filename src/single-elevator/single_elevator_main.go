package single_elevator

import (
	//"fmt"
	"project/single-elevator/elevio"
	"time"
)

func Run_elevator(
	requests_chan <-chan [N_FLOORS][N_BUTTONS]bool,
	completed_request_chan chan<- elevio.ButtonEvent,
) {
	elevio.Init("localhost:15657", N_FLOORS)

	drv_floors := make(chan int)
	drv_obstr := make(chan bool)

	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)

	door_timeout := time.NewTimer(0)
	// remove initial trigger, as we don't want trigger to be possible before timer_start is called
	<-door_timeout.C

	elevator := elevator_init(drv_floors)

	for {
		select {
		case requests := <-requests_chan:
			ElevatorPrint(elevator)

			elevator.requests = requests

			switch elevator.behaviour {
			case DOOR_OPEN:
				if !elevio.IsObstruction() {
					timer_start(door_timeout)
				}
			case IDLE:
				elevator.direction, elevator.behaviour = Requests_chooseNewState(elevator)

				ElevatorPrint(elevator)
				switch elevator.behaviour {
				case DOOR_OPEN:
					elevio.SetDoorOpenLamp(true)
					if !elevio.IsObstruction() {
						timer_start(door_timeout)
					}

				case MOVING:
					elevio.SetMotorDirection(direction_converter(elevator.direction))
				}
			}
		case newFloor := <-drv_floors:
			ElevatorPrint(elevator)

			elevator.floor = newFloor
			elevio.SetFloorIndicator(elevator.floor)

			if elevator.behaviour == MOVING && Requests_shouldStop(elevator) {

				elevio.SetMotorDirection(elevio.MD_Stop)
				elevio.SetDoorOpenLamp(true)

				elevator.behaviour = DOOR_OPEN

				if !elevio.IsObstruction() {
					timer_start(door_timeout)
				}

			}
		case <-door_timeout.C:
			ElevatorPrint(elevator)

			if elevator.behaviour == DOOR_OPEN {
				if Request_shouldClearCab(elevator) {
					elevator.requests[elevator.floor][elevio.BT_Cab] = false
					completed_request_chan <- elevio.ButtonEvent{Floor: elevator.floor, Button: elevio.BT_Cab}
				}

				if Request_shouldClearUp(elevator) {
					elevator.requests[elevator.floor][elevio.BT_HallUp] = false
					completed_request_chan <- elevio.ButtonEvent{Floor: elevator.floor, Button: elevio.BT_HallUp}
				} else if Request_shouldClearDown(elevator) {
					elevator.requests[elevator.floor][elevio.BT_HallDown] = false
					completed_request_chan <- elevio.ButtonEvent{Floor: elevator.floor, Button: elevio.BT_HallDown}
				}

				elevator.direction, elevator.behaviour = Requests_chooseNewState(elevator)

				switch elevator.behaviour {
				case DOOR_OPEN:
					if !elevio.IsObstruction() {
						timer_start(door_timeout)
					}
				case MOVING, IDLE:
					elevio.SetDoorOpenLamp(false)
					elevio.SetMotorDirection(direction_converter(elevator.direction))
				}
			}
		case isObstructed := <-drv_obstr:
			if elevator.behaviour == DOOR_OPEN {
				if isObstructed {
					timer_kill(door_timeout)
				} else {
					timer_start(door_timeout)
				}
			}
		}
	}
}

func elevator_init(drv_floors <-chan int) Elevator_t {
	elevio.SetDoorOpenLamp(false)

	for f := 0; f < N_FLOORS; f++ {
		for b := elevio.ButtonType(0); b < N_BUTTONS; b++ {
			elevio.SetButtonLamp(b, f, false)
		}
	}

	elevio.SetMotorDirection(elevio.MD_Down)
	current_floor := <-drv_floors
	elevio.SetMotorDirection(elevio.MD_Stop)

	elevio.SetFloorIndicator(current_floor)

	return Elevator_t{floor: current_floor, direction: DIR_STOP, requests: [N_FLOORS][N_BUTTONS]bool{}, behaviour: IDLE}
}

func timer_start(t *time.Timer) {
	t.Reset(TIMEOUT_SEC * time.Second)
}

func timer_kill(t *time.Timer) {
	if !t.Stop() {
		<-t.C
	}
}

func direction_converter(dir Direction_t) elevio.MotorDirection {
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
