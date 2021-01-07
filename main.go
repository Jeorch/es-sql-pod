package main

import (
	"fmt"
	"github.com/PharbersDeveloper/bp-go-lib/env"
	"github.com/PharbersDeveloper/bp-go-lib/log"
	"github.com/PharbersDeveloper/es-sql-pods/handler"
	"github.com/PharbersDeveloper/es-sql-pods/utils"
	"net/http"
	"os"
)

func main() {

	//本地测试用，部署时请注释掉setLocalEnv()
	//setLocalEnv()

	mux := http.NewServeMux()

	mux.HandleFunc(utils.RouteSql, handler.SqlHandler)

	phLogger := log.NewLogicLoggerBuilder().Build()
	phLogger.Infof("Listening port=%s", utils.Port)
	http.ListenAndServe(fmt.Sprint(":", utils.Port), mux)
}

func setLocalEnv() {
	//项目范围内的环境变量
	_ = os.Setenv(env.ProjectName, "es-sql-pods")

	//log
	_ = os.Setenv(env.LogTimeFormat, "2006-01-02 15:04:05")
	_ = os.Setenv(env.LogOutput, "console")
	//_ = os.Setenv(env.LogOutput, "./tmp/es-sql-pods.log")
	_ = os.Setenv(env.LogLevel, "info")

	//es
	_ = os.Setenv(utils.KeyEsServer, "http://59.110.31.215:9200")

}

