package helper

import (
	"log"
	"net/url"
	"regexp"

	"github.com/go-resty/resty/v2"
	"github.com/linode/linodego"
)

// Global502Retry retries on intermittent 5xx errors on arbitrary endpoints from the API
func Global502Retry(response *resty.Response, _ error) bool {
	return response.StatusCode() == 502
}

func LinodeInstance500Retry() func(response *resty.Response, err error) bool {
	linodeGetRegex, err := regexp.Compile("linode/instances/[0-9]+/ips+")
	if err != nil {
		log.Fatal(err)
	}
	return GenericRetryCondition(500, linodeGetRegex)
}

// ImageUpload500Retry for [500] error when uploading an image
func ImageUpload500Retry() func(response *resty.Response, err error) bool {
	ImageUpload, err := regexp.Compile("images/upload")
	if err != nil {
		log.Fatal(err)
	}
	return GenericRetryCondition(500, ImageUpload)
}

func GenericRetryCondition(statusCode int, pathPattern *regexp.Regexp) func(response *resty.Response, err error) bool {
	return func(response *resty.Response, _ error) bool {
		if response.StatusCode() != statusCode || response.Request == nil {
			return false
		}

		requestURL, err := url.ParseRequestURI(response.Request.URL)
		if err != nil {
			log.Printf("[WARN] failed to parse request URL: %s", err)
			return false
		}

		// Check whether the string matches
		return pathPattern.MatchString(requestURL.Path)
	}
}

func ApplyAllRetryConditions(client *linodego.Client) {
	client.AddRetryCondition(Global502Retry)
	client.AddRetryCondition(LinodeInstance500Retry())
	client.AddRetryCondition(ImageUpload500Retry())
}
