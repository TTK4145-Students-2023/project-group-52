package main

import (
	//"project/single-elevator/elevio"
	nc "project/network-control"
	elevator "project/single-elevator"
	"project/single-elevator/elevio"
)

func main() {
	requests_chan := make(chan [elevator.N_FLOORS][elevator.N_BUTTONS]bool)
	completed_request_chan := make(chan elevio.ButtonEvent)

	go elevator.Run_elevator(requests_chan, completed_request_chan)
	go nc.RunNetworkControl(requests_chan, completed_request_chan)
	


	for {

	}
}
