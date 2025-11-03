package entity

type Threads struct {
	ThreadId           *int64
	NAME               string
	TYPE               string
	ProcesslistId      *int64
	ProcesslistUser    string
	ProcesslistHost    string
	ProcesslistDb      string
	ProcesslistCommand string
	ProcesslistTime    *int64
	ProcesslistState   string
	ProcesslistInfo    string
	ParentThreadId     *int64
	ROLE               string
	INSTRUMENTED       string
	HISTORY            string
	ConnectionType     string
	ThreadOsId         *int64
	ResourceGroup      string
}
