package single_elevator

import (
	"project/single-elevator/elevio"
	"time"
	//"fmt"
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

func FSM_onFloorArrival(elevator *Elevator_t, timer *time.Timer, newFloor int, completed_request_chan chan<- elevio.ButtonEvent){
	elevator.floor = newFloor

	elevio.SetFloorIndicator(elevator.floor)

	if elevator.behaviour == MOVING {
		if(Requests_shouldStop(*elevator)){
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			
			elevator.behaviour = DOOR_OPEN
			
			if !elevio.GetObstruction(){
				timer_start(timer)
			}

			if clearOrdersIfNeeded(elevator, completed_request_chan) {
				println(1)
			}		
		}
	}
}

func FSM_NewOrdersAssigned(elevator *Elevator_t, timer *time.Timer, completed_request_chan chan<- elevio.ButtonEvent){
	switch(elevator.behaviour){
	case MOVING:
		// do nothing
	case DOOR_OPEN:
		if clearOrdersIfNeeded(elevator, completed_request_chan) {
			println(2)
			if !elevio.GetObstruction(){
				timer_start(timer)
			}
		}

	case IDLE:
		pair := Requests_chooseDirection(*elevator)
		elevator.direction = pair.direction
		elevator.behaviour = pair.behaviour
		ElevatorPrint(*elevator)


		switch(elevator.behaviour){
		case DOOR_OPEN:
			elevio.SetDoorOpenLamp(true)
			if !elevio.GetObstruction(){
				timer_start(timer)
			}

			if clearOrdersIfNeeded(elevator, completed_request_chan) {
				println(3)
			}				
			
		case MOVING:
			elevio.SetMotorDirection(direction_converter(elevator.direction))

		case IDLE:
			//do nothing
		}
	}
}

func FSM_onDoorTimeout(elevator *Elevator_t, timer *time.Timer, completed_request_chan chan<- elevio.ButtonEvent){
	if elevator.behaviour == DOOR_OPEN {
		pair := Requests_chooseDirection(*elevator)
		elevator.direction = pair.direction
		elevator.behaviour = pair.behaviour

		switch(elevator.behaviour){
		case DOOR_OPEN:
			if !elevio.GetObstruction(){
				timer_start(timer)
			}
			if clearOrdersIfNeeded(elevator, completed_request_chan) {
				println(4)
			}	
		case MOVING, IDLE:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(direction_converter(elevator.direction))
		}
	}
}

func FSM_obstructionTrigger(elevator *Elevator_t, timer *time.Timer, isObstructed bool){
	if elevator.behaviour == DOOR_OPEN {
		if isObstructed {
			timer_kill(timer)
		} else {
			timer_start(timer)
		}
	}
}

func clearOrdersIfNeeded(elevator * Elevator_t, completed_request_chan chan<- elevio.ButtonEvent) bool {
	ordersCleared := false

	if Request_shouldClearCab(*elevator) {
		ordersCleared = true
		elevator.requests[elevator.floor][elevio.BT_Cab] = false
		completed_request_chan <- elevio.ButtonEvent{Floor: elevator.floor, Button: elevio.BT_Cab}
	}

	if Request_shouldClearUp(*elevator) {
		ordersCleared = true
		elevator.requests[elevator.floor][elevio.BT_HallUp] = false
		// BUG: PROGRAM HALTS HERE WHEN ON A FLOOR AND UP AND DOWN BUTTON PRESSED AT THE SAME TIME ish
		completed_request_chan <- elevio.ButtonEvent{Floor: elevator.floor, Button: elevio.BT_HallUp}
	} else if Request_shouldClearDown(*elevator) {
		ordersCleared = true
		elevator.requests[elevator.floor][elevio.BT_HallDown] = false
		completed_request_chan <- elevio.ButtonEvent{Floor: elevator.floor, Button: elevio.BT_HallDown}
	}

	return ordersCleared
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

func timer_start(t *time.Timer){
	t.Reset(TIMEOUT_SEC * time.Second)
}

func timer_kill(t *time.Timer){
	if !t.Stop() {
		<-t.C
	}
}