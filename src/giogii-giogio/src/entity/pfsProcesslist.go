package entity

type Processlist struct {
	ID      *int64
	USER    string
	HOST    string
	DB      string
	COMMAND string
	TIME    *int64
	STATE   string
	INFO    string
}
