package types

import "sync"

const N_FLOORS = 4
const N_BUTTONS = 3
const N_HALL_BUTTONS = 2

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
	Floor     int
	Direction Direction_t
	Requests  [N_FLOORS][N_BUTTONS]bool
	Behaviour Behaviour_t
}

type ElevatorSharedInfo_t struct {
	Mutex     sync.RWMutex
	Available bool
	Behaviour Behaviour_t
	Direction Direction_t
	Floor     int
}
