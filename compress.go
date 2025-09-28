package createrepo

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/ulikunitz/xz"
)

func compress(data []byte, algo string) (compressed []byte, checksum *checksum, suffix string, err error) {
	switch algo {
	case "xz":
		compressed, err = xzCompress(data)
		suffix = ".xz"
	case "gz":
		compressed, err = gzCompress(data)
		suffix = ".gz"
	default:
		return nil, nil, "", fmt.Errorf("unsupported compress algo: %s", algo)
	}

	checksum = getChecksumOfBytes(compressed)
	return compressed, checksum, suffix, err
}

func xzCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	z, err := xz.NewWriter(&buf)
	if err != nil {
		return nil, err
	}

	if _, err := z.Write(data); err != nil {
		return nil, err
	}

	if err := z.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func gzCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	z := gzip.NewWriter(&buf)

	_, err := z.Write(data)
	if err != nil {
		return nil, err
	}

	if err := z.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
