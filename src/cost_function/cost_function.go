package cost_function

import (
	"project/hardware/elevio"
	. "project/types"
)

func RequestDistributor(
	hallRequests [N_FLOORS][N_HALL_BUTTONS]Request_t,
	allCabRequest map[string][N_FLOORS]Request_t,
	latestInfoElevators map[string]ElevatorInfo_t,
	local_id string,
) [N_FLOORS][N_BUTTONS]bool {

	bool_requests := [N_FLOORS][N_BUTTONS]bool{}
	localCabRequest := allCabRequest[local_id]

	for floor_num := 0; floor_num < N_FLOORS; floor_num++ {
		for button_num := 0; button_num < N_HALL_BUTTONS; button_num++ {
			if hallRequests[floor_num][button_num].State == ASSIGNED {
				bool_requests[floor_num][button_num] = true
			}
		}
		if localCabRequest[floor_num].State == ASSIGNED {
			bool_requests[floor_num][elevio.BT_Cab] = true
		}
	}
	return bool_requests
}
