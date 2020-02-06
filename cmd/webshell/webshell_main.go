// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/maoqide/kubeutil/pkg/kube"
	"github.com/maoqide/kubeutil/webshell"
	"github.com/maoqide/kubeutil/webshell/wsterminal"
)

var (
	addr = flag.String("addr", ":8090", "http service address")
	cmd  = []string{"/bin/sh"}
)

func internalError(ws *websocket.Conn, msg string, err error) {
	log.Println(msg, err)
	ws.WriteMessage(websocket.TextMessage, []byte("Internal server error."))
}

func serveTerminal(w http.ResponseWriter, r *http.Request) {
	// auth
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "./frontend/terminal.html")
}

func serveLogs(w http.ResponseWriter, r *http.Request) {
	// auth
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "./frontend/logs.html")
}

func serveWsTerminal(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	namespace := pathParams["namespace"]
	podName := pathParams["pod"]
	containerName := pathParams["container_name"]
	log.Printf("exec pod: %s, container: %s, namespace: %s\n", podName, containerName, namespace)

	pty, err := wsterminal.NewTerminalSession(w, r, nil)
	if err != nil {
		log.Printf("get pty failed: %v\n", err)
		return
	}
	defer func() {
		log.Println("close session.")
		pty.Close()
	}()

	client, err := kube.GetClient()
	if err != nil {
		log.Printf("get kubernetes client failed: %v\n", err)
		return
	}
	pod, err := client.PodBox.Get(podName, namespace)
	if err != nil {
		log.Printf("get kubernetes client failed: %v\n", err)
		return
	}
	ok, err := webshell.ValidatePod(pod, containerName)
	if !ok {
		// msg := fmt.Sprintf("Invalid pod!! namespace: %s, pod: %s, container: %s", namespace, pod, containerName)
		msg := fmt.Sprintf("Validate pod error! err: %v", err)
		log.Println(msg)
		pty.Write([]byte(msg))
		pty.Done()
		return
	}
	err = client.PodBox.Exec(cmd, pty, namespace, podName, containerName)
	if err != nil {
		msg := fmt.Sprintf("Exec to pod error! err: %v", err)
		log.Println(msg)
		pty.Write([]byte(msg))
		pty.Done()
	}
	return
}

func main() {
	router := mux.NewRouter()
	router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./frontend/"))))
	router.HandleFunc("/terminal", serveTerminal)
	router.HandleFunc("/ws/{namespace}/{pod}/{container_name}/webshell", serveWsTerminal)
	// TODO
	router.HandleFunc("/ws/{namespace}/{pod}/{container_name}/logs", serveLogs)
	log.Fatal(http.ListenAndServe(*addr, router))
}
