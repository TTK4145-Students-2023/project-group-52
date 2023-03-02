package request_control

import (
	. "project/types"
	"project/network/bcast"
	"project/network/peers"
	elev "project/elevator_control"
	"project/hardware/elevio"
	"time"
	printing "project/debug_printing"
)

const (
	PEER_PORT    = 42069
	MSG_PORT     = 42068
	SEND_TIME_MS = 400
)

func RunRequestControl(
	local_id string,
	requests_chan chan<- [N_FLOORS][N_BUTTONS]bool,
	completed_request_chan <-chan elevio.ButtonEvent,
) {
	drv_buttons := make(chan elevio.ButtonEvent)
	go elevio.PollButtons(drv_buttons)

	messageTx := make(chan NetworkMessage_t)
	messageRx := make(chan NetworkMessage_t)
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)

	go peers.Transmitter(PEER_PORT, local_id, peerTxEnable)
	go peers.Receiver(PEER_PORT, peerUpdateCh)

	go bcast.Transmitter(MSG_PORT, messageTx)
	go bcast.Receiver(MSG_PORT, messageRx)

	send_timer := time.NewTimer(SEND_TIME_MS * time.Millisecond)

	peerList := []string{local_id}

	requests := [N_FLOORS][N_BUTTONS]Request_t{}

	for {
		select {
		case btn := <-drv_buttons:
			request := &requests[btn.Floor][btn.Button]
			switch request.State {
			case COMPLETED:
				request.State = NEW
				request.AwareList = []string{local_id}

				if is_subset(peerList, request.AwareList) {
					request.State = ASSIGNED
					request.AwareList = []string{local_id}
					requests_chan <- requestDistributor(requests)
					elevio.SetButtonLamp(btn.Button, btn.Floor, true)
				}
			}
		case btn := <-completed_request_chan:
			request := &requests[btn.Floor][btn.Button]
			switch request.State {
			case ASSIGNED:
				request.State = COMPLETED
				request.AwareList = []string{local_id}
				request.Count++
				elevio.SetButtonLamp(btn.Button, btn.Floor, false)
			}
		case <-send_timer.C:
			send_timer.Reset(SEND_TIME_MS * time.Millisecond)
			floor, behaviour, direction := elev.GetElevatorState()
			newMessage := NetworkMessage_t{
				Sender_id:           local_id,
				Available:           true,
				Behaviour:           behaviour,
				Floor:               floor,
				Direction:           direction,
				Sender_requests:     requests,
				ExternalCabRequests: []CabRequests_t{},
			}
			messageTx <- newMessage
		case p := <-peerUpdateCh:
			printing.PrintPeers(p)
			peerList = p.Peers
		case message := <-messageRx:
			if message.Sender_id == local_id {
				printing.PrintMessage(message)
				break
			}

			isRequestsUpdated := false

			for floor := 0; floor < N_FLOORS; floor++ {
				for btn := 0; btn < N_BUTTONS; btn++ {
					if !shouldAcceptMessage(requests[floor][btn], message.Sender_requests[floor][btn]) {
						continue
					}

					accepted_request := message.Sender_requests[floor][btn]

					accepted_request.AwareList = addToAwareList(accepted_request.AwareList, local_id)

					isRequestsUpdated = true

					switch accepted_request.State {
					case COMPLETED:
						elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
					case NEW:
						elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
						if is_subset(peerList, accepted_request.AwareList) {
							accepted_request.State = ASSIGNED
							accepted_request.AwareList = []string{local_id}
							elevio.SetButtonLamp(elevio.ButtonType(btn), floor, true)
						}
					case ASSIGNED:
						elevio.SetButtonLamp(elevio.ButtonType(btn), floor, true)
					}

					requests[floor][btn] = accepted_request
				}
			}
			if isRequestsUpdated {
				requests_chan <- requestDistributor(requests)
			}
		}

	}
}

// move in different module
func shouldAcceptMessage(local_request Request_t, message_request Request_t) bool {
	if message_request.State == UNKNOWN {
		return false
	}
	if local_request.State == UNKNOWN {
		return true
	}
	if message_request.Count < local_request.Count {
		return false
	}
	if message_request.Count > local_request.Count {
		return true
	}
	if message_request.State == local_request.State && is_subset(message_request.AwareList, local_request.AwareList) {
		// count is equal
		return false
	}

	switch local_request.State {
	case COMPLETED:
		switch message_request.State {
		case COMPLETED:
			return true
		case NEW:
			return true
		case ASSIGNED:
			println("FROM COMPLETED TO ASSIGNED (should not happen)")
			return true
		}
	case NEW:
		switch message_request.State {
		case COMPLETED:
			return false
		case NEW:
			return true
		case ASSIGNED:
			return true
		}
	case ASSIGNED:
		switch message_request.State {
		case COMPLETED:
			return false
		case NEW:
			return false
		case ASSIGNED:
			return true
		}
	}
	print("shouldAcceptMessage() did not return")
	return false
}

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

func addToAwareList(AwareList []string, id string) []string {
	for i := range AwareList {
		if AwareList[i] == id {
			return AwareList
		}
	}
	return append(AwareList, id)
}

func requestDistributor(requests [N_FLOORS][N_BUTTONS]Request_t) [N_FLOORS][N_BUTTONS]bool {
	bool_requests := [N_FLOORS][N_BUTTONS]bool{}
	for floor_num := 0; floor_num < N_FLOORS; floor_num++ {
		for button_num := 0; button_num < N_BUTTONS; button_num++ {

			if requests[floor_num][button_num].State == ASSIGNED {
				bool_requests[floor_num][button_num] = true
			}
		}
	}
	return bool_requests
}
