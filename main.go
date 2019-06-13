package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
	"github.com/tftp-go-team/libgotftp/src"
	"google.golang.org/api/option"
)

var (
	TFTP_UDP_HOST		 string = os.Getenv("TFTP_UDP_HOST")
	TFTP_ENABLE_HTTP     string = os.Getenv("TFTP_ENABLE_HTTP")
	GCS_CREDENTIALS_FILE string = mustGetenv("GCS_CREDENTIALS_FILE")
	GCS_BUCKET           string = mustGetenv("GCS_BUCKET")

	storageBucket *storage.BucketHandle
)

func main() {
	ctx := context.Background()

	client, err := storage.NewClient(ctx, option.WithCredentialsFile(GCS_CREDENTIALS_FILE))
	if err != nil {
		log.Fatalf("Storage Client Error: %s", err)
	}

	storageBucket = client.Bucket(GCS_BUCKET)

	if TFTP_ENABLE_HTTP == "true" {
		go func() {
			log.Printf("Starting HTTP endpoint on port 8080")
			log.Fatalf(
				http.ListenAndServe(":8080", http.StripPrefix("/",
					http.HandlerFunc(httpHandleRequest),
				)).Error(),
			)
		}()
	}

	addr, err := net.ResolveUDPAddr("udp", TFTP_UDP_HOST+":69")
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %s", err)
	}

	server, err := tftp.NewTFTPServer(addr)
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
