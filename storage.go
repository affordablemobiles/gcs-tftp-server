package main

import (
	"context"
	"io"
	"log"
)

func StorageGetFile(path string) (io.ReadCloser, int64, error) {
	object := storageBucket.Object(path)

	log.Printf("GCS: Reading file %s", path)

	ctx := context.Background()

	attrs, err := object.Attrs(ctx)
	if err != nil {
		log.Printf("GCS: Error Reading File Attributes %s: %s", path, err)
		return nil, 0, err
	}

	reader, err := object.NewReader(ctx)
	if err != nil {
		log.Printf("GCS: Error Reading File %s: %s", path, err)
		return nil, 0, err
	}

	return reader, attrs.Size, nil
}
