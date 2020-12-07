/**
一定要记得在confin.json配置这个模块的参数,否则无法使用
*/
package webapp

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"beta/echo/filters"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/module/base"
	"github.com/liangdas/mqant/rpc/pb"
)

// Module web 模块
var Module = func() *WebApp {
	web := new(WebApp)
	return web
}

// Web 结构对象基于 BaseModule
type WebApp struct {
	basemodule.BaseModule
	StaticPath   string
	ResourcePath string
	Port         int
}

// GetType 获取模块类型标识
func (self *WebApp) GetType() string {
	//很关键,需要与配置文件中的Module配置对应
	return "webapp"
}

//Version 获取Web模块版本号
func (self *WebApp) Version() string {
	//可以在监控时了解代码版本
	return "1.0.0"
}

//OnInit Web模块初始化方法
func (self *WebApp) OnInit(app module.App, settings *conf.ModuleSettings) {
	self.BaseModule.OnInit(self, app, settings)
	self.StaticPath = self.GetModuleSettings().Settings["StaticPath"].(string)
	self.ResourcePath = self.GetModuleSettings().Settings["ResourcePath"].(string)
	self.Port = int(self.GetModuleSettings().Settings["Port"].(float64))
}

type req struct {
	id string
}

func (this *req) Marshal() ([]byte, error) {
	return []byte(this.id), nil
}
func (this *req) Unmarshal(data []byte) error {
	this.id = string(data)
	return nil
}
func (this *req) String() string {
	return "ss"
}

type rsp struct {
	id string
}

func (this *rsp) Marshal() ([]byte, error) {
	return []byte(this.id), nil
}
func (this *rsp) Unmarshal(data []byte) error {
	this.id = string(data)
	return nil
}
func (this *rsp) String() string {
	return "rsp"
}
func (self *WebApp) testMarshal(req req) (*rsp, error) {
	log.Info("testMarshal %v", req.id)
	r := &rsp{id: req.id}
	return r, nil
}
func (self *WebApp) testProto(req *rpcpb.ResultInfo) (*rpcpb.ResultInfo, error) {
	log.Info("testProto %v", req.Error)
	r := &rpcpb.ResultInfo{Error: *proto.String("hello Proto返回内容")}
	return r, nil
}
func registerFilter(e *echo.Echo) {
	// middleware
	e.Pre(filters.Before())
	e.Pre(filters.SetLogTrace())
	e.Use(filters.After())
	e.Use(filters.GetLogMiddleware())
	e.Use(middleware.Recover())

}

//Run Web模块启动方法
func (self *WebApp) Run(closeSig chan bool) {
	//这里如果出现异常请检查8080端口是否已经被占用

	e := echo.New()
	//注册路由
	registerFilter(e)
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
		AllowCredentials: true,
	}))
	e.Static("/doc", "apidoc")
	e.Static("/static", "static")
	e.Static("/static", "static")
	e.Static("/resource", self.ResourcePath)
	go func() {
		log.Info("webapp server Listen : %s", fmt.Sprintf(":%d", self.Port))
		// Start server
		e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", self.Port)))
	}()

	<-closeSig
	log.Info("webapp server Shutting down...")
	e.Close()
}

//OnDestroy Web模块注销方法
func (self *WebApp) OnDestroy() {
	//一定别忘了关闭RPC
	self.GetServer().OnDestroy()
}
