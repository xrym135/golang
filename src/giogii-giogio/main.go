package main

import (
	"bufio"
	"flag"
	"fmt"
	"giogii/src/check"
	"giogii/src/flashback"
	"giogii/src/lock"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
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

	flag.StringVar(&sourceUserInfo, "s", "", "")
	flag.StringVar(&sourceSocket, "si", "", "")
	flag.StringVar(&targetUserInfo, "t", "", "")
	flag.StringVar(&targetSocket, "ti", "", "")
	flag.StringVar(&parameter, "c", "", "")
	flag.StringVar(&bigTrx, "m", "", "")
	flag.StringVar(&fb, "f", "", "")
	flag.StringVar(&sshUser, "u", "", "")
	flag.StringVar(&sshPass, "p", "", "")
	flag.StringVar(&call, "C", "", "")

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

func ReadConfig() (sourceUserInfo string, targetUserInfo string, sshInfo string) {
	path := "./gii.conf"
	f, err := os.Open(path)
	if err != nil {
		log.Println("打开文件失败")
		log.Fatal(err)
		os.Exit(-1)
	}
	reader := bufio.NewReader(f)
	for i := 0; i < 3; i++ {
		readLine, _, err := reader.ReadLine()
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		value := string(readLine)
		if strings.Contains(value, "sourceUserInfo") {
			sourceUserInfo = strings.Split(value, "=")[1]
			continue
		}
		if strings.Contains(value, "targetUserInfo") {
			targetUserInfo = strings.Split(value, "=")[1]
			continue
		}
		if strings.Contains(value, "sshInfo") {
			sshInfo = strings.Split(value, "=")[1]
		}

	}
	return
}

func CallInteractive() (user string, pass string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("$ ")
		cmdString, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		cmdString = strings.TrimSuffix(cmdString, "\n")
		//cmd := exec.Command(cmdString)
		//cmd.Stderr = os.Stderr
		//cmd.Stdout = os.Stdout
		//err = cmd.Run()
		//if err != nil {
		//	fmt.Fprintln(os.Stderr, err)
		//}
		if user == "" {
			user = cmdString
		} else {
			pass = cmdString
			return
		}

	}
}
