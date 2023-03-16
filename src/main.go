/*
TTK4145 elevator project
spring 2023
Anders Mørk, Erlend Dahle, Filip Strømstad
*/


package main

import (
	//"project/single-elevator/elevio"
	"flag"
	"project/elevator_control"
	"project/hardware/elevio"
	"project/request_control"
	. "project/types"
)

func main() {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	if id == "" {
		println("-id not provided")
		return
	}

	requests_chan := make(chan [N_FLOORS][N_BUTTONS]bool)
	completed_request_chan := make(chan elevio.ButtonEvent)

	go elevator_control.RunElevatorControl(requests_chan, completed_request_chan)
	go request_control.RunRequestControl(id, requests_chan, completed_request_chan)

	select {} //keeps main from exiting without using CPU power
}
