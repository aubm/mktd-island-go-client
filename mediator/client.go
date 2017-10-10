package mediator

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"mktd5/mktd-island/client/utils"

	"github.com/pkg/errors"
)

var (
	ErrGameFull             = errors.New("game is full")
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
)

type Client struct {
	AppConfig *utils.AppConfig `inject:""`
}

func (c *Client) GameState() (state State, err error) {
	if _, err := c.jsonRequest(http.MethodGet, "/map", nil, nil, &state); err != nil {
		return state, errors.Wrap(err, "failed to perform json request")
	}
	return
}

func (c *Client) Move(options MoveOptions) (res MoveResult, err error) {
	resp, err := c.jsonRequest(http.MethodPost, "/map", options, map[string]string{"uuid": options.ID}, nil)
	if err != nil {
		if errors.Cause(err) == ErrUnexpectedStatusCode && resp.StatusCode == http.StatusBadRequest {
			return res, nil
		}
		return res, errors.Wrap(err, "failed to perform json request")
	}
	res.Accepted = true
	return res, nil
}

func (c *Client) Register(options RegisterOptions) (res RegisterResult, err error) {
	resp, err := c.jsonRequest(http.MethodPost, "/player", options, nil, &res)
	if err != nil {
		if errors.Cause(err) == ErrUnexpectedStatusCode && resp.StatusCode == http.StatusLocked {
			return res, ErrGameFull
		}
		return res, errors.Wrap(err, "failed to perform json request")
	}

	return res, nil
}

func (c *Client) jsonRequest(method, url string, body interface{}, headers map[string]string, target interface{}) (*http.Response, error) {
	var payload io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		payload = bytes.NewReader(b)
	}

	resp, err := c.executeHttpRequest(method, url, payload, headers)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute http request, method = %s, url = %s", method, url)
	}

	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return resp, errors.Wrapf(ErrUnexpectedStatusCode, "got status code %v", resp.StatusCode)
	}

	if target == nil {
		return resp, nil
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return nil, errors.Wrap(err, "failed to decode http request json response")
	}

	return resp, nil
}

func (c *Client) executeHttpRequest(method, url string, body io.Reader, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, c.AppConfig.BaseMediatorURL+url, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build http request")
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to perform http request")
	}

	return resp, nil
}

type RegisterOptions struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
}

type RegisterResult struct {
	PlayerID int `json:"id"`
}

type MoveOptions struct {
	ID   string    `json:"-"`
	Move Direction `json:"move"`
}

type MoveResult struct {
	Accepted bool
}
