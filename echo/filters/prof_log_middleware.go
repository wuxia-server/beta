package filters

import (
	"fmt"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/registry"
	utils2 "github.com/liangdas/mqant/utils"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
	"os"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo"
	"beta/tools"
)

var (
	smp         = sync.Map{}
	NodeId      = ""
	lk          = sync.Mutex{}
	watcherLock = sync.Mutex{}
	extraData   = make(map[string]interface{})
)

//使用sort排序，参数必须为[]int
type PerSecond struct {
	Success []int `json:"success"`
	Fail    []int `json:"fail"`
}

//处理之前的中间件
func Before() echo.MiddlewareFunc {
	return func(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) (err error) {
			//记录请求初始时间戳，纳秒
			ctx.Set("start", time.Now().UnixNano())
			return handlerFunc(ctx)
		}
	}
}

// 处理之后的中间件
func After() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer Catch("Http After", "")
			if err := next(c); err != nil {
				c.Error(err)
			}
			//计算请求耗时
			now := time.Now()
			nowUnix := now.Unix()
			nowNano := now.UnixNano()
			startNano := c.Get("start").(int64)
			Cost := int((nowNano - startNano) / 1e3) //转换成微秒

			UrlPath := c.Request().URL.Path
			//先取时间戳对应的值
			lk.Lock()
			defer lk.Unlock()
			if vals, ok := smp.Load(nowUnix); ok {
				//有值则取对应接口的值，
				if val, ok := vals.(map[string]*PerSecond)[UrlPath]; ok {
					//有对应接口的值则取出添加该次请求的数据
					if c.Response().Status == 200 {
						val.Success = append(val.Success, Cost)
						smp.Store(nowUnix, vals)
					} else {
						val.Fail = append(val.Fail, Cost)
						smp.Store(nowUnix, vals)
					}
				} else {
					ps := &PerSecond{}
					if c.Response().Status == 200 {
						ps.Success = append(ps.Success, Cost)
					} else {
						ps.Fail = append(ps.Fail, Cost)
					}
					vals.(map[string]*PerSecond)[UrlPath] = ps
					smp.Store(nowUnix, vals)
				}
			} else {
				ps := &PerSecond{}
				if c.Response().Status == 200 {
					ps.Success = append(ps.Success, Cost)
				} else {
					ps.Fail = append(ps.Fail, Cost)
				}
				v := map[string]*PerSecond{
					UrlPath: ps,
				}
				smp.Store(nowUnix, v)
			}
			return nil
		}
	}
}

//在最开始可以调用
func ProfReport() {
	//启用一个go协程，不阻塞后续逻辑
	go func() {
		ticker := time.NewTicker(time.Second * 1)
		for {
			<-ticker.C
			tools.NotifyWatcher()
			//一下为每个模块启动go程，是为了前面的模块不阻塞后面的模块计算,
			// 下次优化点，每个模块从始至终一个协程，通过channel接收信息开始打点，不用每秒生成4个协程，然后销毁

			go HttpSummary()
			go RpcServerSummary()
			go MachineProfSummary()
			go RpcClientSummary()
		}
	}()

}

func NodeDelayProfReport(rpcModule module.RPCModule, interval int) {
	go NodeDelayPing(rpcModule, interval)
}

