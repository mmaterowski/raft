package services

import (
	"context"

	log "github.com/sirupsen/logrus"
)

var SyncRequestService syncRequestServiceInterface = &syncRequestService{cancelHandlesForOngoingSyncRequest: make(map[string]context.CancelFunc)}

type syncRequestService struct {
	cancelHandlesForOngoingSyncRequest map[string]context.CancelFunc
}

type syncRequestServiceInterface interface {
	CancelOngoingRequests()
	AddHandler(serverId string, handle context.CancelFunc)
	DeleteHandlerFor(serverId string)
}

func (s *syncRequestService) CancelOngoingRequests() {
	if len(s.cancelHandlesForOngoingSyncRequest) > 0 {
		log.Print("Cancelling sync requests: ", len(s.cancelHandlesForOngoingSyncRequest))
		for _, cancelFunction := range s.cancelHandlesForOngoingSyncRequest {
			if cancelFunction != nil {
				cancelFunction()
			}
		}
		s.cancelHandlesForOngoingSyncRequest = make(map[string]context.CancelFunc)
	}

}

func (s *syncRequestService) AddHandler(serverId string, handle context.CancelFunc) {
	s.cancelHandlesForOngoingSyncRequest[serverId] = handle
}

func (s *syncRequestService) DeleteHandlerFor(serverId string) {
	s.cancelHandlesForOngoingSyncRequest[serverId] = nil
}
