package single_elevator

const N_FLOORS = 4
const N_BUTTONS = 3
const TIMEOUT_SEC = 3

type Behaviour_t int
const (
	IDLE Behaviour_t = iota
	DOOR_OPEN
	MOVING
)

type Direction_t int
const (
	DIR_UP Direction_t = iota
	DIR_STOP
	DIR_DOWN
)

type Elevator_t struct {
	floor     int
	direction Direction_t
	requests  [N_FLOORS][N_BUTTONS]bool
	behaviour Behaviour_t
}

type DirectionBehaviourPair struct {
    direction   Direction_t;
    behaviour	Behaviour_t;
}
