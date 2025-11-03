package entity

import "database/sql"

type DataServers struct {
	Servername         sql.NullString
	Host               sql.NullString
	Port               sql.NullString
	Username           sql.NullString
	Status             sql.NullString
	MasterOnlineStatus sql.NullString
	MasterBackup       sql.NullString
	RemoteUser         sql.NullString
	RemotePort         sql.NullString
	MaxNeededConn      sql.NullString
	MasterPriority     sql.NullString
}
