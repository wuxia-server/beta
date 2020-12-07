package tools

import (
	"container/list"
	"fmt"
	"github.com/liangdas/mqant/log"
	beegolog "github.com/liangdas/mqant/log/beego"
	"github.com/liangdas/mqant/module"
	"os"
	"beta/echo/utils"
)

var (
	prof      *beegolog.BeeLogger
	observers *list.List
)

func InitProfReport(app module.App, PlatformName, PlatformDomain string) error {
	platformDomain = PlatformDomain
	platformName = PlatformName
	defaultProfPath := fmt.Sprintf("%s/prof", app.WorkDir())
	_, err := os.Open(defaultProfPath)
	if err != nil {
		//文件不存在
		err := os.Mkdir(defaultProfPath, os.ModePerm) //
		if err != nil {
			fmt.Println(err)
		}
	}
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
	prof = log.NewBeegoLogger(false, app.GetProcessID(), defaultProfPath, mp)
	return nil
}

type ProfEvent struct {
	NodeId      string `json:"node_id"`
	ProcessId   int    `json:"process_id"`
	MchHostName string `json:"machine_name"` //机器名
	EventId     string `json:"event_id"`     //http请求、rpc请求、机器性能监控
	Api         string `json:"api"`

	//请求相关的数据
	Timestamp    int64   `json:"timestamp"` //时间戳，秒
	Count        int64   `json:"count"`
	SuccessCount int64   `json:"success_count"`
	FailCount    int64   `json:"fail_count"`
	FailRate     float64 `json:"fail_rate"`   //失败率
	SuccessAvg   float64 `json:"success_avg"` //成功平均延迟,单位秒，保留小数后面3位小数
	FailAvg      float64 `json:"fail_avg"`    //失败平均延迟,单位秒，保留小数后面3位小数
	P90          float64 `json:"p90"`         //过去一段时间内，最慢的10%的请求耗时的平均值，,单位秒，保留小数后面3位小数
	P99          float64 `json:"p99"`         // 过去一段时间内，最慢的1%的请求耗时的平均值，,单位秒，保留小数后面3位小数

	ProcMemUseRate float64 `json:"proc_mem_use_rate"` //当前进程内存使用率
	ProcCpuUseRate float64 `json:"proc_cpu_use_rate"` //当前进程CPU使用率
	MemUsedRate    float64 `json:"mem_used_rate"`     //内存使用率
	MemAvailable   float64 `json:"mem_available"`     //剩余可获取到的内存
	CpuUse         float64 `json:"cpu_use"`           //cpu使用率
	CpuLoad1       float64 `json:"cpu_load_1"`        //1,5,15分钟的cpu负载，阈值cpu核心数*0.7-0.8
	CpuLoad5       float64 `json:"cpu_load_5"`
	CpuLoad15      float64 `json:"cpu_load_15"`

	ServerMemUse   float64 `json:"server_mem_use"`   //当前服务从操作系统获得的内存总数.
	GcNum          uint32  `json:"gc_num"`           //服务垃圾回收的次数
	LastGc         uint64  `json:"last_gc"`          //服务上次gc的时间
	ServerStackUse float64 `json:"server_stack_use"` //服务正在使用的栈大小
	ServerHeapUse  float64 `json:"server_heap_use"`  //服务正在使用的堆的大小，字节
	GcSys          float64 `json:"gc_sys"`           // 垃圾回收元数据使用的内存字节数.

	NumGoroutine    int    `json:"num_goroutine"`      //服务内的协程数
	SysPauseTotalUs uint64 `json:"sys_pause_total_us"` //调用readmemstats所造成的goroutine暂停时间，微秒

	OnlineCount int               `json:"online_count"`
	ExtraParams map[string]string `json:"extra_info"`
	/**
	{"param01":"v1","param02":"v2"}
	extra_info字段为必需字段，字段值为json格式，
	主要是针对不同类型的后台重要补充信息的扩展，param01 param02字段名可以自定义，其值尽可能是字符串。
	字段数量可自行添加，如param03、param04。比如扩展字段可以包括url。字段中的value都需要是string类型。
	这样是为了避免 当不同后台扩展字段中出现同样的key时，value的数据类型不冲突。
	*/
}

func ProfReport(event ProfEvent) {
	prof.BiReport(utils.Struct2String(event))
}

type Observer interface {
	Watch()
}

func init() {
	observers = list.New()
}

//注册观察者
func RegisterWatcher(observe Observer) {
	observers.PushBack(observe)
}

func NotifyWatcher() { //通知所有观察者
	defer utils.CatchException(func(s string) {
		log.Error("Catch NotifyWatcher Error:%v", s)
	})
	for ob := observers.Front(); ob != nil; ob = ob.Next() {
		ob.Value.(Observer).Watch()
	}
}
