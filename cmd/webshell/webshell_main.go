// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/maoqide/kubeutil/kube"
	"github.com/maoqide/kubeutil/utils"
	"github.com/maoqide/kubeutil/webshell"
	"github.com/maoqide/kubeutil/webshell/wsterminal"
)

var (
	addr = flag.String("addr", ":8090", "http service address")
)

func internalError(ws *websocket.Conn, msg string, err error) {
	log.Println(msg, err)
	ws.WriteMessage(websocket.TextMessage, []byte("Internal server error."))
}

var upgrader = websocket.Upgrader{}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func serve(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	namspace := pathParams["namespace"]
	pod := pathParams["pod"]
	containerName := pathParams["container_name"]
	log.Printf("exec pod: %s, container: %s, namespace: %s", pod, containerName, namspace)

	pty, err := wsterminal.NewTerminalSession(w, r, nil)
	if err != nil {
		log.Printf("get pty failed: %v", err)
		return
	}
	defer func() {
		log.Println("close session.")
		pty.Close()
	}()
	kubeConfig, _ := utils.ReadFile("./config")
	cfg, _ := kube.LoadKubeConfig(kubeConfig)
	kubeC, _ := kube.NewKubeOutClusterClient(kubeConfig)
	err = webshell.ExecPod(kubeC, cfg, []string{"/bin/bash"}, pty, namspace, pod, containerName)
	if err != nil {
		log.Printf("exec err %v", err)
	}
	return
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", serveHome)
	router.HandleFunc("/ws/{namespace}/{pod}/{container_name}/webshell", serve)
	log.Fatal(http.ListenAndServe(*addr, router))

}
