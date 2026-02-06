package checks

import (
	"fmt"
	"net/http"
	"time"

	"github.com/northstar-group-demo/democtl/internal/scenario"
)

// runHTTPGet executes an HTTP GET check with retry logic
func (r *Runner) runHTTPGet(check scenario.Check) CheckResult {
	result := CheckResult{
		Type:        check.Type,
		Description: check.Description,
	}

	// Extract parameters with defaults
	url := check.URL
	timeoutSeconds := check.TimeoutSeconds
	if timeoutSeconds == 0 {
		timeoutSeconds = 30
	}
	retryInterval := check.RetryInterval
	if retryInterval == 0 {
		retryInterval = 2
	}

	expectStatus := check.Expect.Status
	expectStatusNot := check.Expect.StatusNot

	r.logVerbose(fmt.Sprintf("GET %s (timeout: %ds, retry interval: %ds)", url, timeoutSeconds, retryInterval))

	// Create HTTP client with timeouts
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}

	startTime := time.Now()
	deadline := startTime.Add(time.Duration(timeoutSeconds) * time.Second)

	for {
		// Make request
		resp, err := client.Get(url)
		statusCode := 0
		if err == nil {
			statusCode = resp.StatusCode
			resp.Body.Close()
		}

		r.logVerbose(fmt.Sprintf("Response status: %d", statusCode))

		// Check negative assertion (status_not)
		if len(expectStatusNot) > 0 {
			for _, notStatus := range expectStatusNot {
				if statusCode == notStatus {
					remaining := time.Until(deadline).Seconds()
					if remaining <= 0 {
						result.Status = "fail"
						result.Message = fmt.Sprintf("got %d, expected not in %v", statusCode, expectStatusNot)
						return result
					}
					r.logVerbose(fmt.Sprintf("Check failed, retrying in %ds... (%.0fs remaining)", retryInterval, remaining))
					time.Sleep(time.Duration(retryInterval) * time.Second)
					continue
				}
			}
		}

		// Check positive assertion (status)
		if len(expectStatus) > 0 {
			found := false
			for _, expectedStatus := range expectStatus {
				if statusCode == expectedStatus {
					found = true
					break
				}
			}
			if !found {
				remaining := time.Until(deadline).Seconds()
				if remaining <= 0 {
					result.Status = "fail"
					result.Message = fmt.Sprintf("got %d, expected one of %v", statusCode, expectStatus)
					return result
				}
				r.logVerbose(fmt.Sprintf("Check failed, retrying in %ds... (%.0fs remaining)", retryInterval, remaining))
				time.Sleep(time.Duration(retryInterval) * time.Second)
				continue
			}
		}

		// Check passed
		result.Status = "pass"
		return result
	}
}
