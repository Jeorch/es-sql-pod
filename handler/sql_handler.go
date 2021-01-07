package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PharbersDeveloper/bp-go-lib/log"
	"github.com/PharbersDeveloper/es-sql-pods/model"
	"github.com/PharbersDeveloper/es-sql-pods/utils"
	"io/ioutil"
	"net/http"
	"os"
)

func SqlHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")                          //允许访问所有域
	w.Header().Set("Access-Control-Allow-Credentials", "true")                  //允许访问所有域
	w.Header().Set("Access-Control-Allow-Methods", "*")                         //允许访问所有域
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Access-Token") //允许访问所有域
	w.Header().Set("Access-Control-Expose-Headers", "*")                        //允许访问所有域
	w.Header().Set("Content-Type", "application/json")

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		BpHttpCommonErrorHandler(err, w)
		return
	}

	var reqMap map[string]interface{}
	err = json.Unmarshal(reqBody, &reqMap)
	if err != nil {
		BpHttpCommonErrorHandler(err, w)
		return
	}
	phLogger := log.NewLogicLoggerBuilder().Build()
	phLogger.Infof("Request Json = %v", reqMap)

	var sqlStr string
	var xValues interface{}
	if sql, ok := reqMap[utils.KeyRequestSQL]; ok {
		sqlStr = sql.(string)
	} else {
		panic("no sql found")
	}
	if x, ok := reqMap[utils.KeyRequestXValues]; ok {
		xValues = x
	}

	values := map[string]string{utils.KeyRequestSQL: sqlStr}
	jsonValue, _ := json.Marshal(values)

	esServer := os.Getenv(utils.KeyEsServer)
	if esServer == "" {
		err = errors.New("No ES_SERVER env set. ")
		BpHttpCommonErrorHandler(err, w)
		return
	}
	resp, err := http.Post(fmt.Sprint(esServer, utils.ESRouteSql),
		"application/json",
		bytes.NewBuffer(jsonValue))

	if err != nil {
		BpHttpCommonErrorHandler(err, w)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		BpHttpCommonErrorHandler(err, w)
		return
	}

	var model model.EsSQLResponse
	err = json.Unmarshal(body, &model)
	if err != nil {
		BpHttpCommonErrorHandler(err, w)
		return
	}
	if model.Hits.Hits == nil && model.Aggregations == nil {
		_, err = w.Write(body)
		if err != nil {
			BpHttpCommonErrorHandler(err, w)
		}
		return
	}

	params := make(map[string]interface{}, 0)

	for k, v := range r.URL.Query() {
		params[k] = v[0]
	}
	params[utils.KeyRequestXValues] = xValues

	source, err := model.FormatSource(params)
	if err != nil {
		BpHttpCommonErrorHandler(err, w)
		return
	}

	result, err := json.Marshal(source)
	if err != nil {
		BpHttpCommonErrorHandler(err, w)
		return
	}

	_, err = w.Write(result)
	if err != nil {
		BpHttpCommonErrorHandler(err, w)
	}
}
