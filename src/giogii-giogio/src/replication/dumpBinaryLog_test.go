package replication

import (
	"context"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
	"os"
	"testing"
	"time"
)

func TestDumpBinlog(t *testing.T) {
	// Create a binlog syncer with a unique server id, the server id must be different from other MySQL's.
	// flavor is mysql or mariadb
	cfg := replication.BinlogSyncerConfig{
		ServerID: 100,
		Flavor:   "mysql",
		Host:     "172.17.128.152",
		Port:     16315,
		User:     "admin",
		Password: "!QAZ2wsx",
	}
	syncer := replication.NewBinlogSyncer(cfg)

	// Start sync with specified binlog file and position
	streamer, _ := syncer.StartSync(mysql.Position{"greatdb-bin.000001", 96034464})

	// or you can start a gtid replication like
	// streamer, _ := syncer.StartSyncGTID(gtidSet)
	// the mysql GTID set likes this "de278ad0-2106-11e4-9f8e-6edd0ca20947:1-2"
	// the mariadb GTID set likes this "0-1-100"

	for {
		ev, _ := streamer.GetEvent(context.Background())
		// Dump event
		ev.Dump(os.Stdout)
	}

	// or we can use a timeout context
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		ev, err := streamer.GetEvent(ctx)
		cancel()

		if err == context.DeadlineExceeded {
			// meet timeout
			continue
		}

		ev.Dump(os.Stdout)
	}

}
