package utils

import (
	"fmt"
	"github.com/liangdas/mqant/log"
	logs "github.com/liangdas/mqant/log/beego"
	"testing"
)

//基准测试的代码文件必须以_test.go结尾
//基准测试的函数必须以Benchmark开头，必须是可导出的
//基准测试函数必须接受一个指向Benchmark类型的指针作为唯一参数
//基准测试函数不能有返回值
//b.ResetTimer是重置计时器，这样可以避免for循环之前的初始化代码的干扰
//最后的for循环很重要，被测试的代码要放到循环里
//b.N是基准测试框架提供的，表示循环的次数，因为需要反复调用测试的代码，才可以评估性能
//https://msd.misuland.com/pd/3127746505234973848
/*
BenchmarkLogger_JErrorWithOMap-4     	   39554	     27410 ns/op
BenchmarkLogger_JErrorWithMap-4      	   37336	     30145 ns/op
BenchmarkLogger_JErrorWithStruct-4   	   42775	     25336 ns/op
BenchmarkLogger_Errorf-4             	   60502	     19507 ns/op
BenchmarkLog-4                       	   53410	     22267 ns/op
PASS
*/
func init() {
	mp := make(map[string]interface{})
	mp2 := make(map[string]interface{})

	mp["contenttype"] = "application/json"
	mp2["perm"] = "0600"
	mp2["prefix"] = "prof."
	mp2["daily"] = true
	mp2["rotate"] = true
	mp2["maxsize"] = 0
	mp2["maxLines"] = 0
	mp["file"] = mp2
	log.InitLog(false, "development", "", mp)
	log.LogBeego().SetFormatFunc(logs.DefineErrorLogFunc("development", 4))
}

type AuthenticationData struct {
	Namespace   string `json:"namespace" query:"namespace" form:"namespace" valid:"required"`
	ClientId    string `json:"client_id" query:"client_id" form:"client_id" valid:"optional"`
	UserId      string `json:"user_id" query:"user_id" form:"user_id" valid:"required"`
	Duid        string `json:"duid" query:"duid" form:"duid" valid:"optional"`
	CitizenId   string `json:"citizen_id" query:"citizen_id" form:"citizen_id" valid:"required"`
	CitizenName string `json:"citizen_name" query:"citizen_name" form:"citizen_name" valid:"required"`
	Token       string `json:"token" query:"token" form:"token" valid:"optional"`
	AuthorCode  string `json:"authorCode" query:"authorCode" form:"authorCode" valid:"optional"`
}

type UpdateAuthenticationData struct {
	AuthenticationData
	OldCitizenId   string `json:"old_citizen_id" query:"old_citizen_id" form:"old_citizen_id" valid:"required"`
	OldCitizenName string `json:"old_citizen_name" query:"old_citizen_name" form:"old_citizen_name" valid:"required"`
}

//渠道登录返回的实名认证信息
type ThirdAuthenticationData struct {
	ClientId    string `json:"client_id" query:"client_id" form:"client_id" valid:"required"`
	UserId      string `json:"user_id" query:"user_id" form:"user_id" valid:"required"`
	Duid        string `json:"duid" query:"duid" form:"duid" valid:"optional"`
	SnsId       string `json:"sns_id" query:"sns_id" form:"sns_id" valid:"optional"` //渠道标识
	CitizenId   string `json:"citizen_id" query:"citizen_id" form:"citizen_id" valid:"optional"`
	CitizenName string `json:"citizen_name" query:"citizen_name" form:"citizen_name" valid:"optional"`
	Birthday    string `json:"birthday" query:"birthday" form:"birthday" valid:"optional"`     //生日 YYYYMMdd 格式
	Age         string `json:"age" query:"age" form:"age" valid:"optional"`                    //年龄
	AgeRange    string `json:"age_range"  query:"age_range" form:"age_range" valid:"optional"` //年龄段 1：age<8  2：8<=age<16 3：16<=age<18 4：age>=18
	Ignore      string `json:"ignore" query:"ignore" form:"ignore" valid:"optional"`           //是否由渠道限制，默认 0。0：由我们限制 1：由渠道限制
}

