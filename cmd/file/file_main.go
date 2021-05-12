// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/gorilla/mux"

	_ "github.com/maoqide/kubeutil/initialize"
	"github.com/maoqide/kubeutil/pkg/copy"
	"github.com/maoqide/kubeutil/pkg/kube"
)

var (
	addr = flag.String("addr", ":8091", "http service address")
)

func serveFile(w http.ResponseWriter, r *http.Request) {
	// auth
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "./frontend/file.html")
}

func download(w http.ResponseWriter, r *http.Request) {

	pathParams := mux.Vars(r)
	namespace := pathParams["namespace"]
	podName := pathParams["pod"]
	containerName := pathParams["container"]
	file := r.URL.Query().Get("file")
	log.Printf("exec pod: %s, container: %s, namespace: %s, file: %s\n",
		podName, containerName, namespace, file)
	if len(file) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	client, err := kube.GetClient()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	cpOpt := copy.New(client, namespace, podName, containerName)
	reader, fileName, err := cpOpt.CopyFromPod(file)
	if err != nil {
		log.Printf("CopyFromPod error: %+v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.tar", fileName))
	io.Copy(w, reader)
	return
}

func main() {
	router := mux.NewRouter()
	router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	router.HandleFunc("/file", serveFile)
	// http://127.0.0.1:8091/copy/default/nginx-deployment-8d8d4dc86-sqfcx/nginx/download?file=/root/sss
	// curl http://127.0.0.1:8091/copy/default/nginx-deployment-8d8d4dc86-sqfcx/nginx/download\?file\=/root/sss -o xxx.tar
	router.HandleFunc("/copy/{namespace}/{pod}/{container}/download", download)
	log.Fatal(http.ListenAndServe(*addr, router))
}
