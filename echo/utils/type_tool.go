package utils

import (
	"bytes"
	"fmt"
	"github.com/json-iterator/go"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/utils/fatih/structs"
	"reflect"
	"sort"
	"strconv"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func Struct2String(src interface{}) string {
	b, err := json.Marshal(src)
	if err != nil {
		log.Error("Struct2String:%v Error:%v", src, err.Error())
		return ""
	}
	return string(b)
}

// Struct2Map 结构转map
func Struct2Map(src interface{}) map[string]interface{} {
	if reflect.TypeOf(src).Kind() == reflect.Struct {
		str := structs.New(src)
		str.TagName = "json"
		return str.Map()
	}
	switch s := src.(type) {
	case map[string]interface{}:
		return s
	}
	b, err := json.Marshal(src)
	if err != nil {
		log.Error("Struct2Map:%v Error:%v", src, err.Error())
		return make(map[string]interface{})
	}
	return Byte2Map(&b)
}

func FilterMap(src map[string]interface{}) map[string]interface{} {
	m := map[string]interface{}{}
	for k, v := range src {
		switch v := v.(type) {
		case int, int64:
			if v != 0 {
				m[k] = v
			}
		case float64:
			if int(v) != 0 {
				m[k] = v
			}
		case string:
			if v != "" {
				m[k] = v
			}
		default:
			if v != nil {
				m[k] = v
			}
		}
	}
	return m
}

//Byte2Map byte to map
func Byte2Map(data *[]byte) map[string]interface{} {
	m := map[string]interface{}{}
	_, err := ByteToObj(data, &m)
	if err != nil {
		log.Error("Byte2Map:%v Error:%v", data, err.Error())
		return make(map[string]interface{})
	}
	return m
}

//ByteToObj byte to obj
func ByteToObj(data *[]byte, s interface{}) (bool, error) {
	//以下方法大数字不会被转换为科学计数法
	d := json.NewDecoder(bytes.NewReader(*data))
	d.UseNumber()
	err := d.Decode(&s)
	//以下方法大数字会显示科学计数法
	//err := json.Unmarshal(*data, s)
	if err != nil {
		return false, err
	}
	return true, nil
}

// String2Int 字符串转int
func String2Int(v string) int {
	if v == "" {
		return 0
	}
	ret, err := strconv.Atoi(v)
	if err != nil {
		log.Error("String2Int:%v Error:%v", v, err.Error())
		return 0
	}
	return ret
}

func String2Int64(v string) int64 {
	if v == "" {
		return 0
	}
	i64, err := strconv.ParseInt(v, 10, 0)
	if err != nil {
		log.Error("String2Int64:%v Error:%v", v, err.Error())
		return 0
	}
	return i64
}

//Int2String int to string
func Int2String(v int) string {
	if v == 0 {
		return ""
	}
	return strconv.Itoa(v)
}

//Int642String int64 to string
func Int642String(v int64) string {
	if v == 0 {
		return ""
	}
	return strconv.FormatInt(v, 10)
}

//sortMap2Str 排序map并拼接成字符串
func SortMap2Str(ma map[string]string) string {
	var keys []string
	for k := range ma {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sign := ""
	for _, k := range keys {
		sign += fmt.Sprintf("%s=%s", k, ma[k])
	}
	return sign
}
