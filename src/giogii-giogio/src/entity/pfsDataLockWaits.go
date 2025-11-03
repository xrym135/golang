package entity

type DataLockWaits struct {
	Engine                        string
	RequestingEngineLockId        string
	RequestingEngineTransactionId *int64
	RequestingThreadId            *int64
	RequestingEventId             *int64
	RequestingObjectInstanceBegin *int64
	BlockingEngineLockId          string
	BlockingEngineTransactionId   *int64
	BlockingThreadId              *int64
	BlockingEventId               *int64
	BlockingObjectInstanceBegin   *int64
}
