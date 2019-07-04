package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/tftp-go-team/libgotftp/src"
)

func tftpHandleRRQ(res *tftp.RRQresponse) {
	started := time.Now()
	status, size := http.StatusNotFound, 0
	defer func() {
		tftpLogRequest(started, status, size, res)
	}()

	path := res.Request.Path
	if len(path) >= 1 {
		if path[0:1] == "/" {
			path = path[1:]
		}
	} else {
		status = http.StatusNotFound
		res.WriteError(tftp.NOT_FOUND, "invalid path")
		return
	}

	log.Printf(fmt.Sprintf(
		"%s: blocksize %d from %s",
		tftpLogRequestPrefix(res),
		res.Request.Blocksize,
		*res.Request.Addr,
	))

	defer res.End()

	fileReader, fileSize, err := StorageGetFile(path)
	if err != nil {
		res.WriteError(tftp.NOT_FOUND, "storage error")
		return
	}
	defer fileReader.Close()

	size = int(fileSize)
	if res.Request.TransferSize != -1 {
		res.TransferSize = int(fileSize)
	}

	if err := res.WriteOACK(); err != nil {
		status = http.StatusInternalServerError
		log.Printf("%s: Failed to write OACK: %s", tftpLogRequestPrefix(res), err)
		return
	}

	b := make([]byte, res.Request.Blocksize)

	totalBytes := 0

	for {
		bytesRead, err := fileReader.Read(b)
		totalBytes += bytesRead

		if err == io.EOF {
			if _, err := res.Write(b[:bytesRead]); err != nil {
				status = http.StatusInternalServerError
				log.Printf("%s: Failed to write last bytes of the reader: %s", tftpLogRequestPrefix(res), err)
				return
			}
			break
		} else if err != nil {
			status = http.StatusInternalServerError
			log.Printf("%s: Error while reading: %s", tftpLogRequestPrefix(res), err)
			res.WriteError(tftp.UNKNOWN_ERROR, err.Error())
			return
		}

		if _, err := res.Write(b[:bytesRead]); err != nil {
			status = http.StatusInternalServerError
			log.Printf("%s: Failed to write bytes: %s", tftpLogRequestPrefix(res), err)
			return
		}
	}

	took := time.Since(started)

	speed := float64(totalBytes) / took.Seconds() / 1024 / 1024

	status = http.StatusOK

	log.Printf("%s: Sent %v bytes in %v %f MB/s\n", tftpLogRequestPrefix(res), totalBytes, took, speed)
}

func tftpLogRequestPrefix(res *tftp.RRQresponse) string {
	req := res.Request

	host, _, err := net.SplitHostPort((*req.Addr).String())

	if err != nil {
		host = (*req.Addr).String()
	}

	return fmt.Sprintf("TFTP: GET %s from %s", req.Path, host)
}

func tftpLogRequest(ts time.Time, status int, size int, res *tftp.RRQresponse) {
	req := res.Request

	username := "-"

	host, _, err := net.SplitHostPort((*req.Addr).String())

	if err != nil {
		host = (*req.Addr).String()
	}

	uri := req.Path

	buf := make([]byte, 0)
	buf = append(buf, host...)
	buf = append(buf, " - "...)
	buf = append(buf, username...)
	buf = append(buf, " ["...)
	buf = append(buf, ts.Format("02/Jan/2006:15:04:05 -0700")...)
	buf = append(buf, `] "`...)
	buf = append(buf, `GET`...)
	buf = append(buf, " "...)
	buf = appendQuoted(buf, uri)
	buf = append(buf, " "...)
	buf = append(buf, `TFTP/0.0`...)
	buf = append(buf, `" `...)
	buf = append(buf, strconv.Itoa(status)...)
	buf = append(buf, " "...)
	buf = append(buf, strconv.Itoa(size)...)

	log.Printf("%s\n", string(buf))
}
