package util

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"os"
)

// MD5FromFile calculate the hash value of the file
// If you reuse the file reader, please set its offset to start position first, like os.File.Seek
func MD5FromFile(file io.Reader, bufSize int) (hash string, err error) {
	if file == nil {
		err = errors.New("file is nil")
		return hash, err
	}
	block := make([]byte, bufSize)
	md5Provider := md5.New()
	reader := bufio.NewReader(file)
	for {
		n, err := reader.Read(block)
		if err == io.EOF {
			break
		}
		if err != nil {
			return hash, err
		}
		_, err = md5Provider.Write(block[:n])
		if err != nil {
			return hash, err
		}
	}
	sum := md5Provider.Sum(nil)
	hash = hex.EncodeToString(sum)
	return hash, nil
}

// MD5FromFileName calculate the hash value of the file
func MD5FromFileName(path string, bufSize int) (hash string, err error) {
	if len(path) == 0 {
		err = errors.New("file path can't be empty")
		return hash, err
	}
	f, err := os.Open(path)
	if err != nil {
		return hash, err
	}
	defer f.Close()
	return MD5FromFile(f, bufSize)
}

// MD5 calculate the hash value of the string
func MD5(s string) (hash string, err error) {
	md5Provider := md5.New()
	_, err = md5Provider.Write([]byte(s))
	if err != nil {
		return hash, err
	}
	sum := md5Provider.Sum(nil)
	hash = hex.EncodeToString(sum)
	return hash, nil
}
