package single_elevator

import "project/single-elevator/elevio"


func requests_above(e Elevator_t) bool {
	for f := e.floor+1; f < N_FLOORS; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_below(e Elevator_t) bool {
	for f := 0; f < e.floor; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_here(e Elevator_t) bool {
	for btn := 0; btn < N_BUTTONS; btn++ {
		if e.requests[e.floor][btn] {
			return true
		}
	}
	return false
}

func Requests_chooseDirection(e Elevator_t) DirectionBehaviourPair {
	switch(e.direction){
	case DIR_UP:
		if requests_above(e) {
			return DirectionBehaviourPair{DIR_UP, MOVING}
		} else if requests_here(e) {
			return DirectionBehaviourPair{DIR_DOWN, DOOR_OPEN}
		} else if requests_below(e) {
			return DirectionBehaviourPair{DIR_DOWN, MOVING}
		} else {
			return DirectionBehaviourPair{DIR_STOP, IDLE}
		}
	case DIR_DOWN:
		if requests_below(e) {
			return DirectionBehaviourPair{DIR_DOWN, MOVING}
		} else if requests_here(e) {
			return DirectionBehaviourPair{DIR_UP, DOOR_OPEN}
		} else if requests_above(e) {
			return DirectionBehaviourPair{DIR_UP, MOVING}
		} else {
			return DirectionBehaviourPair{DIR_STOP, IDLE}
		}
	case DIR_STOP:
		if requests_here(e) {
			return DirectionBehaviourPair{DIR_STOP, DOOR_OPEN}
		} else if requests_above(e) {
			return DirectionBehaviourPair{DIR_UP, MOVING}
		} else if requests_below(e) {
			return DirectionBehaviourPair{DIR_DOWN, MOVING}
		} else {
			return DirectionBehaviourPair{DIR_STOP, IDLE}
		}
	}
	return DirectionBehaviourPair{DIR_STOP, IDLE}
}

func Requests_shouldStop(e Elevator_t) bool {
	switch(e.direction){
	case DIR_DOWN:
		return e.requests[e.floor][elevio.BT_HallDown] || e.requests[e.floor][elevio.BT_Cab] || !requests_below(e)
	case DIR_UP:
		return e.requests[e.floor][elevio.BT_HallUp] || e.requests[e.floor][elevio.BT_Cab] || !requests_above(e)
	}

	return true
}

func Request_shouldClearCab(e Elevator_t) bool {
	return e.requests[e.floor][elevio.BT_Cab]
}

func Request_shouldClearUp(e Elevator_t) bool {
	if !e.requests[e.floor][elevio.BT_HallUp] {
		return false 
	}

	switch(e.direction){
	case DIR_UP, DIR_STOP:
		return true
	case DIR_DOWN:
		return !requests_below(e) && !e.requests[e.floor][elevio.BT_HallDown]
	}

	return false
}


func Request_shouldClearDown(e Elevator_t) bool {
	if !e.requests[e.floor][elevio.BT_HallDown] {
		return false 
	}

	switch(e.direction){
	case DIR_DOWN, DIR_STOP:
		return true
	case DIR_UP:
		return !requests_above(e) && !e.requests[e.floor][elevio.BT_HallUp]
	}

	return false
}


