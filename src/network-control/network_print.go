package network_control

import (
	"fmt"
	. "project/network-control/types"
	"project/network/peers"
	elev "project/single-elevator"
)

func REQ_toString(requests [elev.N_BUTTONS]Request_t)string{
	text := ""
	for _, req := range requests {
		for len(req.AwareList) < 3 {
			req.AwareList = append(req.AwareList, " ")
		}
		text = text + "State: " + RS_toString(req.State) + " - Count:" + fmt.Sprintf("%d", req.Count) + " - AwareList:" + fmt.Sprintf("%+v",req.AwareList) + " | "
	}
	return text
}

func RS_toString(state RequestState_t) string {
	if(state==COMPLETED){
		return "COM"
	}
	if(state==ASSIGNED){
		return "ASS"
	}
	if(state==NEW){
		return "NEW"
	}
	if(state == UNKNOWN){
		return "UNK"
	}
	return "???"
}

func printMessage (message NetworkMessage_t) {
	fmt.Printf("id: %+v\n", message.Sender_id)
	fmt.Printf("behaviour: %+v\n", elev.Eb_toString(message.Behaviour))
	fmt.Printf("floor: %+v\n", message.Floor)
	fmt.Printf("direction: %+v\n", elev.Ed_toString(message.Direction))
	fmt.Printf("    Up                                         Down                                       Cab\n")
	for i,rq := range message.Sender_requests{
		fmt.Printf("%d - %s\n", i+1, REQ_toString(rq))
	}
	fmt.Printf("###################################################################################################################################|\n")
}

func printPeers (p peers.PeerUpdate){
	fmt.Printf("Peer update:\n")
	fmt.Printf("  Peers:    %q\n", p.Peers)
	fmt.Printf("  New:      %q\n", p.New)
	fmt.Printf("  Lost:     %q\n", p.Lost)
}