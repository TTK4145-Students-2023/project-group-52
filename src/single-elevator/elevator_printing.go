package single_elevator

import("fmt")

func eb_toString(behaviour Behaviour_t)string{
	if(behaviour==IDLE){
		return "IDLE"
	}
	if(behaviour==DOOR_OPEN){
		return "DOOR_OPEN"
	}
	if(behaviour==MOVING){
		return "MOVING"
	}
	return "UNKNown"
}
func ed_toString(dir Direction_t)string{
	if(dir==DIR_UP){
		return "DIR_UP"
	}
	if(dir==DIR_DOWN){
		return "DIR_DOWN"
	}
	if(dir==DIR_STOP){
		return "DIR_STOP"
	}
	return "UNDEFINED"
}


func ElevatorPrint(elevator Elevator_t){
	fmt.Printf(
		" |floor: %-2d|\n"+
		" |dirn: %-12.12s|\n"+
		" |behaviour: %-12.12s|\n",
		elevator.floor,
		ed_toString(elevator.direction),
		eb_toString(elevator.behaviour),
		
	)
}