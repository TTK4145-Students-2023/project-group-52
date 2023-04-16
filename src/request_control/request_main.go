package request_control

import (
	printing "project/debug_printing"
	elev "project/elevator_control"
	elevio "project/hardware"
	"project/network/bcast"
	"project/network/peers"
	"project/request_control/request_assigner"
	. "project/types"
	"time"
)

const (
	PEER_PORT               = 30052
	MSG_PORT                = 30051
	SEND_TIME_MS            = 200
	ASSIGN_REQUESTS_TIME_MS = 1000
)

func RunRequestControl(
	localID string,
	requestsCh chan<- [N_FLOORS][N_BUTTONS]bool,
	completedRequestCh <-chan ButtonEvent_t,
) {
	buttonEventCh := make(chan ButtonEvent_t)
	go elevio.PollButtons(buttonEventCh)

	messageTx := make(chan NetworkMessage_t)
	messageRx := make(chan NetworkMessage_t)
	peerUpdateCh := make(chan peers.PeerUpdate)

	go peers.Transmitter(PEER_PORT, localID, nil)
	go peers.Receiver(PEER_PORT, peerUpdateCh)
	go bcast.Transmitter(MSG_PORT, messageTx)
	go bcast.Receiver(MSG_PORT, messageRx)

	sendTicker := time.NewTicker(SEND_TIME_MS * time.Millisecond)
	assignRequestTicker := time.NewTicker(ASSIGN_REQUESTS_TIME_MS * time.Millisecond)

	peerList := []string{}
	connectedToNetwork := false

	hallRequests := [N_FLOORS][N_HALL_BUTTONS]Request_t{}
	allCabRequests := make(map[string][N_FLOORS]Request_t)
	latestInfoElevators := make(map[string]ElevatorInfo_t)

	allCabRequests[localID] = [N_FLOORS]Request_t{}
	latestInfoElevators[localID] = elev.GetElevatorInfo()

	for {
		select {
		case btn := <-buttonEventCh:
			request := Request_t{}
			if btn.Button == BT_Cab {
				request = allCabRequests[localID][btn.Floor]
			} else {
				if !connectedToNetwork {
					break
				}
				request = hallRequests[btn.Floor][btn.Button]
			}

			switch request.State {
			case COMPLETED:
				request.State = NEW
				request.AwareList = []string{localID}
				if isSubset(peerList, request.AwareList) {
					request.State = ASSIGNED
					request.AwareList = []string{localID}
					elevio.SetButtonLamp(btn.Button, btn.Floor, true)
				}
			case NEW:
				if isSubset(peerList, request.AwareList) {
					request.State = ASSIGNED
					request.AwareList = []string{localID}
					elevio.SetButtonLamp(btn.Button, btn.Floor, true)
				}
			}

			if btn.Button == BT_Cab {
				localCabRequest := allCabRequests[localID]
				localCabRequest[btn.Floor] = request
				allCabRequests[localID] = localCabRequest
			} else {
				hallRequests[btn.Floor][btn.Button] = request
			}

		case btn := <-completedRequestCh:
			request := Request_t{}
			if btn.Button == BT_Cab {
				request = allCabRequests[localID][btn.Floor]
			} else {
				request = hallRequests[btn.Floor][btn.Button]
			}

			switch request.State {
			case ASSIGNED:
				request.State = COMPLETED
				request.AwareList = []string{localID}
				request.Count++
				elevio.SetButtonLamp(btn.Button, btn.Floor, false)
			}

			if btn.Button == BT_Cab {
				localCabRequest := allCabRequests[localID]
				localCabRequest[btn.Floor] = request
				allCabRequests[localID] = localCabRequest
			} else {
				hallRequests[btn.Floor][btn.Button] = request
			}

		case <-sendTicker.C:
			info := elev.GetElevatorInfo()
			latestInfoElevators[localID] = info

			newMessage := NetworkMessage_t{
				SenderID:           localID,
				Available:          info.Available,
				Behaviour:          info.Behaviour,
				Floor:              info.Floor,
				Direction:          info.Direction,
				SenderHallRequests: hallRequests,
				AllCabRequests:     allCabRequests,
			}

			if connectedToNetwork {
				messageTx <- newMessage
			}

		case <-assignRequestTicker.C:
			select {
			case requestsCh <- request_assigner.RequestAssigner(hallRequests, allCabRequests, latestInfoElevators, peerList, localID):
			default:
				// Avoid deadlock
			}

		case p := <-peerUpdateCh:
			peerList = p.Peers

			if p.New == localID {
				connectedToNetwork = true
			}

			if isSubset([]string{localID}, p.Lost) {
				connectedToNetwork = false
			}

		case message := <-messageRx:
			if message.SenderID == localID {
				printing.PrintMessage(message)
				break
			}

			if !connectedToNetwork {
				// Not accepting messages until we are on the peerlist
				break
			}

			latestInfoElevators[message.SenderID] = ElevatorInfo_t{
				Available: message.Available,
				Behaviour: message.Behaviour,
				Direction: message.Direction,
				Floor:     message.Floor,
			}

			for id, cabRequests := range message.AllCabRequests {

				if _, idExist := allCabRequests[id]; !idExist {
					// First informaton about this elevator
					for floor := range cabRequests {
						cabRequests[floor].AwareList = addToAwareList(cabRequests[floor].AwareList, localID)
					}
					allCabRequests[id] = cabRequests
					continue
				}

				for floor := 0; floor < N_FLOORS; floor++ {
					if !shouldAcceptRequest(allCabRequests[id][floor], cabRequests[floor]) {
						continue
					}

					acceptedRequest := cabRequests[floor]
					acceptedRequest.AwareList = addToAwareList(acceptedRequest.AwareList, localID)

					if acceptedRequest.State == NEW && isSubset(peerList, acceptedRequest.AwareList) {
						acceptedRequest.State = ASSIGNED
						acceptedRequest.AwareList = []string{localID}
					}

					if id == localID && acceptedRequest.State == ASSIGNED {
						elevio.SetButtonLamp(BT_Cab, floor, true)
					}

					tmpCabRequests := allCabRequests[id]
					tmpCabRequests[floor] = acceptedRequest
					allCabRequests[id] = tmpCabRequests
				}
			}

			for floor := 0; floor < N_FLOORS; floor++ {
				for btn := 0; btn < N_HALL_BUTTONS; btn++ {
					if !shouldAcceptRequest(hallRequests[floor][btn], message.SenderHallRequests[floor][btn]) {
						continue
					}

					acceptedRequest := message.SenderHallRequests[floor][btn]
					acceptedRequest.AwareList = addToAwareList(acceptedRequest.AwareList, localID)

					switch acceptedRequest.State {
					case COMPLETED:
						elevio.SetButtonLamp(ButtonType_t(btn), floor, false)
					case NEW:
						elevio.SetButtonLamp(ButtonType_t(btn), floor, false)
						if isSubset(peerList, acceptedRequest.AwareList) {
							acceptedRequest.State = ASSIGNED
							acceptedRequest.AwareList = []string{localID}
							elevio.SetButtonLamp(ButtonType_t(btn), floor, true)
						}
					case ASSIGNED:
						elevio.SetButtonLamp(ButtonType_t(btn), floor, true)
					}

					hallRequests[floor][btn] = acceptedRequest
				}
			}
		}
	}
}
