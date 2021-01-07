package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestEsSQLResponse(t *testing.T) {

	esServer  := "http://192.168.100.174:9200"
	reqUrl := fmt.Sprint(esServer, "/_sql")
	sql := "select 月份, 药品名, count(药品名.keyword) as 数量  from hx2 group by 月份.keyword, 药品名.keyword order by 月份.keyword"
	params := map[string]interface{}{
		"x-axis": "月份",
		"y-axis": "数量",
		"dimensionKeys": "药品名",
	}
	fmt.Println(params)

	values := map[string]string{"sql": sql}
	jsonValue, _ := json.Marshal(values)

	resp, err := http.Post(reqUrl,
		"application/json",
		bytes.NewBuffer(jsonValue))

	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	var model EsSQLResponse
	err = json.Unmarshal(body, &model)
	if err != nil {
		fmt.Println(err)
		return
	}

	source, err := model.FormatSource(params)

	if err == nil {
		fmt.Println(source)
	}

}
