package single_elevator_main

import (
	"project/elevio"
	"project/single-elevator/elevio"
)

type behaviour_t int

const (
	idle behaviour_t = iota
	door_open
	moving
)

type direction_t int

const (
	up direction_t = iota
	stop
	down
)

type elevator_t struct {
	behaviour behaviour_t
	direction direction_t
	floor     int
}

func run_elevator(orders_chan chan int) {

	numFloors := 4
	numButtons := 3

	elevio.Init("localhost:15657", numFloors)

	elevator := elevator_init()

	drv_floors := make(chan int)
	drv_obstr := make(chan bool)

	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)

	orders = [numFloors][numButtons]bool{
		{true, false, true},
		{false, false, false},
		{false, false, true},
		{true, false, false},
	}

	for {
		switch elevator.behaviour {
		case idle:
			if true in orders {
				FSM_start_moving()
			}
		case door_open:
			//Sjekk timer
			FSM_close_door()
		case moving:
			//Skal vi stoppe?
		}
	}
}

func elevator_init() elevator_t {
	elevio.SetMotorDirection(elevio.MD_Down)
	current_floor := <-drv_floors
	elevio.SetMotorDirection(elevio.MD_Stop)

	return elevator_t{behaviour: idle, direction: stop, floor: current_floor}
}

func FSM_arrive_floor(elevator elevator_t) {
	// check orders

	// if order on floor
	// set state to door open
	// stop motor
	// open the door
	// inform order is taken

	// if no order in direction (redundant)
	// print OMG!!!
	// set state to idle
	// stop motor

	// else
	// continue moving
	return
}

func FSM_close_door(elevator elevator_t) {
	// set state to idle

	// close door
}

func FSM_start_moving(elevator elevator_t) {
	// set state to moving

	// run algo for choose dir
	// run motor
}


func anyOrderExist()