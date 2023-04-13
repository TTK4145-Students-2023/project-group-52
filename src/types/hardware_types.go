package types

type MotorDirection_t int

const (
	MD_Up   MotorDirection_t = 1
	MD_Down MotorDirection_t = -1
	MD_Stop MotorDirection_t = 0
)

type ButtonType_t int

const (
	BT_HallUp ButtonType_t = iota
	BT_HallDown
	BT_Cab
)

type ButtonEvent_t struct {
	Floor  int
	Button ButtonType_t
}
