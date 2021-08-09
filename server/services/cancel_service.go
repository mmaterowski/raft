package services

import (
	"context"

	log "github.com/sirupsen/logrus"
)

type SyncRequestService struct {
	cancelHandlesForOngoingSyncRequest map[string]context.CancelFunc
}

func NewSyncRequestService() *SyncRequestService {
	return &SyncRequestService{cancelHandlesForOngoingSyncRequest: make(map[string]context.CancelFunc)}
}
func (s *SyncRequestService) CancelOngoingRequests() {
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

func (s *SyncRequestService) AddHandler(serverId string, handle context.CancelFunc) {
	s.cancelHandlesForOngoingSyncRequest[serverId] = handle
}

func (s *SyncRequestService) DeleteHandlerFor(serverId string) {
	s.cancelHandlesForOngoingSyncRequest[serverId] = nil
}
