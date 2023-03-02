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

type ElevatorInfo_t struct {
	Available   bool
	Behaviour   Behaviour_t
	Floor       int
	Direction   Direction_t
	CabRequests [N_FLOORS]Request_t
}

type NetworkMessage_t struct {
	Sender_id          string
	Available          bool
	Behaviour          Behaviour_t
	Floor              int
	Direction          Direction_t
	SenderHallRequests [N_FLOORS][N_HALL_BUTTONS]Request_t
	AllCabRequests     []CabRequests_t
}
