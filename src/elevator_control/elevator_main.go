package elevator_control

import (
	req "project/elevator_control/local_requests"
	elevio "project/hardware"
	. "project/types"
	"time"
)

const DOOR_TIMEOUT_SEC = 3
const MOBILITY_TIMOEUT_SEC = 4

func RunElevatorControl(
	requestsCh <-chan [N_FLOORS][N_BUTTONS]bool,
	completedRequestCh chan<- ButtonEvent_t,
) {
	floorSensorCh := make(chan int)
	obstructionSwitCh := make(chan bool)

	go elevio.PollFloorSensor(floorSensorCh)
	go elevio.PollObstructionSwitch(obstructionSwitCh)

	elevator := elevatorInit(floorSensorCh)
	updateElevatorInfo(elevator)
	setElevatorAvailability(true)

	doorTimeout := time.NewTimer(0)
	timerKill(doorTimeout)
	mobilityTimeout := time.NewTimer(0)
	timerKill(mobilityTimeout)

	for {
		select {
		case elevator.Requests = <-requestsCh:

			if elevator.Behaviour != IDLE {
				break
			}

			elevator.Direction, elevator.Behaviour = req.ChooseNewDirectionAndBehavior(elevator)

			switch elevator.Behaviour {
			case DOOR_OPEN:
				elevio.SetDoorOpenLamp(true)

				if elevio.IsObstruction() {
					setElevatorAvailability(false)
				} else {
					timerRestart(doorTimeout, DOOR_TIMEOUT_SEC)
				}
			case MOVING:
				timerRestart(mobilityTimeout, MOBILITY_TIMOEUT_SEC)
				elevio.SetMotorDirection(directionConverter(elevator.Direction))
			}

			updateElevatorInfo(elevator)

		case elevator.Floor = <-floorSensorCh:
			if elevator.Behaviour != MOVING {
				break
			}

			timerRestart(mobilityTimeout, MOBILITY_TIMOEUT_SEC)
			setElevatorAvailability(true)

			elevio.SetFloorIndicator(elevator.Floor)

			if req.ShouldStop(elevator) {
				timerKill(mobilityTimeout)

				elevio.SetMotorDirection(MD_Stop)

				if req.ShouldClearUp(elevator) {
					elevator.Direction = DIR_UP
				} else if req.ShouldClearDown(elevator) {
					elevator.Direction = DIR_DOWN
				} else if req.ShouldClearCab(elevator) {
					// no need to update direction
				} else {
					elevator.Behaviour = IDLE
					updateElevatorInfo(elevator)
					break
				}

				elevio.SetDoorOpenLamp(true)
				elevator.Behaviour = DOOR_OPEN

				if elevio.IsObstruction() {
					setElevatorAvailability(false)
				} else {
					timerRestart(doorTimeout, DOOR_TIMEOUT_SEC)
				}
			}

			updateElevatorInfo(elevator)

		case <-doorTimeout.C:
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
				timerRestart(doorTimeout, DOOR_TIMEOUT_SEC)
			case IDLE:
				elevio.SetDoorOpenLamp(false)
			case MOVING:
				elevio.SetDoorOpenLamp(false)
				timerRestart(mobilityTimeout, MOBILITY_TIMOEUT_SEC)
				elevio.SetMotorDirection(directionConverter(elevator.Direction))
			}

			updateElevatorInfo(elevator)

		case isObstructed := <-obstructionSwitCh:
			if elevator.Behaviour != DOOR_OPEN {
				break
			}

			if isObstructed {
				setElevatorAvailability(false)
				timerKill(doorTimeout)
			} else {
				setElevatorAvailability(true)
				timerRestart(doorTimeout, DOOR_TIMEOUT_SEC)
			}

		case <-mobilityTimeout.C:
			setElevatorAvailability(false)
		}
	}
}
