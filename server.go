package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strconv"
)

var (
	address      = flag.String("address", "127.0.0.1", "IP address to listen on")
	port         = flag.Int("port", 10000, "HTTP port to listen on")
	clang_format = flag.String("clang-format", "clang-format",
		"Path to the clang-format executable to use")
	style = flag.String("style",
		"{BasedOnStyle: Google, DerivePointerBinding: false}",
		"Style specification passed to clang-format")
)

func formatHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s -- %s", r.Method, r.URL.Path, r.Proto, r.RemoteAddr)
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	cmd := exec.Command(*clang_format, "-style", *style)
	cmd.Stdin = r.Body
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Printf("clang-format failed: %v", err)
		if stderr.String() != "" {
			log.Printf("stderr: %s", stderr.String())
		}
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(200)
	_, err = io.Copy(w, bytes.NewReader(stdout.Bytes()))
	if err != nil {
		log.Printf("failed to write to response: %v", err)
		return
	}
}

func main() {
	flag.Parse()

	http.HandleFunc("/format", formatHandler)
	http.ListenAndServe(*address+":"+strconv.Itoa(*port), nil)
}
