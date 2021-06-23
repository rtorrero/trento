package internal

import (
	"crypto/md5"
	"encoding/hex"
	"hash/crc32"
	"io"
	"os"
)

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func Md5sum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func CRC32hash(input []byte) int {
	crc32Table := crc32.MakeTable(crc32.IEEE)
	return int(crc32.Checksum(input, crc32Table))
}
