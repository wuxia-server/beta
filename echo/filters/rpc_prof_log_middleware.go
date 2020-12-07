package filters

import (
	mqrpc "github.com/liangdas/mqant/rpc"
	rpcpb "github.com/liangdas/mqant/rpc/pb"
	"os"
	"sync"
	"time"
	"beta/tools"
)

var (
	rpcMap       = sync.Map{}
	rpcLk        = sync.Mutex{}
	rpcClientMap = sync.Map{}
	rpcClientLk  = sync.Mutex{}
)

//客户端的监控
func RpcClientListener(NodeId string, rpcinfo rpcpb.RPCInfo, result interface{}, err string, exec_time int64) {
	//取出时间进行计算
	defer Catch("RpcClientListener", "")
	now := time.Now().Unix()
	//转换成微秒
	Cost := int(exec_time / 1e3)
	Topic := rpcinfo.Fn
	//先取时间戳对应的值
	rpcClientLk.Lock()
	defer rpcClientLk.Unlock()
	//存储结构:  时间戳:{节点id:{方法名:{成功[]、失败[]}}}
	if vals, ok := rpcClientMap.Load(now); ok {
		//有该秒的时间戳时则取对应nodeId的值，
		if val, ok := vals.(map[string]map[string]*PerSecond)[NodeId]; ok {
			//有NodeId时，取其接口的值
			if topic, ok := val[Topic]; ok {
				//有对应接口的值则取出添加该次请求的数据
				if err == "" {
					topic.Success = append(topic.Success, Cost)
				} else {
					topic.Fail = append(topic.Fail, Cost)
				}
				rpcClientMap.Store(now, vals)
			} else {
				//没有对应接口时，则创建该接口的数据map
				ps := &PerSecond{}
				if err == "" {
					ps.Success = append(ps.Success, Cost)
				} else {
					ps.Fail = append(ps.Fail, Cost)
				}
				vals.(map[string]map[string]*PerSecond)[NodeId][Topic] = ps
				rpcClientMap.Store(now, vals)
			}
		} else {
			//没有该NodeId时创建该NodeId的值
			ps := &PerSecond{}
			if err == "" {
				ps.Success = append(ps.Success, Cost)
			} else {
				ps.Fail = append(ps.Fail, Cost)
			}
			fn := map[string]*PerSecond{
				Topic: ps,
			}
			vals.(map[string]map[string]*PerSecond)[NodeId] = fn
			rpcClientMap.Store(now, vals)
		}
	} else {
		//没有该秒的时间戳时创建该秒，该nodeId，该方法的map
		ps := &PerSecond{}
		if err == "" {
			ps.Success = append(ps.Success, Cost)
		} else {
			ps.Fail = append(ps.Fail, Cost)
		}
		fn := map[string]*PerSecond{
			Topic: ps,
		}
		v := make(map[string]map[string]*PerSecond)
		v[NodeId] = fn
		rpcClientMap.Store(now, v)
	}
}

//服务端的监控
func RpcServerListener(NodeId string, callInfo *mqrpc.CallInfo, result *rpcpb.ResultInfo, exec_time int64) {
	defer Catch("RpcServerListener", "")
	//取出时间进行计算
	now := time.Now().Unix()
	//转换成微秒
	Cost := int(exec_time / 1e3)
	Topic := callInfo.RPCInfo.Fn
	//先取时间戳对应的值
	rpcLk.Lock()
	defer rpcLk.Unlock()
	if vals, ok := rpcMap.Load(now); ok {
		//有该秒的时间戳时则取对应nodeId的值，
		if val, ok := vals.(map[string]map[string]*PerSecond)[NodeId]; ok {
			//有NodeId时，取其接口的值
			if topic, ok := val[Topic]; ok {
				//有对应接口的值则取出添加该次请求的数据
				if result.Error == "" {
					topic.Success = append(topic.Success, Cost)
				} else {
					topic.Fail = append(topic.Fail, Cost)
				}
				rpcMap.Store(now, vals)
			} else {
				//没有对应接口时，则创建该接口的数据map
				ps := &PerSecond{}
				if result.Error == "" {
					ps.Success = append(ps.Success, Cost)
				} else {
					ps.Fail = append(ps.Fail, Cost)
				}

				vals.(map[string]map[string]*PerSecond)[NodeId][Topic] = ps
				rpcMap.Store(now, vals)
			}
		} else {
			//没有该NodeId时创建该NodeId的值
			ps := &PerSecond{}
			if result.Error == "" {
				ps.Success = append(ps.Success, Cost)
			} else {
				ps.Fail = append(ps.Fail, Cost)
			}
			fn := map[string]*PerSecond{
				Topic: ps,
			}
			vals.(map[string]map[string]*PerSecond)[NodeId] = fn
			rpcMap.Store(now, vals)
		}
	} else {
		//没有该秒的时间戳时创建该秒，该nodeId，该方法的map
		ps := &PerSecond{}
		if result.Error == "" {
			ps.Success = append(ps.Success, Cost)
		} else {
			ps.Fail = append(ps.Fail, Cost)
		}
		fn := map[string]*PerSecond{
			Topic: ps,
		}
		v := make(map[string]map[string]*PerSecond)
		v[NodeId] = fn
		rpcMap.Store(now, v)
	}
}

