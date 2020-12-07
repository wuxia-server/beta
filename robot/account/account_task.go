 // Copyright 2014 hey Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

 import (
	 "github.com/liangdas/armyant/task"
	 "github.com/prometheus/common/log"
	 "beta/robot/account/work"
	 "os"
	 "os/signal"
 )

func main() {

	task := task.LoopTask{
		C: 1, //并发数
	}
	manager:= work.NewManager(task) //房间模型的demo
	log.Info("开始压测请等待")
	task.Run(manager)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	task.Stop()
	os.Exit(1)
}
