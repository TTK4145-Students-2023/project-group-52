package types

type MotorDirection_t int

const (
	MD_Up   MotorDirection_t = 1
	MD_Down                = -1
	MD_Stop                = 0
)

type ButtonType_t int

const (
	BT_HallUp   ButtonType_t = 0
	BT_HallDown            = 1
	BT_Cab                 = 2
)

type ButtonEvent_t struct {
	Floor  int
	Button ButtonType_t
}