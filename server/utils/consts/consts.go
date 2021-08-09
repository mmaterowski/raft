package consts

import "time"

const (
	Integrationtest      = "IntegrationTest"
	Prod                 = "Prod"
	Local                = "Local"
	LocalWithPersistence = "LocalWithPersistence"
)

var TermInitialValue = 1
var TermUninitializedValue = 0
var NoPreviousEntryValue = 0
var LeaderCommitInitialValue = 0
var FirstEntryIndex = 1

var HeartbeatInterval = 3 * time.Second
var RickyPort = 6970
var KimPort = 6969
var RickyServiceName = "ricky"
var KimServiceName = "kim"
var LaszloServiceName = "laszlo"
var KimId = "Kim"
var LaszloId = "Laszlo"
var RickyId = "Ricky"
