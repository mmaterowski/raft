package main

type RequestVoteRequest struct {
}

func RequestVoteRPC(term int, candidateId string, lastLogIndex int, lastLogTerm int) (int, bool) {
	currentTerm := 2
	voteGranted := true
	return currentTerm, voteGranted
}

// func AppendEntriesRPC(request AppendEntriesRequest, sendTo string) (int, bool) {
// 	currentTerm := 5
// 	success := true
// 	j, err := json.Marshal(request.Entries)
// 	fmt.Println(string(j), err)
// 	return currentTerm, success
// }
