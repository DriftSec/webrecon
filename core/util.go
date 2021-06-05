package core

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func ErrStr(e error) string {
	return fmt.Sprintf("%s", e)

}

func SliceContains(lst []string, word string) bool {
	for _, v := range lst {
		if v == word {
			return true
		}
	}
	return false
}

// CleanSlice takes a slice in, removes emtpy or null items, and returns a clean slice
func CleanSlice(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" && str != "null" {
			r = append(r, str)
		}
	}
	return r
}

// UniqueSlice returns a string slice with duplicates removed
func UniqueSlice(strSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range strSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return CleanSlice(list)
}

func WriteSliceToFile(slice []string, pathtofile string) error {
	dir := filepath.Dir(pathtofile)
	err := MakeDir(dir)
	if err != nil {
		return err
	}
	var sz int
	_, err = os.Stat(pathtofile)
	if !os.IsNotExist(err) {
		return errors.New("file exsists")
	}

	f, err := os.Create(pathtofile)

	if err != nil {
		return err
	}

	defer f.Close()

	for _, line := range slice {
		outstr := fmt.Sprintf("%s\n", line)
		sz += len(outstr)
		_, err2 := f.WriteString(outstr)
		if err2 != nil {
			return err2
		}
	}
	return nil
}

func MakeDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

// // CheckIPAddress validates an IPv4 or IPv6 address. returns true if valid
// func CheckIPAddress(ip string) bool {
// 	return net.ParseIP(ip) != nil
// }

// func IsIPv4(address string) bool {
// 	return netaddr.MustNewIPAddress(address).IsIPv4()
// }

// func IsIPv6(address string) bool {
// 	return netaddr.MustNewIPAddress(address).IsIPv6()
// }

// ReadLines reads a file and returns a slice of lines
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func RandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
