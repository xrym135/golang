package main

import (
	"flag"
	"fmt"
	"giogii/src/check"
	"giogii/src/flashback"
	"giogii/src/lock"
	"strings"
	"testing"
)

func TestM(t *testing.T) {
	var sourceUserInfo string
	var sourceSocket string
	var targetUserInfo string
	var targetSocket string
	var parameter string
	var bigTrx string
	var fb string
	var sshUser string
	var sshPass string
	var call string

	/*flag.StringVar(&sourceUserInfo, "s", "root:drACgwoqtM", "")
	flag.StringVar(&sourceSocket, "si", "172.17.128.49:13336", "")
	flag.StringVar(&targetUserInfo, "t", "wjy_root:Wjy123456", "")
	flag.StringVar(&targetSocket, "ti", "rm-2ze5j9oqx3x70jzd94o.mysql.rds.aliyuncs.com:3306", "")
	flag.StringVar(&parameter, "c", "8c32gb", "")*/

	flag.StringVar(&sourceUserInfo, "s", "admin:!QAZ2wsx", "")
	flag.StringVar(&sourceSocket, "si", "172.17.139.37:16310", "")
	flag.StringVar(&targetUserInfo, "t", "admin:!QAZ2wsx", "")
	flag.StringVar(&targetSocket, "ti", "172.17.139.147:16310", "")
	flag.StringVar(&parameter, "c", "", "")
	flag.StringVar(&bigTrx, "m", "", "")
	flag.StringVar(&fb, "f", "", "")
	flag.StringVar(&sshUser, "u", "mysql", "")
	flag.StringVar(&sshPass, "p", "mysql", "")
	flag.StringVar(&call, "C", "C", "")

	flag.Parse()

	if strings.Trim(parameter, " ") == "c" {
		check.InitCheckParameterConf(sourceUserInfo, sourceSocket, "greatrds", targetUserInfo, targetSocket, "information_schema")
		check.DoCheckParameter(parameter)
	} else if strings.Trim(bigTrx, " ") == "m" {
		lock.InitConf(sourceUserInfo, sourceSocket, "performance_schema")
		lock.DoMonitorLock()
	} else if strings.Trim(fb, " ") == "start" {
		flashback.InitMasterConnection(sourceUserInfo, sourceSocket)
		flashback.InitSlaveConnection(targetUserInfo, targetSocket)
		flashback.DoStartFlashback(targetUserInfo, targetSocket, sshUser, sshPass)
	} else if strings.Trim(fb, " ") == "stop" {
		flashback.InitMasterConnection(sourceUserInfo, sourceSocket)
		flashback.InitSlaveConnection(targetUserInfo, targetSocket)
		flashback.DoStopFlashback(sourceUserInfo, targetUserInfo, targetSocket, sshUser, sshPass)
	} else if strings.Trim(fb, " ") == "begin" {
		sInfo, tInfo, _ := ReadConfig()
		flashback.DoBeginFlashback(sInfo, sourceSocket, tInfo, targetSocket)
	} else if strings.Trim(fb, " ") == "end" {
		sInfo, tInfo, sshInfo := ReadConfig()
		sshUser = strings.Split(sshInfo, ":")[0]
		sshPass = strings.Split(sshInfo, ":")[1]
		flashback.DoEndFlashback(sInfo, sourceSocket, tInfo, targetSocket, sshUser, sshPass)
	} else if strings.Trim(call, " ") == "C" {
		sInfo, tInfo, sshInfo := ReadConfig()
		fmt.Println(sInfo, tInfo, sshInfo)
	} else {
		check.InitCheckConsistentConf(sourceUserInfo, sourceSocket, "information_schema", targetUserInfo, targetSocket, "information_schema")
		check.DoCheck()
	}

}
