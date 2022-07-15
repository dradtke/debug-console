package dap

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func getZip(dst, src string) error {
	log.Printf("Downloading zip archive from %s to %s", src, dst)
	resp, err := http.Get(src)
	if err != nil {
		return fmt.Errorf("getZip: %w", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("getZip: %w", err)
	}

	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return fmt.Errorf("getZip: %w", err)
	}

	// TODO: need to save these files to dapDir/name, preserving their structure
	for _, f := range r.File {
		r, err := f.Open()
		if err != nil {
			return fmt.Errorf("getZip: %w", err)
		}
		fileDest := filepath.Join(dst, f.Name)
		if err := os.MkdirAll(filepath.Dir(fileDest), 0755); err != nil {
			return fmt.Errorf("getZip: %w", err)
		}
		f, err := os.OpenFile(fileDest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("getZip: %w", err)
		}
		if _, err := io.Copy(f, r); err != nil {
			return fmt.Errorf("getZip: %w", err)
		}
	}
	log.Print("Download complete")

	return nil
}
