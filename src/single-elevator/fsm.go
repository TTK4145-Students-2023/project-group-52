package fsm

func FSM_init() {
	// set state to idle

	// go to defined floor
}

func FSM_arrive_floor() {
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
}

func FSM_close_door() {
	// set state to idle

	// close door
}

func FSM_start_moving() {
	// set state to moving

	// run algo for choose dir
	// run motor
}
