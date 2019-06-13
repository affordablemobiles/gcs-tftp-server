package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

func httpHandleRequest(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	status := http.StatusOK
	size := 0
	defer func(){
		httpLogRequest(startTime, status, size, r)
	}()

	fileReader, fileSize, err := StorageGetFile(r.URL.Path)
	if err != nil {
		status = http.StatusNotFound
		http.Error(w, "File Not Found", status)
		return
	}
	defer fileReader.Close()

	size = int(fileSize)

	w.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))

	w.WriteHeader(status)

	if r.Method != "HEAD" {
		io.CopyN(w, fileReader, fileSize)
	}
}

func httpLogRequest(ts time.Time, status int, size int, req *http.Request) {
	username := "-"
	if req.URL.User != nil {
		if name := req.URL.User.Username(); name != "" {
			username = name
		}
	}

	host, _, err := net.SplitHostPort(req.RemoteAddr)

	if err != nil {
		host = req.RemoteAddr
	}

	uri := req.RequestURI

	// Requests using the CONNECT method over HTTP/2.0 must use
	// the authority field (aka r.Host) to identify the target.
	// Refer: https://httpwg.github.io/specs/rfc7540.html#CONNECT
	if req.ProtoMajor == 2 && req.Method == "CONNECT" {
		uri = req.Host
	}
	if uri == "" {
		uri = req.URL.RequestURI()
	}

	buf := make([]byte, 0)
	buf = append(buf, host...)
	buf = append(buf, " - "...)
	buf = append(buf, username...)
	buf = append(buf, " ["...)
	buf = append(buf, ts.Format("02/Jan/2006:15:04:05 -0700")...)
	buf = append(buf, `] "`...)
	buf = append(buf, req.Method...)
	buf = append(buf, " "...)
	buf = appendQuoted(buf, uri)
	buf = append(buf, " "...)
	buf = append(buf, req.Proto...)
	buf = append(buf, `" `...)
	buf = append(buf, strconv.Itoa(status)...)
	buf = append(buf, " "...)
	buf = append(buf, strconv.Itoa(size)...)

	log.Printf("%s\n", string(buf))
}
