package entity

type MetadataLocks struct {
	ObjectType       string
	LockType         string
	LockStatus       string
	ProcesslistId    *int64
	ProcesslistState string
	ProcesslistInfo  string
	ProcesslistTime  string
}
