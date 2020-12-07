/**
一定要记得在confin.json配置这个模块的参数,否则无法使用
*/
package mgate

import (
	"beta/echo/filters"
	"beta/tools"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/gate/base"
	"github.com/liangdas/mqant/gate/uriroute"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/url"
	"os"
	"time"
)
var logger *zap.Logger
var Module = func() module.Module {
	this := new(Gate)
	return this
}

type LocalUserData struct {
	ProHeartbeatTime time.Time
}

type Gate struct {
	basegate.Gate //继承
	RedisUrl      string
	Route         *uriroute.URIRoute
}

func (this *Gate) GetType() string {
	//很关键,需要与配置文件中的Module配置对应
	return "gate"
}
func (this *Gate) Version() string {
	//可以在监控时了解代码版本
	return "1.0.0"
}
func (this *Gate) OnInit(app module.App, settings *conf.ModuleSettings) {
	encoder:=zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	file, _ := os.Create("./test.log")
	writeSyncer := zapcore.AddSync(file)
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)

	logger = zap.New(core)
	route := uriroute.NewURIRoute(this,
		uriroute.Selector(this.Selector),
		uriroute.DataParsing(func(topic string, u *url.URL, msg []byte) (bean interface{}, err error) {
			return
		}),
		uriroute.CallTimeOut(10*time.Second),
	)
	//注意这里一定要用 gate.Gate 而不是 module.BaseModule
	this.Gate.OnInit(this, app, settings,
		gate.Heartbeat(time.Second*10),
		gate.BufSize(2048*2),
		gate.SetRouteHandler(route),
		gate.SetSessionLearner(this),
		gate.SetStorageHandler(this),
	)
	this.RedisUrl = settings.Settings["RedisUrl"].(string)
	_ = this.Gate.SetJudgeGuest(func(session gate.Session) bool {
		if session.GetUserIdInt64() <= 0 {
			return true
		}
		return false
	})
	this.GetServer().Options().Metadata["weight"] = "10"

	tools.RegisterWatcher(this)
	this.GetServer().RegisterGO("ping", this.ping)
}

//当连接建立  并且MQTT协议握手成功
func (this *Gate) Connect(session gate.Session) {
	_ = session.SetLocalUserData(&LocalUserData{
		ProHeartbeatTime: time.Now(),
	})
}

//当连接关闭	或者客户端主动发送MQTT DisConnect命令 ,这个函数中Session无法再继续后续的设置操作，只能读取部分配置内容了
func (this *Gate) DisConnect(session gate.Session) {
}

/**
存储用户的Session信息
Session Bind Userid以后每次设置 settings都会调用一次Storage
*/
func (this *Gate) Storage(session gate.Session) (err error) {
	return nil
}

/**
强制删除Session信息
*/
func (this *Gate) Delete(session gate.Session) (err error) {
	return
}

/**
获取用户Session信息
用户登录以后会调用Query获取最新信息
*/
func (this *Gate) Query(Userid string) ([]byte, error) {
	return nil, errors.New("tt")
}

/**
用户心跳,一般用户在线时60s发送一次
可以用来延长Session信息过期时间
*/
func (this *Gate) Heartbeat(session gate.Session) {
	//log.Info("用户在线的心跳包")
	//system://local/pingpong
	if session.GetUserIdInt64() <= 0 {
		_, errstr := this.Gate.GetGateHandler().Send(nil, session.GetSessionID(), "system://local/relogin/", []byte("{}"))
		//errstr:=session.Send("system://local/pingpong",b)
		if errstr != "" {
			log.Error("send relogin error %v", errstr)
		}
		return
	}
	userdata, ok := session.LocalUserData().(*LocalUserData)
	if ok {
		hbtime := time.Now().Sub(userdata.ProHeartbeatTime)
		if hbtime.Seconds() > 60*55 {
			//超过55分钟么有设置心跳了，设置一下
			userdata.ProHeartbeatTime = time.Now()
		}
	}
}

func (this *Gate) Watch() {
	onlineCount := this.GetGateHandler().GetAgentNum()
	mp := make(map[string]interface{})
	mp["online_count"] = onlineCount
	filters.UpdateProfExtraData(mp)
	return
}

func (this *Gate) ping(nodeId string, ts int64, traceId string) (result interface{}, errstr string) {
	//回调回去，传入的nodeId是master的id
	//logger := utils2.GetRootLogger()
	_, err := this.GetApp().RpcInvoke(this, nodeId, "pong", this.GetServerId(), ts, traceId)
	if err != "" {
		//logger.JError(utils2.LogMsg{EventId: "ping", SubEventId: "200", ErrorMsg: err, ErrorReport: "rpc Callback Error",
		//	EventParams: map[string]interface{}{"nodeId": nodeId, "traceId": traceId}})
		return ts, err
	}
	return ts, ""
}
