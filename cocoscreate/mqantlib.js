const paho = require('./paho-mqtt');

var hashmap = function () { };
hashmap.prototype = {
    constructor: hashmap,
    add: function (k, v) {
        if (!this.hasOwnProperty(k)) {
            this[k] = v;
        }
    },
    remove: function (k) {
        if (this.hasOwnProperty(k)) {
            delete this[k];
        }
    },
    has: function (k) {
        var type = typeof k;
        if (type === 'string' || type === 'number') {
            return this.hasOwnProperty(k);
        } else if (type === 'function' && this.some(k)) {
            return true;
        }
        return false;
    },
    clear: function () {
        for (var k in this) {
            if (this.hasOwnProperty(k)) {
                delete this[k];
            }
        }
    },
    each: function (fn) {
        for (var k in this) {
            if (this.hasOwnProperty(k)) {
                fn.call(this, this[k], k, this);
            }
        }
    },
    join: function (split) {
        split = split !== undefined ? split : ',';
        var rst = [];
        this.each(function (v) {
            rst.push(v);
        });
        return rst.join(split);
    },
    some: function (fn) {
        for (var k in this) {
            if (this.hasOwnProperty(k)) {
                if (fn.call(this, this[k], k, this)) {
                    return true;
                }
            }
        }
        return false;
    },
    find: function (k) {
        var type = typeof k;
        if (type === 'string' || type === 'number' && this.has(k)) {
            return this[k];
        } else if (type === 'function') {
            for (var _k in this) {
                if (this.hasOwnProperty(_k) && k.call(this, this[_k], _k, this)) {
                    return this[_k];
                }
            }
        }
        return null;
    }
};
//添加url参数
function addParam(url, paramKey, paramVal) {
    var andStr = "?";
    var beforeparam = url.indexOf("?");
    if (beforeparam != -1) {
        andStr = "&";
    }
    return url + andStr + paramKey + "=" + encodeURIComponent(paramVal);
}
//获取url参数
function getParam(url, paramKey) {
    //获取要取得的get参数位置
    var get = url.indexOf(paramKey + "=");
    if (get == -1) {
        return null;
    }
    //截取字符串
    var getParamStr = url.slice(paramKey.length + get + 1);
    //判断截取后的字符串是否还有其他get参数
    var nextparam = getParamStr.indexOf("&");
    if (nextparam != -1) {
        getParamStr = getParamStr.slice(0, nextparam);
    }
    return decodeURIComponent(getParamStr);
}

