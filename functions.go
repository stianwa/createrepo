package createrepo

import (
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

const (
	xmlHeader = `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
)

func xmlencode(a any) ([]byte, error) {
	b, err := xml.MarshalIndent(a, "", "  ")
	if err != nil {
		return nil, err
	}
	b = append(b, '\n')

	return append([]byte(xmlHeader), b...), nil
}

func writeFile(name string, data []byte) (uint64, error) {
	tmpFile := name + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0666); err != nil {
		return 0, err
	}

	if err := os.Rename(tmpFile, name); err != nil {
		os.Remove(tmpFile)
		return 0, err
	}

	fi, err := os.Stat(name)
	if err != nil {
		return 0, err
	}

	return uint64(fi.ModTime().Unix()), nil
}

func getChecksumOfFile(name string) (*checksum, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}

	return &checksum{
		Type:  "sha256",
		PkgID: "YES",
		Data:  fmt.Sprintf("%x", h.Sum(nil)),
	}, nil
}

func getChecksumOfBytes(data []byte) *checksum {
	h := sha256.New()

	h.Write(data)
	
	return &checksum{
		Type:  "sha256",
		PkgID: "YES",
		Data:  fmt.Sprintf("%x", h.Sum(nil)),
	}
}
