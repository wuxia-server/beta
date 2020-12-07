package utils

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type ServerStatus int

type ExtraDataFunc func() map[string]interface{}
type StatusChangedFunc func(status ServerStatus, msg string)

const (
	ServerStatusOk ServerStatus = iota
	ServerStatusMongoDBError
	ServerStatusRedisError
	ServerStatusNATSError
	ServerStatusOtherError
)

var (
	once        = sync.Once{}
	monitor     *monitorController
	start       = time.Now() // 服务启动时间
	processId   = os.Getpid()
	hostname, _ = os.Hostname()
	cpuCores    = runtime.NumCPU()
	msg         = map[ServerStatus]string{
		ServerStatusOk:           "ok",
		ServerStatusMongoDBError: "mongoDB 链接出现异常",
		ServerStatusRedisError:   "redis 链接出现异常",
		ServerStatusNATSError:    "nats 链接出现异常",
		ServerStatusOtherError:   "发生未知错误",
	}
)

type monitorResponse struct {
	Code ServerStatus           `json:"code"` // 0为正常，非0为异常
	Msg  string                 `json:"msg"`  // 状态说明 "error","unknow","....."
	Name string                 `json:"name"` // 服务名称
	Data map[string]interface{} `json:"data"`
}

//额外字段
type extraData struct {
	ProcessId  int     `json:"process_id"`  //进程 id
	StartTime  int64   `json:"start_time"`  //服务启动时间，秒级时间戳
	Runtime    int64   `json:"runtime"`     //服务运行时间，单位秒
	CPUCores   int     `json:"cpu_cores"`   //cpu 核数
	Goroutine  int     `json:"goroutine"`   //协程数
	CPUUsed    float64 `json:"cpu_used"`    //cpu 使用率
	MemoryUsed float64 `json:"memory_used"` //内存使用率
	Hostname   string  `json:"hostname"`    //
}

type monitorController struct {
	name          string
	serverStatus  ServerStatus
	msg           string
	extraDataFunc ExtraDataFunc
	statusChanged StatusChangedFunc
}

func GetMonitor() *monitorController {
	once.Do(func() {
		monitor = new(monitorController)
		monitor.serverStatus = ServerStatusOk
		monitor.msg = msg[monitor.serverStatus]
	})
	return monitor
}

func (this *monitorController) SetName(serverName string) *monitorController {
	this.name = serverName
	return this
}

func (this *monitorController) GetName() string {
	return this.name
}

func (this *monitorController) SetExtra(extra ExtraDataFunc) *monitorController {
	this.extraDataFunc = extra
	return this
}

func (this *monitorController) GetCode() ServerStatus {
	return this.serverStatus
}

func (this *monitorController) SetServerStatus(status ServerStatus, errMsg string) *monitorController {
	this.serverStatus = status
	if errMsg != "" {
		this.msg = errMsg
	} else {
		this.msg = msg[status]
	}
	if this.statusChanged != nil {
		this.statusChanged(this.serverStatus, this.msg)
	}
	return this
}

func (this *monitorController) StatusChangedHandler(cb StatusChangedFunc) {
	this.statusChanged = cb
}

func (this *monitorController) RegisterRouter(group *echo.Group) {
	group.Any("/monitor.json", this.monitorHandle)
}

//状态信息
func (this *monitorController) monitorHandle(ctx echo.Context) error {

	resp := new(monitorResponse)
	resp.Name = this.name
	resp.Code = this.serverStatus
	resp.Msg = this.msg

	data := new(extraData)
	data.ProcessId = processId
	data.StartTime = start.Unix()
	data.Runtime = int64(time.Since(start).Seconds())
	data.CPUCores = cpuCores
	data.Goroutine = runtime.NumGoroutine()
	data.Hostname = hostname
	loads, err := load.Avg()
	if err == nil {
		data.CPUUsed = Decimal(loads.Load1)
	}
	memUse, err := mem.VirtualMemory()
	if err == nil {
		data.MemoryUsed = Decimal(memUse.UsedPercent)
	}
	resp.Data = Struct2Map(data)
	if this.extraDataFunc != nil {
		for k, v := range this.extraDataFunc() {
			resp.Data[k] = v
		}
	}

	return ctx.JSON(http.StatusOK, resp)
}

//保留3位小数
func Decimal(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", value), 64)
	return value
}
