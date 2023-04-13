/*
TTK4145 elevator project
spring 2023
Anders Mørk, Erlend Dahle, Filip Strømstad
*/

package main

import (
	"flag"
	"project/elevator_control"
	"project/request_control"
	"project/hardware"
	. "project/types"
)

func main() {
	// program needs ID to run, example:
	// go run main.go -id=1

	var localID string
	var elevatorPort string
	flag.StringVar(&localID, "id", "", "id of this peer")
	flag.StringVar(&elevatorPort, "elevatorPort", "15657", "port for elevator connection")
	flag.Parse()

	if localID == "" {
		println("-id not provided")
		return
	}

	elevio.Init("localhost:"+elevatorPort, N_FLOORS)

	requestsCh := make(chan [N_FLOORS][N_BUTTONS]bool)
	completedRequestCh := make(chan ButtonEvent_t)

	go elevator_control.RunElevatorControl(requestsCh, completedRequestCh)
	go request_control.RunRequestControl(localID, requestsCh, completedRequestCh)

	select {} //keeps main from exiting without using CPU power
}
