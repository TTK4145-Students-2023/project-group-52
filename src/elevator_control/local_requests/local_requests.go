package local_requests

import (
	. "project/types"
)

func isRequestsAbove(e Elevator_t) bool {
	for f := e.Floor + 1; f < N_FLOORS; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func isRequestsBelow(e Elevator_t) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func isRequestsHere(e Elevator_t) bool {
	for btn := 0; btn < N_BUTTONS; btn++ {
		if e.Requests[e.Floor][btn] {
			return true
		}
	}
	return false
}

func getRequestTypeHere(e Elevator_t) ButtonType_t {
	for btn := 0; btn < N_BUTTONS; btn++ {
		if e.Requests[e.Floor][btn] {
			return ButtonType_t(btn)
		}
	}
	print("buttonType not found")
	return BT_Cab
}

func ChooseNewDirectionAndBehavior(e Elevator_t) (Direction_t, Behaviour_t) {
	switch e.Direction {
	case DIR_UP:
		if isRequestsAbove(e) {
			return DIR_UP, MOVING
		} else if isRequestsHere(e) {
			return DIR_DOWN, DOOR_OPEN
		} else if isRequestsBelow(e) {
			return DIR_DOWN, MOVING
		} else {
			return DIR_STOP, IDLE
		}
	case DIR_DOWN:
		if isRequestsBelow(e) {
			return DIR_DOWN, MOVING
		} else if isRequestsHere(e) {
			return DIR_UP, DOOR_OPEN
		} else if isRequestsAbove(e) {
			return DIR_UP, MOVING
		} else {
			return DIR_STOP, IDLE
		}
	case DIR_STOP:
		if isRequestsHere(e) {
			switch getRequestTypeHere(e) {
			case BT_HallUp:
				return DIR_UP, DOOR_OPEN
			case BT_HallDown:
				return DIR_DOWN, DOOR_OPEN
			case BT_Cab:
				return DIR_STOP, DOOR_OPEN
			}
		} else if isRequestsAbove(e) {
			return DIR_UP, MOVING
		} else if isRequestsBelow(e) {
			return DIR_DOWN, MOVING
		} else {
			return DIR_STOP, IDLE
		}
	}
	return DIR_STOP, IDLE
}

func ShouldStop(e Elevator_t) bool {
	switch e.Direction {
	case DIR_DOWN:
		return e.Requests[e.Floor][BT_HallDown] || e.Requests[e.Floor][BT_Cab] || !isRequestsBelow(e)
	case DIR_UP:
		return e.Requests[e.Floor][BT_HallUp] || e.Requests[e.Floor][BT_Cab] || !isRequestsAbove(e)
	}

	return true
}

func ShouldClearCab(e Elevator_t) bool {
	return e.Requests[e.Floor][BT_Cab]
}

func ShouldClearUp(e Elevator_t) bool {
	if !e.Requests[e.Floor][BT_HallUp] {
		return false
	}

	switch e.Direction {
	case DIR_UP, DIR_STOP:
		return true
	case DIR_DOWN:
		return !isRequestsBelow(e) && !e.Requests[e.Floor][BT_HallDown]
	}

	return false
}

func ShouldClearDown(e Elevator_t) bool {
	if !e.Requests[e.Floor][BT_HallDown] {
		return false
	}

	switch e.Direction {
	case DIR_DOWN, DIR_STOP:
		return true
	case DIR_UP:
		return !isRequestsAbove(e) && !e.Requests[e.Floor][BT_HallUp]
	}

	return false
}
