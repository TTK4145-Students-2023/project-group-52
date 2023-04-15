package request_control

import (
	. "project/types"
)

func shouldAcceptRequest(localRequest Request_t, messageRequest Request_t) bool {
	if messageRequest.Count < localRequest.Count {
		return false
	}
	if messageRequest.Count > localRequest.Count {
		return true
	}

	if messageRequest.State == localRequest.State && isSubset(messageRequest.AwareList, localRequest.AwareList) {
		// no new info
		return false
	}

	switch localRequest.State {
	case COMPLETED:
		switch messageRequest.State {
		case COMPLETED:
			return true
		case NEW:
			return true
		case ASSIGNED:
			return true
		}
	case NEW:
		switch messageRequest.State {
		case COMPLETED:
			return false
		case NEW:
			return true
		case ASSIGNED:
			return true
		}
	case ASSIGNED:
		switch messageRequest.State {
		case COMPLETED:
			return false
		case NEW:
			return false
		case ASSIGNED:
			return true
		}
	}
	print("shouldAcceptRequest() did not return")
	return false
}

func isSubset(subset []string, superset []string) bool {
	checkset := make(map[string]bool)
	for _, element := range subset {
		checkset[element] = true
	}
	for _, value := range superset {
		if checkset[value] {
			delete(checkset, value)
		}
	}
	return len(checkset) == 0 //this implies that set is subset of superset
}

func addToAwareList(AwareList []string, id string) []string {
	for i := range AwareList {
		if AwareList[i] == id {
			return AwareList
		}
	}
	return append(AwareList, id)
}
