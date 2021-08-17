package consts

import "time"

const (
	Integrationtest      = "IntegrationTest"
	Prod                 = "Prod"
	Local                = "Local"
	LocalWithPersistence = "LocalWithPersistence"
)

const (
	TermInitialValue         = 1
	TermUninitializedValue   = 0
	NoPreviousEntryValue     = 0
	LeaderCommitInitialValue = 0
	FirstEntryIndex          = 1
	HeartbeatInterval        = 3 * time.Second
)

const (
	LocalSignalRPort  = KimSignalRPort
	LocalApiPort      = KimPort
	RickyPort         = 6971
	RickySignalRPort  = 6981
	LaszloPort        = 6970
	LaszloSignalRPort = 6980
	KimPort           = 6969
	KimSignalRPort    = 6979
	RickyServiceName  = "ricky"
	KimServiceName    = "kim"
	LaszloServiceName = "laszlo"
	KimId             = "Kim"
	LaszloId          = "Laszlo"
	RickyId           = "Ricky"
)
