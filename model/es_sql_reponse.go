package model

import (
	"fmt"
	"github.com/PharbersDeveloper/es-sql-pods/utils"
	"strings"
)

type EsSQLResponse struct {
	Took 			int64		`json:"took"`
	TimeOut 		bool		`json:"time_out"`
	Hits			HitsDetail	`json:"hits"`
	Aggregations	interface{}	`json:"aggregations"`
}

type HitsDetail struct {
	Total		interface{}	`json:"total"`
	MaxScore	float64		`json:"max_score"`
	Hits		[]Hit
}

type Hit struct {
	Index 	string		`json:"_index"`
	Type 	string		`json:"_type"`
	Id		string		`json:"_id"`
	Score	float64		`json:"_score"`
	Source 	interface{}	`json:"_source"`
}

func (esr EsSQLResponse) FormatSource(params interface{}) (interface{}, error) {

	listMap := make([]map[string]interface{}, 0)
	if esr.Aggregations == nil {
		listMap = getHitsRec(esr.Hits.Hits)
	} else {
		// 将Agg嵌套结构平滑成一层list map
		listMap = getAggRec(esr.Aggregations.(map[string]interface{}))
	}

	return listMap2DimensionalArray(listMap, params)

}

func getHitsRec(hits []Hit) (result []map[string]interface{}) {
	for _, hit := range hits {
		result = append(result, hit.Source.(map[string]interface{}))
	}
	return
}

func getAggRec(data map[string]interface{}) (result []map[string]interface{}) {
	lastMap := make(map[string]interface{})

	for aggKey, aggValue := range data {
		formatKet := strings.ReplaceAll(aggKey, ".keyword", "")
		valueMap := aggValue.(map[string]interface{})
		if buckets, ok := valueMap["buckets"]; ok {
			for _, item := range buckets.([]interface{}) {
				bucket := item.(map[string]interface{})
				key := bucket["key"]
				delete(bucket, "key")
				delete(bucket, "doc_count")
				if len(bucket) > 0 {
					for _, sub := range getAggRec(bucket) {
						sub[formatKet] = key
						result = append(result, sub)
					}
				} else {
					result = append(result, map[string]interface{}{formatKet:key})
				}
			}
		} else {
			lastMap[formatKet] = valueMap["value"]
		}
	}
	if len(lastMap) != 0 {
		result = append(result, lastMap)
	}
	return result
}

func listMap2DimensionalArray(listMap []map[string]interface{}, params interface{}) (interface{}, error) {

	if len(listMap) < 1 {
		return nil, nil
	}

	var tag string
	var paramsMap map[string]interface{}
	if m, ok := params.(map[string]interface{}); ok {
		paramsMap = m
	}
	if key, ok := paramsMap[utils.KeyParamTag].(string); ok {
		tag = key
	}

	switch tag {
	case "array":
		return dealArrayTag(listMap, paramsMap)
	case "chart":
		return dealChartTag(listMap, paramsMap)
	case "row2line":
		return dealRow2LineTag(listMap, paramsMap)
	case "listMap":
		return listMap, nil
	default:
		return nil, fmt.Errorf("No implementation for tag=%s. ", tag)
	}

}

func dealRow2LineTag(listMap []map[string]interface{}, paramsMap map[string]interface{}) (interface{}, error) {

	result := make([][]interface{}, len(listMap) + 1)

	var dimensionKeys []string
	if key, ok := paramsMap[utils.KeyParamDimensionKeys].(string); ok {
		dimensionKeys = strings.Split(key, ",")
	}
	existDimensions := make([]bool, len(dimensionKeys))
	//check dimensionKeys in response data
	for k, _ := range listMap[0] {
		for i, d := range dimensionKeys {
			if k == d {
				existDimensions[i] = true
			}
		}
	}
	if len(dimensionKeys) > 0 {
		for i, exist := range existDimensions {
			if !exist {
				return nil, fmt.Errorf("Error! No dimension = %s found in response data. ", dimensionKeys[i])
			}
		}
	}

	for _, k := range dimensionKeys  {
		result[0] = append(result[0], k)
	}

	for i, m := range listMap {
		for _, k := range dimensionKeys  {
			if str, ok := m[k].(string); ok {
				////TODO:临时加一行处理"省"，之后要求聚合后的数据要满足echarts地图省份名字要求
				//tempValue := strings.ReplaceAll(str, "省", "")
				//result[i+1] = append(result[i+1], tempValue)
				result[i+1] = append(result[i+1], str)
			} else {
				result[i+1] = append(result[i+1], m[k])
			}

		}
	}

	return result, nil
}

