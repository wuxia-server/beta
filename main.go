package main

import (
	"beta/account"
	"beta/gate"
	"beta/httpgateway"
	"beta/webapp"
	"errors"
	"fmt"
	"beta/echo/filters"
	"beta/tools"
	"time"
	"github.com/liangdas/mqant"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/log/beego"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/rpc"
	"github.com/liangdas/mqant/rpc/util"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
	"net/http"
	_ "net/http/pprof"
	response_pb "github.com/wuxia-server/protobuf/golang/message"
)

const timeFormat = "20060102T150405Z"
func main() {
	go func() {
		//http://127.0.0.1:7070/debug/pprof/
		http.ListenAndServe("0.0.0.0:7070", nil)
	}()
	app := mqant.CreateApp(
		module.Debug(true), //是否开启debug模式
		module.KillWaitTTL(time.Second*15),
		module.RegisterTTL(time.Second*6),
		module.RegisterInterval(time.Second*3),
		module.RPCExpired(5*time.Second),
		//module.Configure("/work/go/joy-go/bin/conf/server.json"), // 配置
		//module.ProcessID("development"), //模块组ID
		//module.SetClientRPChandler(func(app module.App, server registry.Node, rpcinfo *rpcpb.RPCInfo, result interface{}, err string, exec_time int64) {
		//}),
		module.SetRpcCompleteHandler(func(app module.App, module module.Module, callInfo *mqrpc.CallInfo, input []interface{}, out []interface{}, execTime time.Duration) {
		}),
		//module.SetServerRPCHandler(func(app module.App, server module.Module, callInfo *mqrpc.CallInfo) {
		//}),
	)
_:
	app.SetProtocolMarshal(func(Trace string, Result interface{}, Error string) (module.ProtocolMarshal, string) {
		var result []byte
		if Result != nil {
			//内容不为空,尝试转为[]byte
			switch v2 := Result.(type) {
			case module.ProtocolMarshal:
				result = v2.GetData()
			default:
				_, r, err := argsutil.ArgsTypeAnd2Bytes(app, Result)
				if err != nil {
					Error = err.Error()
				}
				result = r
			}
		}
		r := &response_pb.S2C_Response{
			Error:  *proto.String(Error),
			Trace:  *proto.String(Trace),
			Result: result,
		}
		b, err := proto.Marshal(r)
		if err == nil {
			//解析得到[]byte后用NewProtocolMarshal封装为module.ProtocolMarshal
			return app.NewProtocolMarshal(b), ""
		} else {
			return nil, err.Error()
		}
	})
_:
	app.OnConfigurationLoaded(func(app module.App) {
		fmt.Println(time.Now().UTC().Format(timeFormat))
		natsconfig := app.GetSettings().Settings["Nats"].(map[string]interface{})
		nc, err := nats.Connect(natsconfig["uri"].(string), nats.UserInfo(natsconfig["user"].(string), natsconfig["password"].(string)), nats.MaxReconnects(10000))
		if err != nil {
			panic(errors.New("nats 服务连接异常"))
		}
		app.UpdateOptions(module.Nats(nc))

	})
_:
	app.OnStartup(func(app module.App) {
		_ = tools.InitProfReport(app, "", "")
		filters.ProfReport()
		log.LogBeego().SetFormatFunc(logs.DefineErrorLogFunc(app.GetProcessID(), 4))
		//监听房间信息变化
	})
_:
	app.Run(
		account.Module(),
		mgate.Module(),
		httpgateway.Module(),
		webapp.Module(),
	)
}

