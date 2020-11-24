package extension

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type SubscribeLogsParams struct {
	// ListenPort is the port that our log server listens on.
	ListenPort uint32
	// MaxItems is the maximum number of events to be buffered in memory. (default: 10000, minimum: 1000, maximum: 10000)
	MaxItems uint32
	// MaxBytes is the maximum size in bytes of the logs to be buffered in memory. (default: 262144, minimum: 262144, maximum: 1048576)
	MaxBytes uint32
	// TimeoutMS is the maximum time (in milliseconds) for a batch to be buffered. (default: 1000, minimum: 100, maximum: 30000)
	TimeoutMS uint32
}

// SubscribeResponse is the response body that is received from Logs API on subscribe
type SubscribeResponse struct {
	body string
}

// LogType represents the type of logs subscribed from Logs API
type LogType string

const (
	// Platform is to receive logs emitted by the platform
	Platform LogType = "platform"
	// Function is to receive logs emitted by the function
	Function LogType = "function"
	// Extension is to receive logs emitted by the extension
	Extension LogType = "extension"
)

const (
	logsURL = "/2020-08-15/logs"
)

// SubscribeLogs calls the Logs API to subscribe for the log events.
func (e *Client) SubscribeLogs(ctx context.Context, types []LogType, params SubscribeLogsParams) (res SubscribeResponse, err error) {
	url := e.baseURL + logsURL

	reqBody, err := json.Marshal(map[string]interface{}{
		"destination": map[string]interface{}{
			"protocol": "HTTP",
			"URI":      fmt.Sprintf("http://sandbox:%v", params.ListenPort),
			"encoding": "JSON",
			"method":   "POST",
		},
		"types": types,
		"buffering": map[string]interface{}{
			"timeoutMs": params.TimeoutMS,
			"maxBytes":  params.MaxBytes,
			"maxItems":  params.MaxItems,
		},
	})
	if err != nil {
		return res, err
	}

	// Create a HTTP Request with Context.
	httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return res, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set(extensionIdentifierHeader, e.ExtensionID)

	// Make the request
	httpRes, err := e.httpClient.Do(httpReq)
	if err != nil {
		return res, err
	}
	defer httpRes.Body.Close()
	body, err := ioutil.ReadAll(httpRes.Body)
	if err != nil {
		return res, err
	}
	if httpRes.StatusCode != http.StatusOK {
		return res, fmt.Errorf("extension: SubscribeLogs failed, status: %s, response: %s", httpRes.Status, string(body))
	}

	res.body = string(body)
	return res, nil
}