//rpc节点的延迟探测
func NodeDelayPing(app module.RPCModule, interval int) {
	defer Catch("Node Delay Ping", "")
	//每x秒探测一次
	if interval <= 0 {
		interval = 10
	}
	d := time.Duration(interval)
	ticker := time.NewTicker(time.Second * d)
	for {
		<-ticker.C
		traceId := utils2.GenerateID().String()
		//查出所有的节点
		slist, err := app.GetApp().Options().Registry.ListServices()
		if err != nil {
			log.Error("Node Delay Test Get List Services Error:%v,App:%v,TraceId:%v", err, app.GetServerId(), traceId)
			continue
		}
		services := make([]*registry.Service, 0)
		for _, service := range slist {
			newservices, err := app.GetApp().Options().Registry.GetService(service.Name)
			if err != nil {
				log.Error("Node Delay Test Get Services Error:%v,App:%v,TraceId:%v", err, app.GetServerId(), traceId)
				continue
			}
			for _, s := range newservices {
				services = append(services, s)
			}
		}
		//探测所有节点的延迟
		var wg = sync.WaitGroup{}
		for _, service := range services {
			for _, node := range service.Nodes {
				if node.Id == "consul" {
					continue
				}
				wg.Add(1)
				go func() {
					defer Catch("Node Delay Ping", traceId)
					ch := make(chan interface{})
					timeOut := time.NewTimer(time.Second * 2)
					go func() {
						defer Catch("Node Delay Ping", traceId)
						//带上master的id，和此次探测的traceid
						now := time.Now().UnixNano()
						_, errstr := app.GetApp().RpcInvoke(app, node.Id, "ping", app.GetServerId(), now, traceId)
						if errstr != "" {
							log.Error("Node Delay Test RPC Error:%v,NodeId:%v,TraceId:%v", errstr, node.Id, traceId)
							NodeDelayError(node.Id, -1)
							ch <- false
							return
						}
						ch <- true
					}()
					select {
					case <-ch:

					case <-timeOut.C:
						log.Error("node delay rpc timeout:%v,traceId:%v", node.Id, traceId)
						NodeDelayError(node.Id, -2)
					}
					//timer用完必须释放，否则
					timeOut.Stop()
					wg.Done()
					return
				}()
			}
		}
		wg.Wait()
	}
}

func HttpSummary() {
	defer Catch("Http Summary", "")
	//创建定时器，每隔1秒后，定时器就会给channel发送一个事件(当前时间)
	now := time.Now().Unix() - 2
	if rst, ok := smp.Load(now); ok {
		for key, val := range rst.(map[string]*PerSecond) {
			success := val.Success
			fail := val.Fail
			all := append(success, fail...)
			if len(all) == 0 {
				return
			}
			st := new(tools.ProfEvent)
			st.ProcessId = os.Getpid()
			//一些物理指标
			st.MchHostName, _ = os.Hostname()
			st.Timestamp = now
			st.EventId = "HttpReq"
			st.Api = key
			st.Count = int64(len(all))
			st.SuccessCount = int64(len(success))
			st.FailCount = int64(len(fail))
			st.SuccessAvg = GetAvg(success)
			st.FailAvg = GetAvg(fail)
			st.FailRate = (float64(len(fail)) / float64(len(all))) * 100
			st.P90 = GetP90(all)
			st.P99 = GetP99(all)
			GetProfExtraData(st)
			//打印可被收集日志
			tools.ProfReport(*st)
		}
		//清理第前2秒的数据
		smp.Range(func(key, value interface{}) bool {
			if key.(int64) <= now {
				smp.Delete(key)
			}
			return true
		})
	}
}

//保留3位小数
func Decimal(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", value), 64)
	return value
}

//计算平均值,秒，保留到后面3位小数
//传入的是[]微秒
func GetAvg(target []int) float64 {
	if len(target) == 0 {
		return 0
	}
	rst := 0
	for _, v := range target {
		rst += v
	}
	if rst != 0 {

		return Decimal((float64(rst) / float64(len(target))) / 1e6)
	}
	return 0
}

//计算P90延迟，即长尾请求的延迟
//P90：一段时间内最慢的10%的请求的平均耗时
//传入[]微秒
func GetP90(target []int) float64 {
	if len(target) == 0 {
		return 0
	}
	//先给列表排序
	sort.Ints(target)
	//计算最后10%的索引
	length := float64(len(target))
	index := int(length * 0.9)
	p90 := target[index:]
	sum := 0
	for _, v := range p90 {
		sum += v
	}
	if len(p90) != 0 {
		//微秒转换成秒后，保留3位小数,
		return Decimal((float64(sum) / float64(len(p90))) / 1e6)
	}
	return 0
}

//计算P99延迟，即长尾请求的延迟
//P99：一段时间内最慢的1%的请求的平均耗时
func GetP99(target []int) float64 {
	if len(target) == 0 {
		return 0
	}
	//先给列表排序
	sort.Ints(target)
	//计算最后1%的索引
	length := float64(len(target))
	index := int(length * 0.99)
	p99 := target[index:]
	sum := 0
	for _, v := range p99 {
		sum += v
	}
	if len(p99) != 0 {
		//微秒转换成秒后，保留3位小数,
		return Decimal((float64(sum) / float64(len(p99))) / 1e6)
	}
	return 0
}

