package network_types

import (
	elev "project/single-elevator"
)

const (
	SEND_TIME_MS = 400
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

type CabRequests_t struct {
	Id       string
	Requests [elev.N_FLOORS]Request_t
}

type NetworkMessage_t struct {
	Sender_id           string
	Available           bool
	Behaviour           elev.Behaviour_t
	Floor               int
	Direction           elev.Direction_t
	Sender_requests     [elev.N_FLOORS][elev.N_BUTTONS]Request_t
	ExternalCabRequests []CabRequests_t
}
