package request_control

import (
	. "project/types"
)

func shouldAcceptMessage(local_request Request_t, message_request Request_t) bool {
	if message_request.State == UNKNOWN {
		return false
	}
	if local_request.State == UNKNOWN {
		return true
	}
	if message_request.Count < local_request.Count {
		return false
	}
	if message_request.Count > local_request.Count {
		return true
	}
	if message_request.State == local_request.State && is_subset(message_request.AwareList, local_request.AwareList) {
		// count is equal
		return false
	}

	switch local_request.State {
	case COMPLETED:
		switch message_request.State {
		case COMPLETED:
			return true
		case NEW:
			return true
		case ASSIGNED:
			println("FROM COMPLETED TO ASSIGNED (should not happen)")
			return true
		}
	case NEW:
		switch message_request.State {
		case COMPLETED:
			return false
		case NEW:
			return true
		case ASSIGNED:
			return true
		}
	case ASSIGNED:
		switch message_request.State {
		case COMPLETED:
			return false
		case NEW:
			return false
		case ASSIGNED:
			return true
		}
	}
	print("shouldAcceptMessage() did not return")
	return false
}

func is_subset(subset []string, superset []string) bool {
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