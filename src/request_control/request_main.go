package request_control

import (
	//"fmt"
	"project/cost_function"
	//"project/debug_printing"
	elev "project/elevator_control"
	"project/hardware"
	"project/network/bcast"
	"project/network/peers"
	. "project/types"
	"time"
)

const (
	PEER_PORT          = 30052
	MSG_PORT           = 30051
	SEND_TIME_MS       = 200
	DISTRIBUTE_TIME_MS = 1000
)

func RunRequestControl(
	local_id string,
	requestsCh chan<- [N_FLOORS][N_BUTTONS]bool,
	completedRequestCh <-chan ButtonEvent_t,
) {
	buttonEventCh := make(chan ButtonEvent_t)
	go elevio.PollButtons(buttonEventCh)

	messageTx := make(chan NetworkMessage_t)
	messageRx := make(chan NetworkMessage_t)
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)

	go peers.Transmitter(PEER_PORT, local_id, peerTxEnable)
	go peers.Receiver(PEER_PORT, peerUpdateCh)
	go bcast.Transmitter(MSG_PORT, messageTx)
	go bcast.Receiver(MSG_PORT, messageRx)

	sendTicker := time.NewTicker(SEND_TIME_MS * time.Millisecond)
	distributeTicker := time.NewTicker(DISTRIBUTE_TIME_MS * time.Millisecond)

	peerList := []string{}
	connectedToNetwork := false

	hallRequests := [N_FLOORS][N_HALL_BUTTONS]Request_t{}
	allCabRequests := make(map[string][N_FLOORS]Request_t)
	latestInfoElevators := make(map[string]ElevatorInfo_t)

	allCabRequests[local_id] = [N_FLOORS]Request_t{}
	{
		available, behaviour, direction, floor := elev.GetElevatorState()
		latestInfoElevators[local_id] = ElevatorInfo_t{
			Available: available,
			Behaviour: behaviour,
			Floor:     floor,
			Direction: direction,
		}
	}

	for {
		select {
		case btn := <-buttonEventCh:
			request := Request_t{}
			if btn.Button == BT_Cab {
				request = allCabRequests[local_id][btn.Floor]
			} else {
				if !connectedToNetwork {
					break
				}
				request = hallRequests[btn.Floor][btn.Button]
			}

			switch request.State {
			case COMPLETED:
				request.State = NEW
				request.AwareList = []string{local_id}
				if is_subset(peerList, request.AwareList) {
					request.State = ASSIGNED
					request.AwareList = []string{local_id}
					elevio.SetButtonLamp(btn.Button, btn.Floor, true)
				}
			case NEW:
				if is_subset(peerList, request.AwareList) {
					request.State = ASSIGNED
					request.AwareList = []string{local_id}
					elevio.SetButtonLamp(btn.Button, btn.Floor, true)
				}
			}

			if btn.Button == BT_Cab {
				localCabRequest := allCabRequests[local_id]
				localCabRequest[btn.Floor] = request
				allCabRequests[local_id] = localCabRequest
			} else {
				hallRequests[btn.Floor][btn.Button] = request
			}

		case btn := <-completedRequestCh:
			request := Request_t{}
			if btn.Button == BT_Cab {
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
			if btn.Button == BT_Cab {
				localCabRequest := allCabRequests[local_id]
				localCabRequest[btn.Floor] = request
				allCabRequests[local_id] = localCabRequest
			} else {
				hallRequests[btn.Floor][btn.Button] = request
			}
		case <-sendTicker.C:
			available, behaviour, direction, floor := elev.GetElevatorState()

			latestInfoElevators[local_id] = ElevatorInfo_t{
				Available: available,
				Behaviour: behaviour,
				Floor:     floor,
				Direction: direction,
			}
			newMessage := NetworkMessage_t{
				Sender_id:          local_id,
				Available:          available,
				Behaviour:          behaviour,
				Floor:              floor,
				Direction:          direction,
				SenderHallRequests: hallRequests,
				AllCabRequests:     allCabRequests,
			}

			if connectedToNetwork {
				messageTx <- newMessage
			}
		case <-distributeTicker.C:
			select{
			case requestsCh <- cost_function.RequestDistributor(hallRequests, allCabRequests, latestInfoElevators, peerList, local_id):
			default:
				// Avoid deadlock
			}

		case p := <-peerUpdateCh:
			peerList = p.Peers

			if p.New == local_id {
				connectedToNetwork = true
			}

			if is_subset([]string{local_id}, p.Lost) {
				connectedToNetwork = false
			}

		case message := <-messageRx:
			if message.Sender_id == local_id {
				// fmt.Printf("%+v\n",time.Now())
				// fmt.Printf("Peerlist: %+v\n", peerList)
				// printing.PrintMessage(message)
				break
			}

			if !connectedToNetwork { //avoid race-conditions with peer messages
				break
			}

			latestInfoElevators[message.Sender_id] = ElevatorInfo_t{
				Available: message.Available,
				Behaviour: message.Behaviour,
				Direction: message.Direction,
				Floor:     message.Floor,
			}

			for id, senderCabRequests := range message.AllCabRequests {
				if _, id_exist := allCabRequests[id]; !id_exist {
					for floor := range senderCabRequests {
						senderCabRequests[floor].AwareList = addToAwareList(senderCabRequests[floor].AwareList, local_id)
					}
					allCabRequests[id] = senderCabRequests
					continue
				}
				for floor := 0; floor < N_FLOORS; floor++ {
					if !shouldAcceptMessage(allCabRequests[id][floor], senderCabRequests[floor]) {
						continue
					}

					accepted_request := senderCabRequests[floor]
					accepted_request.AwareList = addToAwareList(accepted_request.AwareList, local_id)

					if accepted_request.State == NEW && is_subset(peerList, accepted_request.AwareList) {
						accepted_request.State = ASSIGNED
						accepted_request.AwareList = []string{local_id}
					}

					if id == local_id && accepted_request.State == ASSIGNED {
						elevio.SetButtonLamp(BT_Cab, floor, true)
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

					accepted_request := message.SenderHallRequests[floor][btn]
					accepted_request.AwareList = addToAwareList(accepted_request.AwareList, local_id)

					switch accepted_request.State {
					case COMPLETED:
						elevio.SetButtonLamp(ButtonType_t(btn), floor, false)
					case NEW:
						elevio.SetButtonLamp(ButtonType_t(btn), floor, false)
						if is_subset(peerList, accepted_request.AwareList) {
							accepted_request.State = ASSIGNED
							accepted_request.AwareList = []string{local_id}
							elevio.SetButtonLamp(ButtonType_t(btn), floor, true)
						}
					case ASSIGNED:
						elevio.SetButtonLamp(ButtonType_t(btn), floor, true)
					}

					hallRequests[floor][btn] = accepted_request
				}
			}
		}
	}
}