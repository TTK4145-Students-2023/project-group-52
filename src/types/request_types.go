package types

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
	Requests [N_FLOORS]Request_t
}

type NetworkMessage_t struct {
	Sender_id           string
	Available           bool
	Behaviour           Behaviour_t
	Floor               int
	Direction           Direction_t
	Sender_requests     [N_FLOORS][N_BUTTONS]Request_t
	ExternalCabRequests []CabRequests_t
}
