package types

type RequestState_t int

const (
	COMPLETED RequestState_t = iota
	NEW
	ASSIGNED
)

type Request_t struct {
	State     RequestState_t
	Count     int
	AwareList []string
}

type ElevatorInfo_t struct {
	Available bool
	Behaviour Behaviour_t
	Direction Direction_t
	Floor     int
}

type NetworkMessage_t struct {
	Sender_id          string
	Available          bool
	Behaviour          Behaviour_t
	Direction          Direction_t
	Floor              int
	SenderHallRequests [N_FLOORS][N_HALL_BUTTONS]Request_t
	AllCabRequests     map[string][N_FLOORS]Request_t
}
