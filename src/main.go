package main

import (
	//"project/single-elevator/elevio"
	"flag"
	nc "project/network-control"
	elevator "project/single-elevator"
	"project/single-elevator/elevio"
)

func main() {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	requests_chan := make(chan [elevator.N_FLOORS][elevator.N_BUTTONS]bool)
	completed_request_chan := make(chan elevio.ButtonEvent)

	go elevator.Run_elevator(requests_chan, completed_request_chan)
	go nc.RunNetworkControl(id, requests_chan, completed_request_chan)

	for {

	}
}
