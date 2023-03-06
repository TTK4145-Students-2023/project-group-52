package cost_function

import (
	"encoding/json"
	"fmt"
	"os/exec"
	. "project/types"
)




type HRAElevState struct {
    Behavior    string      `json:"behaviour"`
    Floor       int         `json:"floor"` 
    Direction   string      `json:"direction"`
    CabRequests [N_FLOORS]bool      `json:"cabRequests"`
}

type HRAInput struct {
    HallRequests    [N_FLOORS][2]bool           `json:"hallRequests"`
    States          map[string]HRAElevState     `json:"states"`
}



func RequestDistributor(
	hallRequests [N_FLOORS][N_HALL_BUTTONS]Request_t,
	allCabRequests map[string][N_FLOORS]Request_t,
	latestInfoElevators map[string]ElevatorInfo_t,
	local_id string,
) [N_FLOORS][N_BUTTONS]bool {

	hraExecutablePath := "hall_request_assigner"

	boolHallRequests := [N_FLOORS][N_HALL_BUTTONS]bool{}
	for floor_num := 0; floor_num < N_FLOORS; floor_num++ {
		for button_num := 0; button_num < N_HALL_BUTTONS; button_num++ {
			if hallRequests[floor_num][button_num].State == ASSIGNED {
				boolHallRequests[floor_num][button_num] = true
			}
		}
	}

	inputStates := map[string]HRAElevState{}

	for id, cabRequests := range allCabRequests {
		elevatorInfo, ok := latestInfoElevators[id]
		if !ok {
			fmt.Println("id not in latestInfo: ", ok)
			return [N_FLOORS][N_BUTTONS]bool{}
		}
		
		if !elevatorInfo.Available {
			continue 
		}

		boolCabRequests := [N_FLOORS]bool{}
		for floor_num := 0; floor_num < N_FLOORS; floor_num++ {
			if cabRequests[floor_num].State == ASSIGNED {
				boolCabRequests[floor_num] = true
			}
		}
		inputStates[id] = HRAElevState{
			Behavior: behaviourToString(elevatorInfo.Behaviour),
			Floor: elevatorInfo.Floor,
			Direction: directionToString(elevatorInfo.Direction),
			CabRequests: boolCabRequests,
		}

	}


    input := HRAInput{
        HallRequests: boolHallRequests,
        States: inputStates,
    }

    jsonBytes, err := json.Marshal(input)
    if err != nil {
        fmt.Println("json.Marshal error: ", err)
        return [N_FLOORS][N_BUTTONS]bool{}
    }
    
    ret, err := exec.Command(hraExecutablePath, "-i", string(jsonBytes), "--includeCab").CombinedOutput()
    if err != nil {
        fmt.Println("exec.Command error: ", err)
        fmt.Println(string(ret))
        return [N_FLOORS][N_BUTTONS]bool{}
    }
    
    output := new(map[string][N_FLOORS][N_BUTTONS]bool)
    err = json.Unmarshal(ret, &output)
    if err != nil {
        fmt.Println("json.Unmarshal error: ", err)
        return [N_FLOORS][N_BUTTONS]bool{}
    }
        
    //fmt.Printf("output: \n")
    //for k, v := range *output {
    //    fmt.Printf("%6v :  %+v\n", k, v)
    //}

	return (*output)[local_id] 
}

func behaviourToString(b Behaviour_t) string {
	switch(b){
	case IDLE:
		return "idle"
	case MOVING:
		return "moving"
	case DOOR_OPEN:
		return "doorOpen"
	}
	return "idle"
}

func directionToString(d Direction_t) string {
	switch d {
	case DIR_DOWN:
		return "down"
	case DIR_UP:
		return "up"
	case DIR_STOP:
		return "stop"
	}
	return "stop"
}