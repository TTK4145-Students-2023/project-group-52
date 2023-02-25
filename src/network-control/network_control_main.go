package network_control

import (
	"fmt"
	"project/network"
	. "project/network-control/types"
	elev "project/single-elevator"
	"project/single-elevator/elevio"
	"time"
)

func RunNetworkControl(
	id string,
	requests_chan chan<- [elev.N_FLOORS][elev.N_BUTTONS]bool,
	completed_request_chan <-chan elevio.ButtonEvent,
) {
	drv_buttons := make(chan elevio.ButtonEvent)
	go elevio.PollButtons(drv_buttons)
	messageTx := make(chan NetworkMessage_t)
	messageRx := make(chan NetworkMessage_t)

	go network.RunNetwork(id, messageTx, messageRx)

	send_timer := time.NewTimer(SEND_TIME_SEC * time.Second)

	peerList := []string{id}

	requests := [elev.N_FLOORS][elev.N_BUTTONS]Request_t{}

	for {
		select {
		case btn := <-drv_buttons:
			request := &requests[btn.Floor][btn.Button]
			switch request.State {
			case COMPLETED:
				request.State = NEW
				request.AwareList = []string{id}

				if is_subset(peerList, request.AwareList) {
					request.State = ASSIGNED
					request.AwareList = []string{id}
					requests_chan <- requestDistributor(requests)
					elevio.SetButtonLamp(btn.Button, btn.Floor, true)
				}
			}
		case btn := <-completed_request_chan:
			request := &requests[btn.Floor][btn.Button]
			switch request.State {
			case ASSIGNED:
				request.State = COMPLETED
				request.AwareList = []string{id}
				request.Count++
				elevio.SetButtonLamp(btn.Button, btn.Floor, false)
			}
		case <-send_timer.C:
			send_timer.Reset(SEND_TIME_SEC * time.Second)
			floor, behaviour, direction := elev.GetElevatorState()
			newMessage := NetworkMessage_t{
				Sender_id:    	 id,
				Available:    	 true,
				Behaviour:    	 behaviour,
				Floor:        	 floor,
				Direction:    	 direction,
				Sender_requests: requests,
				ExternalCabRequests:  []CabRequests_t{},
			}
			messageTx <- newMessage
		case message := <-messageRx:
			fmt.Printf("id: %+v\n", message.Sender_id)
			fmt.Printf("behaviour: %+v\n", message.Behaviour)
			fmt.Printf("floor: %+v\n", message.Floor)
			fmt.Printf("direction: %+v\n", message.Direction)
			fmt.Printf("requests: %+v\n", message.Sender_requests)
			fmt.Printf("-------------------------\n")
		}

		//case: mottar melding fra andre noder
		//gå gjennom alle knappene og kjør FSM på dem
	}
}

// move in different module
func is_subset(subset []string, superset []string) bool {
	checkset := make(map[string]bool)
	for _, element := range subset {
		checkset[element] = true
	}
	for _, value := range superset {
		if checkset[value] {
			delete(checkset, value)
		}
	}
	return len(checkset) == 0 //this implies that set is subset of superset
}

func requestDistributor(requests [elev.N_FLOORS][elev.N_BUTTONS]Request_t) [elev.N_FLOORS][elev.N_BUTTONS]bool {
	bool_requests := [elev.N_FLOORS][elev.N_BUTTONS]bool{}
	for floor_num := 0; floor_num < elev.N_FLOORS; floor_num++ {
		for button_num := 0; button_num < elev.N_BUTTONS; button_num++ {

			if requests[floor_num][button_num].State == ASSIGNED {
				bool_requests[floor_num][button_num] = true
			}
		}
	}
	return bool_requests
}
