package createrepo

import (
	"fmt"
	"github.com/pkg/xattr"
	"strings"
)

func getXattrChecksum(name string) (*checksum, bool) {
	data, err := xattr.Get(name, "user.repo.checksum")
	if err != nil {
		return nil, false
	}
	fields := strings.Fields(string(data))
	if len(fields) != 2 {
		return nil, false
	}

	switch fields[0] {
	case "sha256":
		return &checksum{Type: "sha256", PkgID: "YES", Data: fields[1]}, true
	}

	return nil, false
}

func setXattrChecksum(name string, checksum *checksum) error {
	if err := xattr.Set(name, "user.repo.checksum", []byte(fmt.Sprintf("%s %s", checksum.Type, checksum.Data))); err != nil {
		return err
	}

	return nil
}
