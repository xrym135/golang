package check

import (
	"fmt"
	"giogii/src/mapper"
	"log"
	"strconv"
	"strings"
)

var strSql string
var MasterSqlScaleOperator mapper.SqlScaleOperator
var SlaveSqlScaleOperator mapper.SqlScaleOperator

func InitCheckConsistentConf(sourceUserInfo string, sourceSocket string, sourceDatabase string, targetUserInfo string, targetSocket string, targetDatabase string) {
	s, t := mapper.InitAllConn(sourceUserInfo, sourceSocket, sourceDatabase, targetUserInfo, targetSocket, targetDatabase)
	MasterSqlScaleOperator = &s
	SlaveSqlScaleOperator = &t
}

func DoCheck() {

	/**
	是否需要判断是否是主集群？
	是否需要判断是发是备集群？
	*/
	var rs = 2

	strSql = fmt.Sprint("show master status")
	masterStatus := MasterSqlScaleOperator.DoQueryParseMaster(strSql)

	strSql = fmt.Sprint("show variables like 'server_uuid'")
	masterUuid := MasterSqlScaleOperator.DoQueryParseString(strSql)

	strSql = fmt.Sprint("show slave status")
	slaveStatus := SlaveSqlScaleOperator.DoQueryParseSlave(strSql)

	var masterGtid string
	var slaveGtid string

	// 如果两个语句的返回值里任何一个不包含binlog文件，直接返回2
	if masterStatus.File == "" || slaveStatus.MasterLogFile == "" {
		log.Printf("show master status / show slave status return null")
		fmt.Println(rs)
		return
	}

	// 如果slave读取的binlog文件和主库当前binlog文件不相等，说明延迟很大，直接返回2，不需要进行下面比较
	if masterStatus.File != slaveStatus.MasterLogFile {
		log.Printf("备集群此刻读取主集群的binlog文件和主集群生产的binlog文件不相等，")
		//fmt.Println(rs)
		//return
	}

	// 获取主集群执行的gtid
	masterExecutedGtids := strings.Split(masterStatus.ExecutedGtidSet, ",")
	for i := 0; i < len(masterExecutedGtids); i++ {
		if strings.Contains(masterExecutedGtids[i], masterUuid) {
			masterGtid = strings.Trim(masterExecutedGtids[i], "\n")
			break
		}
	}

	// 获取备集群执行的gtid
	slaveExecutedGtids := strings.Split(slaveStatus.ExecutedGtidSet, ",")
	for i := 0; i < len(slaveExecutedGtids); i++ {
		if strings.Contains(slaveExecutedGtids[i], masterUuid) {
			slaveGtid = strings.Trim(slaveExecutedGtids[i], "\n")
			break
		}
	}

	// 这里的逻辑是判断主集群GTID是否和备集群GTID相等
	if strings.Contains(slaveGtid, masterGtid) {
		rs -= 1
	} else if strings.Contains(slaveGtid, "-") && strings.Contains(masterGtid, "-") {
		slaveLastIndex := strings.LastIndex(slaveGtid, "-")
		masterLastIndex := strings.LastIndex(masterGtid, "-")
		if masterGtid[masterLastIndex+1:] == slaveGtid[slaveLastIndex+1:] {
			rs -= 1
		}
	}

	// 这里的逻辑是判断主集群binlog点位是否和备集群点位相等
	if strconv.Itoa(*masterStatus.Position) == strconv.Itoa(*slaveStatus.ReadMasterLogPos) {
		rs -= 1
	}

	log.Printf("Source Cluster GTID：%s", masterGtid)
	log.Printf("Target Cluster GTID：%s", slaveGtid)
	log.Print("Source Cluster POS: ", *masterStatus.Position)
	log.Print("Target Cluster POS: ", *slaveStatus.ReadMasterLogPos)

	fmt.Println(rs)

	defer func() {
		MasterSqlScaleOperator.DoClose()
		SlaveSqlScaleOperator.DoClose()
	}()

}
