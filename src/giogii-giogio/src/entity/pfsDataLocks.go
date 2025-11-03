package entity

type DataLocks struct {
	Engine              string
	EngineLockId        string
	EngineTransactionId *int64
	ThreadId            *int64
	EventId             *int64
	ObjectSchema        string
	ObjectName          string
	PartitionName       string
	SubpartitionName    string
	IndexName           string
	ObjectInstanceBegin *int64
	LockType            string
	LockMode            string
	LockStatus          string
	LockData            string
}