//清理统计过了的数据
func Clear(ts int64) {
	smp.Delete(ts)
}

//内存单位都为M，保留3位小数
func Byte2MB(s uint64) float64 {
	if s == 0 {
		return 0
	} else {
		return Decimal(float64(s) / float64(1024*1024))
	}
}

//机器性能统计
func MachineProfSummary() {
	defer Catch("Machine Prof Summary", "")
	//创建定时器，每隔1秒后，定时器就会给channel发送一个事件(当前时间)
	//现把时间取了，获取profinfo是耗时操作
	now := time.Now().Unix()
	prof := GetProfInfo()
	prof.Timestamp = now
	prof.EventId = "MachineProf"
	GetProfExtraData(prof)
	tools.ProfReport(*prof)
}

//获取该机器目前性能信息
func GetProfInfo() *tools.ProfEvent {
	st := new(tools.ProfEvent)
	st.ProcessId = os.Getpid()
	//一些物理指标
	st.MchHostName, _ = os.Hostname()
	//负载
	loads, _ := load.Avg()
	st.CpuLoad1 = Decimal(loads.Load1)
	st.CpuLoad5 = Decimal(loads.Load5)
	st.CpuLoad15 = Decimal(loads.Load15)
	memUse, err := mem.VirtualMemory()
	if err != nil {
		log.Info("Get Memory Error:%v", err.Error())
		return st
	}
	//cpu占用率-系统级
	c, _ := cpu.Percent(time.Second, false)
	if len(c) != 0 {
		st.CpuUse = Decimal(c[0])
	}
	//内存使用率，可用-系统级
	st.MemUsedRate = Decimal(memUse.UsedPercent)

	st.MemAvailable = Byte2MB(memUse.Available)
	pid, err := process.NewProcess(int32(st.ProcessId))
	if err == nil {
		//进程级cpu利用率
		cpuUse, err := pid.CPUPercent()
		if err == nil {
			st.ProcCpuUseRate = Decimal(cpuUse)
		}
		//进程级内存利用率
		procMem, err := pid.MemoryPercent()
		if err == nil {
			st.ProcMemUseRate = Decimal(float64(procMem))
		}
	}

	stats := &runtime.MemStats{}
	runtime.ReadMemStats(stats)

	st.ServerStackUse = Byte2MB(stats.StackInuse)
	st.ServerHeapUse = Byte2MB(stats.HeapInuse)
	st.ServerMemUse = Decimal(st.ServerHeapUse + st.ServerStackUse)
	st.LastGc = stats.LastGC
	st.GcNum = stats.NumGC

	st.NumGoroutine = runtime.NumGoroutine()
	st.SysPauseTotalUs = stats.PauseTotalNs / 1e3
	st.GcSys = Byte2MB(stats.GCSys)
	return st
}

//延迟为-1代表请求失败，延迟为-2代表请求超时
func NodeDelayError(nodeId string, delay int64) {
	bidata := new(tools.ProfEvent)
	bidata.NodeId = nodeId
	bidata.Timestamp = time.Now().Unix()
	bidata.SuccessAvg = float64(delay)
	bidata.EventId = "NodeDelay"
	hn, err := os.Hostname()
	if err == nil {
		bidata.MchHostName = hn
	}
	bidata.ProcessId = os.Getpid()
	GetProfExtraData(bidata)
	tools.ProfReport(*bidata)
}

func Catch(loc, traceId string) {
	if r := recover(); r != nil {
		log.Error("Catch %v Error:%v,Trace:%v", loc, r, traceId)
	}
	return
}

func UpdateProfExtraData(extra map[string]interface{}) {
	if extra != nil {
		watcherLock.Lock()
		for key, value := range extra {
			extraData[key] = value
		}
		watcherLock.Unlock()
	}
	return
}

func GetProfExtraData(event *tools.ProfEvent) {
	if event != nil {
		if event.ExtraParams == nil {
			event.ExtraParams = make(map[string]string)
		}
		if extraData != nil {
			watcherLock.Lock()
			for key, value := range extraData {
				event.ExtraParams[key] = fmt.Sprintf("%v", value)
				if key == "online_count" {
					if v, ok := value.(int); ok {
						event.OnlineCount = v
					}
				}
			}
			watcherLock.Unlock()
		}
	}
}
