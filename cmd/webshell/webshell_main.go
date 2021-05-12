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
	corev1 "k8s.io/api/core/v1"

	_ "github.com/maoqide/kubeutil/initialize"
	"github.com/maoqide/kubeutil/pkg/kube"
	kubeLog "github.com/maoqide/kubeutil/pkg/kube/log"
	terminal "github.com/maoqide/kubeutil/pkg/terminal"
	wsterminal "github.com/maoqide/kubeutil/pkg/terminal/websocket"
	"github.com/maoqide/kubeutil/utils"
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
	containerName := pathParams["container"]
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
	ok, err := terminal.ValidatePod(pod, containerName)
	if !ok {
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

func serveWsLogs(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	namespace := pathParams["namespace"]
	podName := pathParams["pod"]
	containerName := pathParams["container"]
	tailLine, _ := utils.StringToInt64(r.URL.Query().Get("tail"))
	follow, _ := utils.StringToBool(r.URL.Query().Get("follow"))
	log.Printf("log pod: %s, container: %s, namespace: %s, tailLine: %d, follow: %v\n", podName, containerName, namespace, tailLine, follow)

	writer, err := kubeLog.NewWsLogger(w, r, nil)
	if err != nil {
		log.Printf("get writer failed: %v\n", err)
		return
	}
	defer func() {
		log.Println("close session.")
		writer.Close()
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
	ok, err := terminal.ValidatePod(pod, containerName)
	if !ok {
		msg := fmt.Sprintf("Validate pod error! err: %v", err)
		log.Println(msg)
		writer.Write([]byte(msg))
		writer.Close()
		return
	}

	opt := corev1.PodLogOptions{
		Container: containerName,
		Follow:    follow,
		TailLines: &tailLine,
	}

	err = client.PodBox.LogStreamLine(podName, namespace, &opt, writer)
	if err != nil {
		msg := fmt.Sprintf("log err: %v", err)
		log.Println(msg)
		writer.Write([]byte(msg))
		writer.Close()
	}
	return
}

func main() {
	router := mux.NewRouter()
	router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	// TODO
	// temporarily use relative path, run by `go run cmd/webshell/webshell_main.go` in project root path.
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./frontend/"))))
	// enter webshell by url like: http://127.0.0.1:8090/terminal?namespace=default&pod=nginx-65f9798fbf-jdrgl&container=nginx
	router.HandleFunc("/terminal", serveTerminal)
	router.HandleFunc("/ws/{namespace}/{pod}/{container}/webshell", serveWsTerminal)
	router.HandleFunc("/logs", serveLogs)
	router.HandleFunc("/ws/{namespace}/{pod}/{container}/logs", serveWsLogs)
	log.Fatal(http.ListenAndServe(*addr, router))
}
