package single_elevator

import (
	"fmt"
	"time"
	"project/single-elevator/elevio"
)

func Run_elevator(
	requests_chan <-chan [N_FLOORS][N_BUTTONS]bool,
	completed_request_chan chan<- elevio.ButtonEvent,
) {
	elevio.Init("localhost:15657", N_FLOORS)

	drv_floors := make(chan int)
	drv_obstr := make(chan bool)

	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)

	timer_timeout := time.NewTimer(0)
	// remove initial trigger, as we don't want trigger to be possible before timer_start is called
	<-timer_timeout.C 

	elevator := elevator_init(drv_floors)

	for {
		select {
		case requests := <-requests_chan:
			ElevatorPrint(elevator)
			elevator.requests = requests
			FSM_NewOrdersAssigned(&elevator, timer_timeout, completed_request_chan)
		case newFloor := <-drv_floors:
			ElevatorPrint(elevator)
			fmt.Println("New floor: ", newFloor)
			FSM_onFloorArrival(&elevator,timer_timeout, newFloor, completed_request_chan)
		case <-timer_timeout.C:
			FSM_onDoorTimeout(&elevator, timer_timeout, completed_request_chan)
			ElevatorPrint(elevator)
		case isObstructed := <-drv_obstr:
			FSM_obstructionTrigger(&elevator, timer_timeout, isObstructed)
		}
	}
}
