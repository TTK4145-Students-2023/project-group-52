package network_control

import (
	elev "project/single-elevator"
	"project/single-elevator/elevio"
)

type RequestState_t int

const (
	COMPLETED RequestState_t = iota
	NEW
	ASSIGNED
	UNKNOWN
)

type Request_t struct {
	State     RequestState_t
	Count     int
	AwareList []string
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

func RunNetworkControl(
	requests_chan chan<- [elev.N_FLOORS][elev.N_BUTTONS]bool,
	completed_request_chan <-chan elevio.ButtonEvent,
) {
	drv_buttons := make(chan elevio.ButtonEvent)
	go elevio.PollButtons(drv_buttons)

	our_id := "us"

	peerList := []string{our_id}

	requests := [elev.N_FLOORS][elev.N_BUTTONS]Request_t{}

	for {
		select {
		case btn := <-drv_buttons:
			request := &requests[btn.Floor][btn.Button]
			switch request.State {
			case COMPLETED:
				request.State = NEW
				request.AwareList = []string{our_id}

				if is_subset(peerList, request.AwareList) {
					request.State = ASSIGNED
					request.AwareList = []string{our_id}
					requests_chan <- requestDistributor(requests)
					elevio.SetButtonLamp(btn.Button,btn.Floor,true)
				}
			}
		case btn := <-completed_request_chan:
			request := &requests[btn.Floor][btn.Button]
			switch request.State {
			case ASSIGNED:
				request.State = COMPLETED
				request.AwareList = []string{our_id}
				request.Count++
				elevio.SetButtonLamp(btn.Button,btn.Floor,false)
			}
		}
		//case: mottar melding fra andre noder
		//gå gjennom alle knappene og kjør FSM på dem
	}
}
