package downloader

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func DownloadFile(url, filepath string) error {
	file, err := os.OpenFile(filepath, os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	err = download(url, file)
	return err
}

func Download(url string) ([]byte, error) {
	var data bytes.Buffer
	err := download(url, &data)
	return data.Bytes(), err
}

func DownloadSimultaneously(urls []string) ([][]byte, error) {
	done := make(chan []byte, len(urls))
	errch := make(chan error, len(urls))
	for _, url := range urls {
		go func(url string) {
			bytes, err := Download(url)
			if err != nil {
				errch <- err
				done <- nil
				return
			}
			done <- bytes
			errch <- nil
		}(url)
	}
	// TODO: with builder
	var allErrors string
	results := make([][]byte, 0, len(urls))
	for i := 0; i < len(urls); i++ {
		results = append(results, <-done)
		if err := <-errch; err != nil {
			// allErrors = fmt.Sprintf("%v %v", allErrors, err)
			allErrors += " " + err.Error()
		}
	}
	var err error
	if allErrors != "" {
		err = errors.New(allErrors)
	}
	return results, err
}

func DownloadFilesSimultaneous(infos []FileInfo) error {
	count := len(infos)
	errch := make(chan error, count)
	for _, fi := range infos {
		go func(url, filepath string) {
			err := DownloadFile(url, filepath)
			if err != nil {
				err = fmt.Errorf("%v downloading failed with: %v", filepath, err)
			}
			errch <- err
		}(fi.URL, fi.Filename)
	}
	var builder strings.Builder
	for i := 0; i < count; i++ {
		if err := <-errch; err != nil {
			builder.WriteString(fmt.Sprintf("%v; ", err))
		}
	}
	if builder.Len() == 0 {
		return nil
	}
	return errors.New(builder.String())
}

func DownloadFilesLimitedSimultaneous(infos []FileInfo, maxSimultaneous int) error {
	if maxSimultaneous > len(infos) {
		maxSimultaneous = len(infos)
	}
	errch := make(chan error, maxSimultaneous)
	var builder strings.Builder
	for i, fi := range infos {
		if i >= maxSimultaneous {
			appendErrorIfNotNil(&builder, <-errch)
		}
		go downloadFileWithChan(fi.URL, fi.Filename, errch)
	}
	for i := 0; i < maxSimultaneous; i++ {
		appendErrorIfNotNil(&builder, <-errch)
	}
	if builder.Len() == 0 {
		return nil
	}
	return errors.New(builder.String())
}

func downloadFileWithChan(url, filepath string, errch chan<- error) {
	err := DownloadFile(url, filepath)
	if err != nil {
		err = fmt.Errorf("%v downloading failed with: %v", filepath, err)
	}
	errch <- err
}

func appendErrorIfNotNil(builder *strings.Builder, err error) {
	if err == nil {
		return
	}
	builder.WriteString(fmt.Sprintf("%v;\n", err))
}

func download(url string, result io.Writer) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return errors.New(response.Status)
	}
	_, err = io.Copy(result, response.Body)
	return err
}
