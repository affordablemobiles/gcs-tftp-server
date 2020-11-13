package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"cloud.google.com/go/storage"
	tftp "github.com/a1comms/libgotftp/src"
)

var (
	TFTP_LISTEN_HOST     string = os.Getenv("TFTP_LISTEN_HOST")
	TFTP_REPLY_HOST      string = os.Getenv("TFTP_REPLY_HOST")
	TFTP_REPLY_PORT_LOW  string = os.Getenv("TFTP_REPLY_PORT_LOW")
	TFTP_REPLY_PORT_HIGH string = os.Getenv("TFTP_REPLY_PORT_HIGH")
	TFTP_ENABLE_HTTP     string = os.Getenv("TFTP_ENABLE_HTTP")
	HTTP_LISTEN_HOST     string = os.Getenv("HTTP_LISTEN_HOST")
	GCS_BUCKET           string = mustGetenv("GCS_BUCKET")

	storageBucket *storage.BucketHandle
)

func main() {
	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Storage Client Error: %s", err)
	}

	storageBucket = client.Bucket(GCS_BUCKET)

	if TFTP_ENABLE_HTTP == "true" {
		go func() {
			log.Printf("Starting HTTP endpoint on port " + HTTP_LISTEN_HOST + ":8080")
			log.Fatalf(
				http.ListenAndServe(HTTP_LISTEN_HOST+":8080", http.StripPrefix("/",
					http.HandlerFunc(httpHandleRequest),
				)).Error(),
			)
		}()
	}

	addr, err := net.ResolveUDPAddr("udp", TFTP_LISTEN_HOST+":69")
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %s", err)
	}

	replyAddr, err := net.ResolveUDPAddr("udp", TFTP_REPLY_HOST+":0")
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %s", err)
	}

	portLow, portHigh := 5000, 5004
	if TFTP_REPLY_PORT_LOW != "" && TFTP_REPLY_PORT_HIGH != "" {
		portLow, err = strconv.Atoi(TFTP_REPLY_PORT_LOW)
		if err != nil {
			log.Fatalf("Invalid TFTP_REPLY_PORT_LOW: %s", TFTP_REPLY_PORT_LOW)
		}

		portHigh, err = strconv.Atoi(TFTP_REPLY_PORT_HIGH)
		if err != nil {
			log.Fatalf("Invalid TFTP_REPLY_PORT_HIGH: %s", TFTP_REPLY_PORT_HIGH)
		}

		if portLow >= portHigh {
			log.Fatalf("TFTP_REPLY_PORT_LOW higher than TFTP_REPLY_PORT_HIGH")
		}
	}

	portArray := []int{}
	for i := portLow; i <= portHigh; i++ {
		portArray = append(portArray, i)
	}

	server, err := tftp.NewTFTPServer(addr, replyAddr, portArray)
	if err != nil {
		log.Fatalf("Failed to listen for TFTP endpoint: %s", err)
		return
	}

	log.Printf("Listening for TFTP endpoint on %s", addr.String())

	for {
		res, err := server.Accept()
		if err != nil {
			log.Printf("TFTP: Bad tftp request: %s", err)
			continue
		}

		go tftpHandleRRQ(res)
	}
}

func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("%s environment variable not set.", k)
	}
	return v
}
