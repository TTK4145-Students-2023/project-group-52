package single_elevator

import(
	"project/single-elevator/elevio"
	"fmt"
)

func elevator_init(drv_floors <-chan int) Elevator_t {
	elevio.SetDoorOpenLamp(false)

	for f := 0; f < N_FLOORS; f++ {
		for b := elevio.ButtonType(0); b < N_BUTTONS; b++ {
			elevio.SetButtonLamp(b,f,false)
		}
	}

	elevio.SetMotorDirection(elevio.MD_Down)
	current_floor := <-drv_floors
	elevio.SetMotorDirection(elevio.MD_Stop)

	elevio.SetFloorIndicator(current_floor)

	return Elevator_t{floor: current_floor, direction: DIR_STOP, requests: [N_FLOORS][N_BUTTONS]bool{}, behaviour: IDLE}
}

func FSM_onFloorArrival(elevator *Elevator_t, newFloor int, completed_request_chan chan<- elevio.ButtonEvent){
	elevator.floor = newFloor

	elevio.SetFloorIndicator(elevator.floor)

	if elevator.behaviour == MOVING {
		if(Requests_shouldStop(*elevator)){
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			
			elevator.behaviour = DOOR_OPEN
			
			if !elevio.GetObstruction(){
				Timer_start()
			}

			if Request_shouldClearCab(*elevator) {
				elevator.requests[elevator.floor][elevio.BT_Cab] = false
				completed_request_chan <- elevio.ButtonEvent{Floor: elevator.floor, Button: elevio.BT_Cab}
			}

			if Request_shouldClearUp(*elevator) {
				elevator.requests[elevator.floor][elevio.BT_HallUp] = false
				completed_request_chan <- elevio.ButtonEvent{Floor: elevator.floor, Button: elevio.BT_HallUp}
			}

			if Request_shouldClearDown(*elevator) {
				elevator.requests[elevator.floor][elevio.BT_HallDown] = false
				completed_request_chan <- elevio.ButtonEvent{Floor: elevator.floor, Button: elevio.BT_HallDown}
			}			
		}
	}
}

func FSM_NewOrdersAssigned(elevator *Elevator_t, completed_request_chan chan<- elevio.ButtonEvent){
	switch(elevator.behaviour){
	case MOVING:
		// do nothing
	case DOOR_OPEN:
		if Request_shouldClearCab(*elevator) {
			Timer_start()
			elevator.requests[elevator.floor][elevio.BT_Cab] = false
			completed_request_chan <- elevio.ButtonEvent{Floor: elevator.floor, Button: elevio.BT_Cab}
		}

		if Request_shouldClearUp(*elevator) {
			Timer_start()
			elevator.requests[elevator.floor][elevio.BT_HallUp] = false
			completed_request_chan <- elevio.ButtonEvent{Floor: elevator.floor, Button: elevio.BT_HallUp}
		}

		if Request_shouldClearDown(*elevator) {
			Timer_start()
			elevator.requests[elevator.floor][elevio.BT_HallDown] = false
			completed_request_chan <- elevio.ButtonEvent{Floor: elevator.floor, Button: elevio.BT_HallDown}
		}

	case IDLE:
		pair := Requests_chooseDirection(*elevator)
		elevator.direction = pair.direction
		elevator.behaviour = pair.behaviour

		switch(elevator.behaviour){
		case DOOR_OPEN:
			elevio.SetDoorOpenLamp(true)
			Timer_start()

			if Request_shouldClearCab(*elevator) {
				elevator.requests[elevator.floor][elevio.BT_Cab] = false
				completed_request_chan <- elevio.ButtonEvent{Floor: elevator.floor, Button: elevio.BT_Cab}
			}

			if Request_shouldClearUp(*elevator) {
				elevator.requests[elevator.floor][elevio.BT_HallUp] = false
				completed_request_chan <- elevio.ButtonEvent{Floor: elevator.floor, Button: elevio.BT_HallUp}
			}

			if Request_shouldClearDown(*elevator) {
				elevator.requests[elevator.floor][elevio.BT_HallDown] = false
				completed_request_chan <- elevio.ButtonEvent{Floor: elevator.floor, Button: elevio.BT_HallDown}
			}				
			
		case MOVING:
			elevio.SetMotorDirection(direction_converter(elevator.direction))

		case IDLE:
			//do nothing
		}
	}
}

func FSM_onDoorTimeout(elevator *Elevator_t){
	if elevator.behaviour == DOOR_OPEN {
		pair := Requests_chooseDirection(*elevator)
		elevator.direction = pair.direction
		elevator.behaviour = pair.behaviour

		switch(elevator.behaviour){
		case DOOR_OPEN:
			Timer_start()
			fmt.Println("State door open after door open")
		case MOVING, IDLE:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(direction_converter(elevator.direction))
		}
	}
}

func FSM_obstructionTrigger(elevator *Elevator_t, isObstructed bool){
	if elevator.behaviour == DOOR_OPEN {
		if isObstructed {
			Timer_kill()
		} else {
			Timer_start()
		}
	}
}


func direction_converter(dir Direction_t) elevio.MotorDirection {
	switch(dir){
	case DIR_UP:
		return elevio.MD_Up
	case DIR_DOWN:
		return elevio.MD_Down
	case DIR_STOP:
		return elevio.MD_Stop
	}
	return elevio.MD_Stop
}