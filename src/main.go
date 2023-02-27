package main

import (
	//"project/single-elevator/elevio"
	"flag"
	nc "project/network-control"
	elevator "project/single-elevator"
	"project/single-elevator/elevio"
	"fmt"
	"os"
	"project/network/localip"
)

func main() {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	requests_chan := make(chan [elevator.N_FLOORS][elevator.N_BUTTONS]bool)
	completed_request_chan := make(chan elevio.ButtonEvent)

	go elevator.Run_elevator(requests_chan, completed_request_chan)
	go nc.RunNetworkControl(id, requests_chan, completed_request_chan)

	for {

	}
}