type QueryUserInfoResponse struct {
	ChannelVerified bool   `json:"channel_verified"` //是否又渠道进行了实名认证
	Verified        bool   `json:"verified"`         //是否已实名认证
	Ignore          bool   `json:"ignore"`           //登录支付限制是否由渠道控制
	CitizenId       string `json:"citizen_id"`       //身份证号
	IsAdult         bool   `json:"is_adult"`         //是否已成年
	//Age               int64  `json:"age"`                 //年龄
	//Birthday          int64  `json:"birthday"`            //生日
	//Sex               string `json:"sex"`                 //性别 male 男 female 女
	IsHoliday         bool   `json:"is_holiday"`          //是否为节假日
	IsLimited         bool   `json:"is_limited"`          //是否被限制登录
	LimitTime         int64  `json:"limit_time"`          //在线时长限制
	RemainingTime     int64  `json:"remaining_time"`      //剩余在线时长
	OnlineTime        int64  `json:"online_time"`         //在线时长
	PayLimit          int64  `json:"pay_limit"`           //单笔支付限额
	MonthPayLimit     int64  `json:"month_pay_limit"`     //每月支付限额
	DayAmount         int64  `json:"day_amount"`          //当日累计充值金额 (单位：分)
	MonthAmount       int64  `json:"month_amount"`        //当月累计充值金额 (单位：分)
	AllowCharge       bool   `json:"allow_charge"`        //是否允许充值
	RemainingCharge   int64  `json:"remaining_charge"`    //当月剩余充值金额
	ExpiresIn         int64  `json:"expires_in"`          //缓存过期时间(秒)，0：不缓存
	LoginLimitedCode  int64  `json:"login_limited_code"`  //限制登录原因
	LoginLimitedMsg   string `json:"login_limited_msg"`   //限制登录原因描述
	ChargeLimitedCode int64  `json:"charge_limited_code"` //限制支付原因
	ChargeLimitedMsg  string `json:"charge_limited_msg"`  //限制支付原因描述
}

//测试新格式日志的打印消耗
func BenchmarkLogger_JErrorWithOMap(b *testing.B) {
	logger := GetRootLogger()
	testStruct := struct {
		TestInt    int
		TestString string
		TestBool   bool
	}{
		1,
		"2",
		true,
	}
	//b.ResetTimer是重置计时器，这样可以避免for循环之前的初始化代码的干扰
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.JError(LogMsg{EventId: "test", SubEventId: "100", ErrorMsg: "test", ErrorReport: "test error",
			Rparam: Struct2Map(testStruct), EventParams: map[string]interface{}{"test": "test"}})
	}
}

func BenchmarkLogger_JErrorWithMap(b *testing.B) {
	logger := GetRootLogger()
	testStruct := struct {
		TestInt    int
		TestString string
		TestBool   bool
	}{
		1,
		"2",
		true,
	}
	//b.ResetTimer是重置计时器，这样可以避免for循环之前的初始化代码的干扰
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.JError(LogMsg{EventId: "test", SubEventId: "100", ErrorMsg: "test", ErrorReport: "test error",
			Rparam: Struct2Map(testStruct), EventParams: Struct2Map(testStruct)})
	}
}

func BenchmarkLogger_JErrorWithStruct(b *testing.B) {
	logger := GetRootLogger()
	testStruct := struct {
		TestInt    int
		TestString string
		TestBool   bool
	}{
		1,
		"2",
		true,
	}
	//b.ResetTimer是重置计时器，这样可以避免for循环之前的初始化代码的干扰
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.JError(LogMsg{EventId: "test", SubEventId: "100", ErrorMsg: "test", ErrorReport: "test error",
			Rparam: Struct2Map(testStruct), EventParams: testStruct})
	}
}

func BenchmarkLogger_Errorf(b *testing.B) {
	testStruct := struct {
		TestInt    int
		TestString string
		TestBool   bool
	}{
		1,
		"2",
		true,
	}
	for i := 0; i < b.N; i++ {
		log.Error("@test", "test_code_100", "Query DB error", fmt.Errorf("not found").Error(), map[string]interface{}{"test": 1}, testStruct)
	}
}

func BenchmarkLog(b *testing.B) {

	logger := GetRootLogger()
	req := UpdateAuthenticationData{
		AuthenticationData: AuthenticationData{
			Namespace:   "namespace",
			ClientId:    "clientId",
			UserId:      "10010",
			Duid:        "213dfgrbrtgewrfqw",
			CitizenId:   "CitizenId",
			CitizenName: "CitizenName",
			Token:       "Token",
			AuthorCode:  "AuthorCode",
		},
		OldCitizenId:   "OldCitizenId",
		OldCitizenName: "OldCitizenName",
	}
	data := QueryUserInfoResponse{
		ChannelVerified:   false,
		Verified:          false,
		Ignore:            false,
		CitizenId:         "efeeytjrhewgrehtjty",
		IsAdult:           false,
		IsHoliday:         false,
		IsLimited:         false,
		LimitTime:         3600,
		RemainingTime:     235,
		OnlineTime:        34654,
		PayLimit:          30000,
		MonthPayLimit:     50000,
		DayAmount:         45734,
		MonthAmount:       12354,
		AllowCharge:       false,
		RemainingCharge:   562924,
		ExpiresIn:         3600,
		LoginLimitedCode:  0,
		LoginLimitedMsg:   "",
		ChargeLimitedCode: 0,
		ChargeLimitedMsg:  "",
	}

	b.StartTimer() //重新开始时间
	for i := 0; i < b.N; i++ {
		logger.JError(LogMsg{
			EventId:     "test",
			SubEventId:  "100",
			ErrorMsg:    "fgfuergfyurfyrvfygewufhfygtrhw",
			ErrorReport: "hfiuergyrfygvhdbcfrtgfquwbsbfe",
			EventParams: data,
			Rparam:      req,
		})
		//logger.Errorf("sfegeg %+v %+v", data, req)
	}
	b.StopTimer()
}
