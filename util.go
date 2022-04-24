package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func Filter[Type any](values []Type, validator func(Type) bool) []Type {
	var result = make([]Type, 0)
	for _, value := range values {
		if validator(value) {
			result = append(result, value)
		}
	}
	return result
}

func reverseStringMap(map_ map[string]string) map[string]string {
	reversed := make(map[string]string, 0)
	for key, value := range map_ {
		reversed[value] = key
	}
	return reversed
}

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}

func gzipCompress(data []byte) []byte {
	buffer := bytes.NewBuffer(make([]byte, 0))
	writer := gzip.NewWriter(buffer)
	_, err := writer.Write(data)
	panicIf(err)

	err = writer.Close()
	panicIf(err)

	return buffer.Bytes()
}

func quakeTextToPlainText(value string) string {
	readableTextBytes := []byte(value)

	var charset = [...]byte{
		' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		'[', ']', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', ' ', ' ', ' ', ' ',
	}

	for i := range value {
		readableTextBytes[i] &= 0x7f

		if value[i] < byte(len(charset)) {
			readableTextBytes[i] = charset[value[i]]
		}
	}

	return strings.TrimSpace(string(readableTextBytes))
}

func stringToIntArray(value string) []int {
	intArr := make([]int, len(value))

	for i := range value {
		intArr[i] = int(value[i])
	}

	return intArr
}

func timeInFuture(delta int) time.Time {
	return time.Now().Add(time.Duration(delta) * time.Millisecond)
}

func Download(url string, dest string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
