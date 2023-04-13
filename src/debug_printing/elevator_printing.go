package printing

import (
	"fmt"
	. "project/types"
)

func Eb_toString(behaviour Behaviour_t) string {
	switch(behaviour){
	case IDLE:
		return "IDLE"
	case DOOR_OPEN:
		return "DOOR_OPEN"
	case MOVING:
		return "MOVING"
	}
	return "UNKNOWN"
}

func Ed_toString(dir Direction_t) string {
	switch(dir){
	case DIR_UP:
		return "DIR_UP"
	case DIR_DOWN:
		return "DIR_DOWN"
	case DIR_STOP:
		return "DIR_STOP"
	}
	return "UNDEFINED"
}

func ElevatorPrint(elevator Elevator_t) {
	fmt.Printf(
		" |floor: %-2d|\n"+
			" |dirn: %-12.12s|\n"+
			" |behaviour: %-12.12s|\n",
		elevator.Floor,
		Ed_toString(elevator.Direction),
		Eb_toString(elevator.Behaviour),
	)
}
