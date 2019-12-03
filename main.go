package main

import (
	"flag"
	"fmt"
	"net/http"
	"log"
	"os"
)

var addr = flag.String("addr", ":80", "http service address")
var addrs = flag.String("addrs", ":443", "https service address")

func init()  {
	file := "./" +"signal"+ ".log"
	logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile) // 将文件设置为log输出的文件
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
	return
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	http.ServeFile(w, r, "home.html")
}

var hub *Hub

func main() {

	flag.Parse()
	hub = newHub()
	go hub.run()
	//http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/wss", wsHandler)
	http.HandleFunc("/", wsHandler)

	http.HandleFunc("/count", func(w http.ResponseWriter, r *http.Request) {
		//fmt.Printf("URL: %s\n", r.URL.String())
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte(fmt.Sprintf("%d", hub.ClientNum)))

	})

	if Exists("cert/crt.pem") && Exists("cert/crt.key") {
		go func() {
			log.Printf("Start to listening the incoming requests on https address: %s\n", *addrs)
			err := http.ListenAndServeTLS(*addrs, "cert/crt.pem", "cert/crt.key", nil)
			if err != nil {
				log.Fatal("ListenAndServe: ", err)
			}
		}()
	}

	log.Printf("Start to listening the incoming requests on http address: %s\n", *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}


}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	//fmt.Printf("URL: %s\n", r.URL.String())
	r.ParseForm()

	defer func() {                            // 必须要先声明defer，否则不能捕获到panic异常
		if err := recover(); err != nil {
			log.Println(err)                  // 这里的err其实就是panic传入的内容
		}
	}()

	id := r.Form.Get("id")
	if id != "" {
		serveWs(hub, w, r, id)
	}
}

// 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path)    //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}