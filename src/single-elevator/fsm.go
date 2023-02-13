package single_elevator

import(
	"project/single-elevator/elevio"
	"fmt"
)

func FSM_NewOrdersAssigned(elevator *Elevator_t){
	switch(elevator.behaviour){
	case DOOR_OPEN,MOVING:
		//Do nothing
	case IDLE:
		pair := Requests_chooseDirection(*elevator)
		elevator.direction = pair.direction
		elevator.behaviour = pair.behaviour

		switch(elevator.behaviour){
		case DOOR_OPEN:
			elevio.SetDoorOpenLamp(true)
			Timer_start()
			fmt.Println("Order taken")
			//Clear orders
			
			
		case MOVING:
			elevio.SetMotorDirection(direction_converter(elevator.direction))

		case IDLE:
			//do nothing
		}
	}
	


}

func elevator_init(drv_floors <-chan int) Elevator_t {
	elevio.SetMotorDirection(elevio.MD_Down)
	current_floor := <-drv_floors
	elevio.SetMotorDirection(elevio.MD_Stop)

	return Elevator_t{floor: current_floor, direction: DIR_STOP, requests: [N_FLOORS][N_BUTTONS]bool{}, behaviour: IDLE}
}

func FSM_onFloorArrival(elevator *Elevator_t, newFloor int){
	elevator.floor = newFloor

	elevio.SetFloorIndicator(elevator.floor)

	if elevator.behaviour == MOVING {
		if(Requests_shouldStop(*elevator)){
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			fmt.Println("order taken")

			Timer_start()

			elevator.behaviour = DOOR_OPEN
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