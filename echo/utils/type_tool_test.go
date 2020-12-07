package utils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestInt2String(t *testing.T) {
	ret := Int2String(100)
	assert.Equal(t, "100", ret)
}

func TestInt642String(t *testing.T) {
	var i int64
	i = 100
	ret := Int642String(i)
	assert.Equal(t, "100", ret)
}

func TestString2Int(t *testing.T) {
	s := "100"
	i := String2Int(s)
	assert.Equal(t, 100, i)
}

type OB struct {
	Name   string                 `json:"name"`
	Gender int                    `json:"gender"`
	Field  map[string]interface{} `json:"field"`
	birth  int64
}

func TestStruct2Map(t *testing.T) {

	ts := OB{
		Name:   "1234",
		Gender: 100000000000,
		birth:  NowTimeStamp(),
		Field: map[string]interface{}{
			"appId": 123,
		},
	}
	ret := Struct2Map(&ts)
	fmt.Println(ret)
	assert.Empty(t, ret["birth"], "转换结构: %v", ret)

	now := time.Now()
	for i := 0; i < 100000; i++ {
		_ = Struct2Map(&ts)
	}
	fmt.Println(time.Since(now))

}

func TestFilterMap(t *testing.T) {

	ts := OB{
		//Name: "1234",
		birth: NowTimeStamp(),
		Field: map[string]interface{}{
			"appId": 123,
		},
	}
	data := Struct2Map(&ts)
	t.Logf("%+v, %T", data, data["gender"])
	ret := FilterMap(data)
	t.Logf("%+v", ret)
	//assert.Empty(t, ret["birth"], "转换结构: %v", ret)
}

func TestRegexp(t *testing.T) {
	ss := "{namespace}:game:test:{app_id}:{app_key}"
	reg := regexp.MustCompile(`\{(\w+)\}`)
	ret := reg.FindAllString(ss, -1)
	data := map[string]interface{}{
		"namespace": "test",
		"app_id":    "id",
		"app_key":   "key",
	}
	regg := regexp.MustCompile(`[\{\}]`)
	for _, s := range ret {
		tmp := regg.ReplaceAllString(s, "")
		t.Logf("%v", tmp)
		ss = strings.Replace(ss, s, data[tmp].(string), 1)
	}
	t.Logf("%+v", ss)
}
