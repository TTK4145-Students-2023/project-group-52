package printing

import (
	"fmt"
	"project/network/peers"
	. "project/types"
)

func REQ_toString(requests [N_HALL_BUTTONS]Request_t, cabRequest Request_t) string {
	text := ""
	for _, req := range requests {
		for len(req.AwareList) < 3 {
			req.AwareList = append(req.AwareList, " ")
		}
		text = text + "State: " + RS_toString(req.State) + " - Count:" + fmt.Sprintf("%d", req.Count) + " - AwareList:" + fmt.Sprintf("%+v", req.AwareList) + " | "
	}
	for len(cabRequest.AwareList) < 3 {
		cabRequest.AwareList = append(cabRequest.AwareList, " ")
	}
	text = text + "State: " + RS_toString(cabRequest.State) + " - Count:" + fmt.Sprintf("%d", cabRequest.Count) + " - AwareList:" + fmt.Sprintf("%+v", cabRequest.AwareList) + " | "
	return text
}

func RS_toString(state RequestState_t) string {
	switch(state) {
	case COMPLETED:
		return "COM"
	case ASSIGNED:
		return "ASS"
	case NEW:
		return "NEW"
	}
	return "???"
}

func PrintMessage(message NetworkMessage_t) {
	fmt.Printf("id: %+v\n", message.SenderID)
	fmt.Printf("behaviour: %+v\n", Eb_toString(message.Behaviour))
	fmt.Printf("floor: %+v\n", message.Floor)
	fmt.Printf("direction: %+v\n", Ed_toString(message.Direction))
	fmt.Printf("available: %+v\n", message.Available)
	fmt.Printf("    Up                                         Down                                       Cab\n")
	for i, rq := range message.SenderHallRequests {
		fmt.Printf("%d - %s\n", i+1, REQ_toString(rq, message.AllCabRequests[message.SenderID][i]))
	}
	fmt.Printf("###################################################################################################################################|\n")
}

func PrintPeers(p peers.PeerUpdate) {
	fmt.Printf("Peer update:\n")
	fmt.Printf("  Peers:    %q\n", p.Peers)
	fmt.Printf("  New:      %q\n", p.New)
	fmt.Printf("  Lost:     %q\n", p.Lost)
}
