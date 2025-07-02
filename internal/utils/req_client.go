package utils

import (
	"crypto/tls"
	"fmt"
	"github.com/go-resty/resty/v2"
	"time"
)

var (
	RestySSEClient = resty.New().
		SetTimeout(1 * time.Minute).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
		SetDoNotParseResponse(true).
		SetHeaders(map[string]string{
			"Content-Type": "application/json",
		}).
		OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
			if resp.StatusCode() != 200 {
				return fmt.Errorf("Jetbrains API error: status %d, body: %s",
					resp.StatusCode(), resp.String())
			}
			return nil
		})
)
