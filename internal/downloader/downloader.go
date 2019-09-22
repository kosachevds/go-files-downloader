package downloader

import (
	"bytes"
	"errors"
	"io"
	"net/http"
)

func Download(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(response.Status)
	}
	var data bytes.Buffer
	_, err = io.Copy(&data, response.Body)
	if err != nil {
		return nil, err
	}
	return data.Bytes(), nil
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