//客户端的监控计算
func RpcClientSummary() {
	defer Catch("Rpc Client Summary", "")
	//创建定时器，每隔1秒后，定时器就会给channel发送一个事件(当前时间)
	now := time.Now().Unix() - 2
	if rst, ok := rpcClientMap.Load(now); ok {
		for node, topics := range rst.(map[string]map[string]*PerSecond) {
			for topic, ps := range topics {
				success := ps.Success
				fail := ps.Fail
				all := append(success, fail...)
				if len(all) == 0 {
					return
				}
				st := new(tools.ProfEvent)
				st.ProcessId = os.Getpid()
				//一些物理指标
				st.MchHostName, _ = os.Hostname()
				st.EventId = "RpcClientReq"
				st.Api = topic
				st.NodeId = node
				st.Timestamp = now
				st.SuccessCount = int64(len(success))
				st.FailCount = int64(len(fail))
				st.Count = int64(len(all))
				st.SuccessAvg = GetAvg(success)
				st.FailAvg = GetAvg(fail)
				st.FailRate = (float64(len(fail)) / float64(len(all))) * 100
				st.P90 = GetP90(all)
				st.P99 = GetP99(all)
				GetProfExtraData(st)
				//打印可被收集日志
				tools.ProfReport(*st)
			}
		}
		//清理2秒之前的所有数据
		rpcClientMap.Range(func(key, value interface{}) bool {
			if key.(int64) <= now {
				rpcClientMap.Delete(key)
			}
			return true
		})
	}
}

//服务端的监控计算
func RpcServerSummary() {
	defer Catch("Rpc Server Summary", "")
	//创建定时器，每隔1秒后，定时器就会给channel发送一个事件(当前时间)
	now := time.Now().Unix() - 2
	if rst, ok := rpcMap.Load(now); ok {
		for node, topics := range rst.(map[string]map[string]*PerSecond) {
			for topic, ps := range topics {
				success := ps.Success
				fail := ps.Fail
				all := append(success, fail...)
				if len(all) == 0 {
					return
				}
				st := new(tools.ProfEvent)
				st.ProcessId = os.Getpid()
				st.MchHostName, _ = os.Hostname()
				st.EventId = "RpcReq"
				st.Api = topic
				st.NodeId = node
				st.Timestamp = now
				st.SuccessCount = int64(len(success))
				st.FailCount = int64(len(fail))
				st.Count = int64(len(all))
				st.SuccessAvg = GetAvg(success)
				st.FailAvg = GetAvg(fail)
				st.FailRate = (float64(len(fail)) / float64(len(all))) * 100
				st.P90 = GetP90(all)
				st.P99 = GetP99(all)
				GetProfExtraData(st)
				//打印可被收集日志
				tools.ProfReport(*st)
			}
		}
		//清理第前2秒的数据
		rpcMap.Range(func(key, value interface{}) bool {
			if key.(int64) <= now {
				rpcMap.Delete(key)
			}
			return true
		})
	}
}

//清理统计过了的数据
func ClearRpcClientMap(ts int64) {
	rpcClientMap.Delete(ts)
}

//清理统计过了的数据
func ClearRpcMap(ts int64) {
	rpcMap.Delete(ts)
}