func dealArrayTag(listMap []map[string]interface{}, paramsMap map[string]interface{}) (interface{}, error) {

	result := make([]interface{}, 0)
	for _, m := range listMap {
		for _, v := range m  {
			result = append(result, v)
		}
	}

	return result, nil
}

func dealChartTag(listMap []map[string]interface{}, paramsMap map[string]interface{}) (interface{}, error) {
	var xAxisKey string
	var yAxisKey string
	var dimensionKeys []string
	xAxisValues := make([]interface{}, 0)
	if key, ok := paramsMap[utils.KeyParamXAxis].(string); ok {
		xAxisKey = key
	}
	if values, ok := paramsMap[utils.KeyRequestXValues].([]interface{}); ok {
		for _, v := range values {
			xAxisValues = append(xAxisValues, v)
		}
	}
	if key, ok := paramsMap[utils.KeyParamYAxis].(string); ok {
		yAxisKey = key
	}
	if key, ok := paramsMap[utils.KeyParamDimensionKeys].(string); ok {
		dimensionKeys = strings.Split(key, ",")
	}

	existX := false
	existY := false
	existDimensions := make([]bool, len(dimensionKeys))
	dimensionsValues := make([]interface{}, 0)

	//check keys in response data
	for k, _ := range listMap[0] {
		if k == xAxisKey {
			existX = true
		}
		if k == yAxisKey {
			existY = true
		}
		for i, d := range dimensionKeys {
			if k == d {
				existDimensions[i] = true
			}
		}
	}
	if !existX {
		return nil, fmt.Errorf("Error! No x-axis = %s found in response data. ", xAxisKey)
	}
	if !existY {
		return nil, fmt.Errorf("Error! No x-axis = %s found in response data. ", yAxisKey)
	}
	if len(dimensionKeys) > 0 {
		for i, exist := range existDimensions {
			if !exist {
				return nil, fmt.Errorf("Error! No dimension = %s found in response data. ", dimensionKeys[i])
			}
		}
	}

	// use value-key map removing duplicate xAxis array
	// AxisValuesMap is used for get one xAxis-value's yAxis-value
	tempAxisValuesMap := make(map[interface{}]interface{}, 0)
	// dimensionValuesMapAxisValuesMap is used for get one dimension'AxisValuesMap
	dimensionValuesMapAxisValuesMap := make(map[interface{}]map[interface{}]interface{}, 0)

	for _, m := range listMap {
		tempAxisValuesMap[m[xAxisKey]] = m[yAxisKey]
		if len(dimensionKeys) > 0 {
			oneDimensionArr := make([]string, 0)
			for _, d := range dimensionKeys {
				oneDimensionArr = append(oneDimensionArr, m[d].(string))
			}
			oneDimension := strings.Join(oneDimensionArr, ",")
			if _, ok := dimensionValuesMapAxisValuesMap[oneDimension]; ok {
				dimensionValuesMapAxisValuesMap[oneDimension][m[xAxisKey]] = m[yAxisKey]
			} else {
				dimensionValuesMapAxisValuesMap[oneDimension] = map[interface{}]interface{}{
					m[xAxisKey]: m[yAxisKey],
				}
			}
		}
	}
	if len(xAxisValues) == 0 {
		for v, _ :=  range tempAxisValuesMap {
			xAxisValues = append(xAxisValues, v.(string))
		}
	}

	for v, _ :=  range dimensionValuesMapAxisValuesMap {
		dimensionsValues = append(dimensionsValues, v)
	}

	result := make([][]interface{}, len(dimensionsValues) + 1)
	result[0] = []interface{}{xAxisKey}
	for _, v := range xAxisValues {
		if x, ok := v.(string); ok {
			////临时加一行处理"省"，之后要求聚合后的数据要满足echarts地图省份名字要求
			//tempValue := strings.ReplaceAll(x, "省", "")
			//result[0] = append(result[0], tempValue)
			result[0] = append(result[0], x)
		} else {
			result[0] = append(result[0], v)
		}
	}

	for i, dimension := range dimensionsValues {

		result[i+1] = []interface{}{dimension}
		////临时加一行处理"省"，之后要求聚合后的数据要满足echarts地图省份名字要求
		//tempValue := strings.ReplaceAll(dimension.(string), "省", "")
		//result[i+1] = []interface{}{tempValue}

		for _, x := range xAxisValues {
			result[i+1] = append(result[i+1], dimensionValuesMapAxisValuesMap[dimension][x])
		}
	}

	return result, nil
}
