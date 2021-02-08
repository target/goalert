package lock

// Defined global lock values.
const (
	GlobalMigrate            = uint32(0x1337) // 4919
	GlobalEngineProcessing   = uint32(0x1234) // 4660
	GlobalMessageSending     = uint32(0x1330) // 4912
	RegionalEngineProcessing = uint32(0x1342) // 4930
	ModularEngineProcessing  = uint32(0x1347) // 4935
	GlobalSwitchOver         = uint32(0x1111) // 4369
)
