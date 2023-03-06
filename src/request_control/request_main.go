package request_control

import (
	"errors"
	printing "project/debug_printing"
	elev "project/elevator_control"
	"project/hardware/elevio"
	"project/network/bcast"
	"project/network/peers"
	. "project/types"
	"time"
)

const (
	PEER_PORT    = 42069
	MSG_PORT     = 42068
	SEND_TIME_MS = 1000
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

	hallRequests := [N_FLOORS][N_HALL_BUTTONS]Request_t{}
	allCabRequests := make(map[string][N_FLOORS]Request_t)
	latestInfoElevators := make(map[string]ElevatorInfo_t)

	allCabRequests[local_id] = [N_FLOORS]Request_t{}
	{
		floor, behaviour, direction := elev.GetElevatorState()
		latestInfoElevators[local_id] = ElevatorInfo_t{
			Available: true,
			Behaviour: behaviour,
			Floor:     floor,
			Direction: direction,
		}
	}

	for {
		select {
		case btn := <-drv_buttons:
			request := Request_t{}
			if btn.Button == elevio.BT_Cab {
				request = allCabRequests[local_id][btn.Floor]
			} else {
				request = hallRequests[btn.Floor][btn.Button]
			}
			switch request.State {
			case COMPLETED:
				request.State = NEW
				request.AwareList = []string{local_id}

				if is_subset(peerList, request.AwareList) {
					request.State = ASSIGNED
					request.AwareList = []string{local_id}
					if btn.Button == elevio.BT_Cab {
						localCabRequest := allCabRequests[local_id]
						localCabRequest[btn.Floor] = request
						allCabRequests[local_id] = localCabRequest
					} else {
						hallRequests[btn.Floor][btn.Button] = request
					}
					requests_chan <- requestDistributor(hallRequests,allCabRequests,latestInfoElevators,local_id)
					elevio.SetButtonLamp(btn.Button, btn.Floor, true)
				}
			}
			if btn.Button == elevio.BT_Cab {
				localCabRequest := allCabRequests[local_id]
				localCabRequest[btn.Floor] = request
				allCabRequests[local_id] = localCabRequest
			} else {
				hallRequests[btn.Floor][btn.Button] = request
			}
		case btn := <-completed_request_chan:
			request := Request_t{}
			if btn.Button == elevio.BT_Cab {
				request = allCabRequests[local_id][btn.Floor]
			} else {
				request = hallRequests[btn.Floor][btn.Button]
			}
			switch request.State {
			case ASSIGNED:
				request.State = COMPLETED
				request.AwareList = []string{local_id}
				request.Count++
				elevio.SetButtonLamp(btn.Button, btn.Floor, false)
			}
			if btn.Button == elevio.BT_Cab {
				localCabRequest := allCabRequests[local_id]
				localCabRequest[btn.Floor] = request
				allCabRequests[local_id] = localCabRequest
			} else {
				hallRequests[btn.Floor][btn.Button] = request
			}
		case <-send_timer.C:
			send_timer.Reset(SEND_TIME_MS * time.Millisecond)
			floor, behaviour, direction := elev.GetElevatorState()
			newMessage := NetworkMessage_t{
				Sender_id:          local_id,
				Available:          true,
				Behaviour:          behaviour,
				Floor:              floor,
				Direction:          direction,
				SenderHallRequests: hallRequests,
				AllCabRequests:     allCabRequests,
			}
			messageTx <- newMessage
		case p := <-peerUpdateCh:
			printing.PrintPeers(p)
			peerList = p.Peers
		case message := <-messageRx:
			if message.Sender_id == local_id {
				printing.PrintMessage(message, local_id)
				break
			}

			isRequestsUpdated := false

			for id, senderCabRequests := range message.AllCabRequests {
				if _, id_exist := allCabRequests[id]; !id_exist {
					allCabRequests[id] = senderCabRequests
					continue
				}
				for floor := 0; floor < N_FLOORS; floor++ {
					if !shouldAcceptMessage(allCabRequests[id][floor], senderCabRequests[floor]){
						continue
					}
					accepted_request := senderCabRequests[floor]
					accepted_request.AwareList = addToAwareList(accepted_request.AwareList, local_id)
					if id == local_id {
						isRequestsUpdated = true

						switch accepted_request.State {
						case NEW:
							if is_subset(peerList, accepted_request.AwareList) {
								accepted_request.State = ASSIGNED
								accepted_request.AwareList = []string{local_id}
								elevio.SetButtonLamp(elevio.BT_Cab, floor, true)
							}
						case ASSIGNED:
							elevio.SetButtonLamp(elevio.BT_Cab, floor, true)
						}
					} else {
						switch accepted_request.State {
						case NEW:
							if is_subset(peerList, accepted_request.AwareList) {
								accepted_request.State = ASSIGNED
								accepted_request.AwareList = []string{local_id}
							}
						}
					}
					cabRequests := allCabRequests[id]
					cabRequests[floor] = accepted_request
					allCabRequests[id] = cabRequests
				}
			}

			for floor := 0; floor < N_FLOORS; floor++ {
				for btn := 0; btn < N_HALL_BUTTONS; btn++ {
					if !shouldAcceptMessage(hallRequests[floor][btn], message.SenderHallRequests[floor][btn]) {
						continue
					}
					isRequestsUpdated = true

					accepted_request := message.SenderHallRequests[floor][btn]
					accepted_request.AwareList = addToAwareList(accepted_request.AwareList, local_id)

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

					hallRequests[floor][btn] = accepted_request
				}
			}
			if isRequestsUpdated {
				requests_chan <- requestDistributor(hallRequests,allCabRequests,latestInfoElevators,local_id)
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

func requestDistributor(
	hallRequests [N_FLOORS][N_HALL_BUTTONS]Request_t,
	allCabRequest map[string][N_FLOORS]Request_t,
	latestInfoElevators map[string]ElevatorInfo_t,
	local_id string,
) [N_FLOORS][N_BUTTONS]bool {

	bool_requests := [N_FLOORS][N_BUTTONS]bool{}
	localCabRequest := allCabRequest[local_id]

	for floor_num := 0; floor_num < N_FLOORS; floor_num++ {
		for button_num := 0; button_num < N_HALL_BUTTONS; button_num++ {
			if hallRequests[floor_num][button_num].State == ASSIGNED {
				bool_requests[floor_num][button_num] = true
			}
		}
		if localCabRequest[floor_num].State == ASSIGNED {
			bool_requests[floor_num][elevio.BT_Cab] = true
		}
	}
	return bool_requests
}

func getCabRequestsIndex(id string, allCabRequests []CabRequests_t) (int, error) {
	for i, req := range allCabRequests {
		if req.Id == id {
			return i, nil
		}
	}
	return 0, errors.New("Id not found")
}
