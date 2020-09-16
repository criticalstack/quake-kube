package http

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

func GetBody(url string) ([]byte, error) {
	client := http.Client{
		Timeout: 5 * time.Minute,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("cannot get url %q: %v", url, http.StatusText(resp.StatusCode))
	}

	return ioutil.ReadAll(resp.Body)
}

func GetUntil(url string, stop <-chan struct{}) error {
	client := http.Client{
		Timeout: 1 * time.Second,
	}
	for {
		select {
		case <-stop:
			return errors.Errorf("not available: %q", url)
		default:
			resp, err := client.Get(url)
			if err != nil {
				continue
			}
			resp.Body.Close()
			return nil
		}
	}
}
