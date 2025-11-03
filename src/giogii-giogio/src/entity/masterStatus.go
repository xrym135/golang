package entity

type MasterStatus struct {
	File            string
	Position        *int
	BinlogDoDB      string
	BinlogIgnoreDB  string
	ExecutedGtidSet string
}
