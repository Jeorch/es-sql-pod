package handler

import (
	"encoding/json"
	uuid "github.com/satori/go.uuid"
	"net/http"
)

type HttpError struct {
	// 保持和EsSql的rest接口相同的错误格式
	ErrorObj map[string]interface{} `json:"error"`
	Status   int32                  `json:"status"`
}

func BpHttpCommonErrorHandler(err error, w http.ResponseWriter) {
	httpError := new(HttpError)
	httpError.Status = 500
	httpError.ErrorObj = make(map[string]interface{})
	httpError.ErrorObj["id"] = uuid.NewV1().String()
	httpError.ErrorObj["type"] = "EsSqlPod Handled Error."
	httpError.ErrorObj["reason"] = err.Error()

	jsValue, e := json.Marshal(httpError)
	if e != nil {
		panic(e.Error())
	}
	_, e = w.Write(jsValue)
	if e != nil {
		panic(e.Error())
	}
	return
}
