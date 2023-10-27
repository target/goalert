package lock

// Defined global lock values.
const (
	// Ensures only a single instance is performing migrations at a time.
	GlobalMigrate = uint32(0x1337) // 4919

	// Currently unused.
	GlobalEngineProcessing = uint32(0x1234) // 4660

	// Ensures only a single instance is sending messages,
	// this includes out-of-transaction processes.
	GlobalMessageSending = uint32(0x1330) // 4912

	// Currently unused.
	RegionalEngineProcessing = uint32(0x1342) // 4930

	// Currently unused.
	ModularEngineProcessing = uint32(0x1347) // 4935

	// A shared lock is grabbed by the application, and exclusive
	// lock during the final sync as a stop-the-world lock for the
	// atomic DB switch.
	GlobalSwitchOver = uint32(0x1111) // 4369

	// Used exclusively by engine instances to elect a leader.
	//
	// Only the instance and connection with this lock is allowed
	// to perform trigger updates and synchronization.
	//
	// It must be acquired before the global switchover lock.
	GlobalSwitchOverExec = uint32(0x1112) // 4370
)
