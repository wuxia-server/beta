/**
一定要记得在confin.json配置这个模块的参数,否则无法使用
*/
package httpgateway

import (
	"fmt"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/httpgateway"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/module/base"
	"net/http"
)

var Module = func() module.Module {
	this := new(httpgate)
	return this
}

type httpgate struct {
	basemodule.BaseModule
	Port int
}

func (self *httpgate) GetType() string {
	//很关键,需要与配置文件中的Module配置对应
	return "httpgateway"
}
func (self *httpgate) Version() string {
	//可以在监控时了解代码版本
	return "1.0.0"
}
func (self *httpgate) OnInit(app module.App, settings *conf.ModuleSettings) {
	self.BaseModule.OnInit(self, app, settings)
	self.Port = int(self.GetModuleSettings().Settings["Port"].(float64))
}

func (self *httpgate) startHttpServer() *http.Server {
	log.Info("httpgate: startHttpServer HTTP server :%v",self.Port)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", self.Port),
		Handler: httpgateway.NewHandler(self.App),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			log.Info("Httpserver: ListenAndServe() error: %s", err)
		}
	}()
	// returning reference so caller can call Shutdown()
	return srv
}

func (self *httpgate) Run(closeSig chan bool) {
	log.Info("httpgate: starting HTTP server :%v",self.Port)
	srv := self.startHttpServer()
	<-closeSig
	log.Info("httpgate: stopping HTTP server")
	// now close the server gracefully ("shutdown")
	// timeout could be given instead of nil as a https://golang.org/pkg/context/
	if err := srv.Shutdown(nil); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}
	log.Info("httpgate: done. exiting")
}

func (self *httpgate) OnDestroy() {
	//别忘了继承
	self.BaseModule.OnDestroy()
}