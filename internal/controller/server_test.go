package controller_test

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func ServeDir(t *testing.T, path string) {
	http.HandleFunc("/file.tar.gz", func(writer http.ResponseWriter, request *http.Request) {
		gw := gzip.NewWriter(writer)
		defer gw.Close()
		tw := tar.NewWriter(gw)
		defer tw.Close()

		_ = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			th, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			fh, err := os.Open(path)
			if err != nil {
				return err
			}
			defer fh.Close()
			if err = tw.WriteHeader(th); err != nil {
				return err
			}
			_, err = io.Copy(tw, fh)
			return err
		})
	})

	log.Println("Starting server....")

	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		t.Errorf("unable to bind %v", err)
	}

	// #nosec
	err = http.Serve(listener, nil)
	if err != nil {
		t.Errorf("unable to serve %v", err)
	}
}
