package single_elevator

import (
	"fmt"
	"project/single-elevator/elevio"
)

func Run_elevator(requests_chan <-chan [N_FLOORS][N_BUTTONS]bool) {
	elevio.Init("localhost:15657", N_FLOORS)

	drv_floors := make(chan int)
	drv_obstr := make(chan bool)

	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)

	timer_timeout := make(chan bool)
	go Timer_main(timer_timeout)

	elevator := elevator_init(drv_floors)

	for {
		select {
		case requests := <-requests_chan:
			elevator.requests = requests
			FSM_NewOrdersAssigned(&elevator)
		case newFloor := <-drv_floors:
			ElevatorPrint(elevator)
			fmt.Println("New floor: ", newFloor)
			FSM_onFloorArrival(&elevator, newFloor)
		case <-timer_timeout:
			FSM_onDoorTimeout(&elevator)
			ElevatorPrint(elevator)
		case isObstructed := <- drv_obstr:
			FSM_obstructionTrigger(&elevator, isObstructed)
		}
	}
}
