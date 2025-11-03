package entity

import "time"

type InfInnodbTrx struct {
	trxId                   *int64
	trxState                string
	trxStarted              time.Time
	trxRequestedLockId      string
	trxWaitStarted          time.Time
	trxWeight               *int64
	trxMysqlThreadId        *int64
	trxQuery                string
	trxOperationState       string
	trxTablesInUse          *int64
	trxTablesLocked         *int64
	trxLockStructs          *int64
	trxLockMemoryBytes      *int64
	trxRowsLocked           *int64
	trxRowsModified         *int64
	trxConcurrencyTickets   *int64
	trxIsolationLevel       string
	trxUniqueChecks         *int64
	trxForeignKeyChecks     *int64
	trxLastForeignKeyError  string
	trxAdaptiveHashLatched  *int64
	trxAdaptiveHashTimeout  *int64
	trxIsReadOnly           *int64
	trxAutocommitNonLocking *int64
	trxScheduleWeight       *int64
}
