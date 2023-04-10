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
	"project/request_control"
	"project/hardware"
	. "project/types"
)

func main() {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var local_id string
	var elevatorPort string
	flag.StringVar(&local_id, "id", "", "id of this peer")
	flag.StringVar(&elevatorPort, "elevatorPort", "15657", "port for elevator connection")
	flag.Parse()

	if local_id == "" {
		println("-id not provided")
		return
	}

	elevio.Init("localhost:"+elevatorPort, N_FLOORS)

	requestsCh := make(chan [N_FLOORS][N_BUTTONS]bool)
	completedRequestCh := make(chan ButtonEvent_t)

	go elevator_control.RunElevatorControl(requestsCh, completedRequestCh)
	go request_control.RunRequestControl(local_id, requestsCh, completedRequestCh)

	select {} //keeps main from exiting without using CPU power
}
