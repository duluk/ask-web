package download

import (
	"fmt"
	"io"
	"net/http"
)

type DownloadError struct {
	StatusCode int
	Message    string
}

func (e *DownloadError) Error() string {
	return fmt.Sprintf("%s (%d)", e.Message, e.StatusCode)
}

func Page(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", &DownloadError{
			StatusCode: resp.StatusCode,
			Message:    resp.Status,
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