var mqantlib = function (mqtt) {
    var mqant = cc.Class({
        properties: {
            curr_id: 0,
            waiting_queue:null,
        },
        ctor:function(){
            this.waiting_queue=new hashmap();
        },
        init:function(prop,context){
            var self=this;
            self.connectcallback=prop["connect"];
            self.errorcallback=prop["error"];
            self.closecallback=prop["close"];
            self.reconnectcallback=prop["reconnect"];
            self.context=context;
            if((self.client!=null)&&self.client.connected){
                return true;
            }
            prop["connect"]=function () {
                self.client.connected=true;
                var args = new Array();
                for(var k in arguments){
                    args.push(arguments[k]);
                }
                if(self.connectcallback){
                    self.connectcallback.apply(self.context,args)
                }
            }
            prop["error"]=function () {
                self.client.connected=false;
                var args = new Array();
                for(var k in arguments){
                    args.push(arguments[k]);
                }
                if(self.errorcallback){
                    self.errorcallback.apply(self.context,args)
                }
            }

            prop["close"]=function () {
                self.client.connected=false;
                var args = new Array();
                for(var k in arguments){
                    args.push(arguments[k]);
                }
                if(self.closecallback){
                    self.closecallback.apply(self.context,args)
                }
            }

            prop["reconnect"]=function () {
                self.client.connected=false;
                var args = new Array();
                for(var k in arguments){
                    args.push(arguments[k]);
                }
                if(self.reconnectcallback){
                    self.reconnectcallback.apply(self.context,args)
                }
            }

            // this.client = mqtt.connect(prop["uri"],{
            
            // //var client = mqtt.connect("egret://127.0.0.1:3653",{
            // //var client = mqtt.connect("laya://127.0.0.1:3653",{
            //     protocolId: 'MQIsdp',
            //     protocolVersion: 3,
            //     clientId:'mqttjs_' + Math.random().toString(16).substr(2, 8),
            //     reconnectPeriod:0, //不自动重连
            // }) // you add a ws:// url here
            // this.client.on('connect', prop["connect"]);
            // this.client.on('reconnect', prop["reconnect"]);
            // this.client.on('error', prop["error"]);
            // this.client.on('close', prop["close"]);
            // this.client.on("message", onMessageArrived);
            // var self=this;
            // function onMessageArrived(topic, payload) {
            //     try{
            //         var callback=self.waiting_queue.find(topic);
            //         if(callback!=null){
            //             //有等待消息的callback 还缺一个信息超时的处理机制
            //             var h=topic.split("/");
            //             if(h.length>2){
            //                 //这个topic存在msgid 那么这个回调只使用一次
            //                 self.waiting_queue.remove(topic)
            //             }
            //             callback["callback"].call(callback["callbackContext"],topic,payload)
            //         }
            //     }catch(e) {
            //         console.log(e);
            //     }
            //}
            prop["useSSL"]=prop["useSSL"]||false
            prop["host"]=prop["host"]||""
            prop["port"]=prop["port"]||0
            prop["path"]=prop["path"]||"/mqtt"
            prop["client_id"]=prop["client_id"]||'mqttjs_' + Math.random().toString(16).substr(2, 8);
            // self.client = new mqtt.Client(prop["host"], prop["port"], prop["client_id"]);
            self.client = new mqtt.Client(prop["host"], prop["port"], prop["path"], prop["client_id"]);
            var connectOptions={
                //onSuccess: prop["connect"],
                onFailure: prop["error"],
                mqttVersion: 3,
                useSSL:prop["useSSL"],
                cleanSession: true,
                reconnect:true,
                timeout:2,
                keepAliveInterval:2,
            }
            if(prop["uri"]){
                connectOptions.uris=[prop["uri"]]
            }
            if(prop["host"]){
                connectOptions.hosts=[prop["host"]]
            }
            if(prop["port"] != undefined){
                connectOptions.ports=[prop["port"]]
            }
            self.client.connect(connectOptions);//连接服务器并注册连接成功处理事件
            self.client.onConnected=prop["connect"];
            self.client.onConnectionLost =prop["close"] ;//注册连接断开处理事件
            self.client.onMessageArrived = onMessageArrived;//注册消息接收处理事件
            
            
            function onMessageArrived(message) {
                try {
                    var callback = self.waiting_queue.find(message.destinationName);
                    if (callback != null) {
                        //有等待消息的callback 还缺一个信息超时的处理机制
                        if (!getParam(message.destinationName, 'msg_id') === null) {
                            //这个topic存在msgid 那么这个回调只使用一次
                            self.waiting_queue.remove(message.destinationName)
                        }
                        callback["callback"].call(callback["callbackContext"], message.destinationName, message.payloadBytes)
                    }
                } catch (e) {
                    console.log(e);
                }
            }
        },
        connected:function(){
            if((this.client!=null)&&this.client.connected){
                return true;
            }
            return false;
        },
        /**
         * 向服务器发送一条消息
         * @param topic
         * @param msg
         * @param callback
         */
        request:function(topic,msg,callback,callbackContext){
            this.curr_id = this.curr_id + 1;
            var topic = addParam(topic, "msg_id", this.curr_id); //给topic加一个msgid 这样服务器就会返回这次请求的结果,否则服务器不会返回结果
            // var payload = JSON.stringify(msg);
            // var payload = msg.buffer;
            // console.log("check buffer: " + payload);    
            if (typeof callback !== 'function') {
                callback = function () { };
            }
            this.on(topic, callback, callbackContext);
            this.client.publish(topic, msg, 0, false);
        },
        /**
         * 向服务器发送一条消息,但不要求服务器返回结果
         * @param topic
         * @param msg
         */
        requestNR:function(topic,msg){
            var payload=JSON.stringify(msg)
            this.client.publish(topic,payload ,0,false);
        },
        /**
         * 监听指定类型的topic消息
         * @param topic
         * @param callback
         */
        on:function(topic,callback,callbackContext){
            //服务器不会返回结果
            if(callbackContext===null){
                callbackContext=this;
            }
            this.waiting_queue.remove(topic);
            this.waiting_queue.add(topic,{
                "callback":callback,
                "callbackContext":callbackContext
            }) //添加这条消息到等待队列
        },
        clearCallback:function () {
            this.waiting_queue.clear();
        },
        destroy:function(){
            this.client.disconnect();
            this.waiting_queue.clear();
        },
        parseUTF8:function(payload){
            if (typeof payload === "string")
                return payload;
            else
                return mqtt.ParseUTF8(payload, 0, payload.length);
        }
    })
    return mqant
};
var mqant = mqantlib(paho);

module.exports = mqant;