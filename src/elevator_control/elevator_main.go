package elevator_control

import (
	req "project/elevator_control/local_requests"
	"project/hardware"
	. "project/types"
	"time"
	"project/music"
)

const DOOR_TIMEOUT_SEC = 3
const MOBILITY_TIMOEUT_SEC = 4

var shared_state ElevatorSharedState_t

func RunElevatorControl(
	requestsCh <-chan [N_FLOORS][N_BUTTONS]bool,
	completedRequestCh chan<- ButtonEvent_t,
) {
	drv_Floors := make(chan int)
	drv_obstr := make(chan bool)

	go elevio.PollFloorSensor(drv_Floors)
	go elevio.PollObstructionSwitch(drv_obstr)

	elevator := elevatorInit(drv_Floors)
	updateElevatorState(elevator)
	setElevatorAvailability(true)

	door_timeout := time.NewTimer(0)
	timerKill(door_timeout)
	mobility_timeout := time.NewTimer(0)
	timerKill(mobility_timeout)

	musicEnabaleCh := make(chan bool)
	go music.MusicPlayer(musicEnabaleCh)

	for {
		select {
		case requests := <-requestsCh:
			elevator.Requests = requests

			switch elevator.Behaviour {
			case IDLE:
				elevator.Direction, elevator.Behaviour = req.ChooseNewDirectionAndBehavior(elevator)

				switch elevator.Behaviour {
				case DOOR_OPEN:
					elevio.SetDoorOpenLamp(true)
					handleObstruction(door_timeout)
				case MOVING:
					timerRestart(mobility_timeout, MOBILITY_TIMOEUT_SEC)
					elevio.SetMotorDirection(directionConverter(elevator.Direction))
					musicEnabaleCh <- true
				}
			}
			updateElevatorState(elevator)
		case newFloor := <-drv_Floors:
			timerRestart(mobility_timeout, MOBILITY_TIMOEUT_SEC)
			setElevatorAvailability(true)

			elevator.Floor = newFloor
			elevio.SetFloorIndicator(elevator.Floor)

			if elevator.Behaviour == MOVING && req.ShouldStop(elevator) {
				timerKill(mobility_timeout)

				elevio.SetMotorDirection(MD_Stop)
				musicEnabaleCh <- false

				if req.ShouldClearCab(elevator) || req.ShouldClearUp(elevator) || req.ShouldClearDown(elevator) {
					elevio.SetDoorOpenLamp(true)
					elevator.Behaviour = DOOR_OPEN
					handleObstruction(door_timeout)
				} else {
					elevator.Behaviour = IDLE
				}
			}
			updateElevatorState(elevator)
		case <-door_timeout.C:
			if elevator.Behaviour != DOOR_OPEN {
				break
			}

			if req.ShouldClearCab(elevator) {
				elevator.Requests[elevator.Floor][BT_Cab] = false
				completedRequestCh <- ButtonEvent_t{Floor: elevator.Floor, Button: BT_Cab}
			}
			if req.ShouldClearUp(elevator) {
				elevator.Requests[elevator.Floor][BT_HallUp] = false
				completedRequestCh <- ButtonEvent_t{Floor: elevator.Floor, Button: BT_HallUp}
			} else if req.ShouldClearDown(elevator) {
				elevator.Requests[elevator.Floor][BT_HallDown] = false
				completedRequestCh <- ButtonEvent_t{Floor: elevator.Floor, Button: BT_HallDown}
			}

			elevator.Direction, elevator.Behaviour = req.ChooseNewDirectionAndBehavior(elevator)

			switch elevator.Behaviour {
			case DOOR_OPEN:
				handleObstruction(door_timeout)
			case IDLE:
				elevio.SetDoorOpenLamp(false)
			case MOVING:
				elevio.SetDoorOpenLamp(false)
				timerRestart(mobility_timeout, MOBILITY_TIMOEUT_SEC)
				elevio.SetMotorDirection(directionConverter(elevator.Direction))
				musicEnabaleCh <- true
			}
			updateElevatorState(elevator)

		case isObstructed := <-drv_obstr:
			if elevator.Behaviour != DOOR_OPEN {
				break
			}

			if isObstructed {
				setElevatorAvailability(false)
				timerKill(door_timeout)
			} else {
				setElevatorAvailability(true)
				timerRestart(door_timeout, DOOR_TIMEOUT_SEC)
			}
		case <-mobility_timeout.C:
			println("\nMOBILITY TIMEOUT\n")
			setElevatorAvailability(false)
		}
	}
}
