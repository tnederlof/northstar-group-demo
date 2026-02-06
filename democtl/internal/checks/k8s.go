package checks

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/northstar-group-demo/democtl/internal/scenario"
)

// runK8sJQEquals checks a K8s resource field with jq
func (r *Runner) runK8sJQEquals(check scenario.Check) CheckResult {
	result := CheckResult{
		Type:        check.Type,
		Description: check.Description,
	}

	if check.Resource == nil {
		result.Status = "fail"
		result.Message = "missing resource field"
		return result
	}

	kind := check.Resource.Kind
	name := check.Resource.Name
	jqExpr := check.JQ
	expected := check.Equals
	namespace := r.getNamespace()

	r.logVerbose(fmt.Sprintf("Checking %s/%s with jq: %s", kind, name, jqExpr))

	// Get resource as JSON
	cmd := exec.Command("kubectl", "--context="+r.opts.KubeContext, "-n", namespace, "get", kind, name, "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		result.Status = "fail"
		result.Message = fmt.Sprintf("failed to get resource: %v", err)
		return result
	}

	// Apply jq expression
	jqCmd := exec.Command("jq", "-r", jqExpr)
	jqCmd.Stdin = strings.NewReader(string(output))
	jqOutput, err := jqCmd.Output()
	if err != nil {
		result.Status = "fail"
		result.Message = fmt.Sprintf("jq failed: %v", err)
		return result
	}

	actual := strings.TrimSpace(string(jqOutput))
	if actual != expected {
		result.Status = "fail"
		result.Message = fmt.Sprintf("expected '%s', got '%s'", expected, actual)
		return result
	}

	result.Status = "pass"
	return result
}

// runK8sPodsContainLog checks pod logs for a substring
func (r *Runner) runK8sPodsContainLog(check scenario.Check) CheckResult {
	result := CheckResult{
		Type:        check.Type,
		Description: check.Description,
	}

	selector := check.Selector
	contains := check.Contains
	sinceSeconds := check.SinceSeconds
	if sinceSeconds == 0 {
		sinceSeconds = 300
	}
	namespace := r.getNamespace()

	r.logVerbose(fmt.Sprintf("Checking logs for pods with selector: %s", selector))

	// Get logs
	cmd := exec.Command("kubectl", "--context="+r.opts.KubeContext, "-n", namespace, "logs",
		"-l", selector, "--since="+strconv.Itoa(sinceSeconds)+"s", "--tail=500")
	output, err := cmd.Output()
	if err != nil {
		result.Status = "fail"
		result.Message = "no logs found"
		return result
	}

	logs := string(output)
	if logs == "" {
		result.Status = "fail"
		result.Message = "no logs found"
		return result
	}

	if !strings.Contains(logs, contains) {
		result.Status = "fail"
		result.Message = fmt.Sprintf("substring '%s' not found in logs", contains)
		return result
	}

	result.Status = "pass"
	return result
}

// runK8sPodTerminationReason checks pod termination reason
func (r *Runner) runK8sPodTerminationReason(check scenario.Check) CheckResult {
	result := CheckResult{
		Type:        check.Type,
		Description: check.Description,
	}

	selector := check.Selector
	expectedReason := check.Reason
	namespace := r.getNamespace()

	r.logVerbose(fmt.Sprintf("Checking termination reason for pods with selector: %s", selector))

	// Get pods as JSON
	cmd := exec.Command("kubectl", "--context="+r.opts.KubeContext, "-n", namespace, "get", "pods",
		"-l", selector, "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		result.Status = "fail"
		result.Message = fmt.Sprintf("failed to get pods: %v", err)
		return result
	}

	// Parse JSON to find termination reason
	var podList struct {
		Items []struct {
			Status struct {
				ContainerStatuses []struct {
					LastState struct {
						Terminated *struct {
							Reason string `json:"reason"`
						} `json:"terminated"`
					} `json:"lastState"`
				} `json:"containerStatuses"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.Unmarshal(output, &podList); err != nil {
		result.Status = "fail"
		result.Message = fmt.Sprintf("failed to parse pod JSON: %v", err)
		return result
	}

	foundReason := ""
	for _, pod := range podList.Items {
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.LastState.Terminated != nil && cs.LastState.Terminated.Reason != "" {
				foundReason = cs.LastState.Terminated.Reason
				break
			}
		}
		if foundReason != "" {
			break
		}
	}

	if foundReason != expectedReason {
		result.Status = "fail"
		if foundReason == "" {
			result.Message = fmt.Sprintf("expected '%s', got 'none'", expectedReason)
		} else {
			result.Message = fmt.Sprintf("expected '%s', got '%s'", expectedReason, foundReason)
		}
		return result
	}

	result.Status = "pass"
	return result
}

// runK8sPodRestartCount checks pod restart count
func (r *Runner) runK8sPodRestartCount(check scenario.Check) CheckResult {
	result := CheckResult{
		Type:        check.Type,
		Description: check.Description,
	}

	selector := check.Selector
	minRestarts := check.MinRestarts
	if minRestarts == 0 {
		minRestarts = 1
	}
	namespace := r.getNamespace()

	r.logVerbose(fmt.Sprintf("Checking restart count for pods with selector: %s", selector))

	// Get restart counts using jsonpath
	cmd := exec.Command("kubectl", "--context="+r.opts.KubeContext, "-n", namespace, "get", "pods",
		"-l", selector, "-o", "jsonpath={.items[*].status.containerStatuses[*].restartCount}")
	output, err := cmd.Output()
	if err != nil {
		result.Status = "fail"
		result.Message = fmt.Sprintf("failed to get pods: %v", err)
		return result
	}

	countsStr := strings.TrimSpace(string(output))
	if countsStr == "" {
		result.Status = "fail"
		result.Message = fmt.Sprintf("restart count 0 < %d", minRestarts)
		return result
	}

	// Parse restart counts
	maxRestarts := 0
	for _, countStr := range strings.Fields(countsStr) {
		count, err := strconv.Atoi(countStr)
		if err == nil && count > maxRestarts {
			maxRestarts = count
		}
	}

	if maxRestarts < minRestarts {
		result.Status = "fail"
		result.Message = fmt.Sprintf("restart count %d < %d", maxRestarts, minRestarts)
		return result
	}

	result.Status = "pass"
	return result
}

// runK8sDeploymentAvailable checks if a deployment is available
func (r *Runner) runK8sDeploymentAvailable(check scenario.Check) CheckResult {
	result := CheckResult{
		Type:        check.Type,
		Description: check.Description,
	}

	name := check.Name
	timeoutSeconds := check.TimeoutSeconds
	if timeoutSeconds == 0 {
		timeoutSeconds = 60
	}
	waitSeconds := check.WaitSeconds
	namespace := r.getNamespace()

	r.logVerbose(fmt.Sprintf("Checking deployment: %s (timeout: %ds, wait: %ds)", name, timeoutSeconds, waitSeconds))

	// If wait_seconds is specified, wait before checking
	if waitSeconds > 0 {
		r.logVerbose(fmt.Sprintf("Waiting %ds for deployment to stabilize...", waitSeconds))
		time.Sleep(time.Duration(waitSeconds) * time.Second)
	}

	// Poll deployment status
	deadline := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)
	for {
		cmd := exec.Command("kubectl", "--context="+r.opts.KubeContext, "-n", namespace, "get", "deployment", name,
			"-o", "jsonpath={.status.conditions[?(@.type==\"Available\")].status}")
		output, err := cmd.Output()
		
		available := strings.TrimSpace(string(output))
		if err == nil && available == "True" {
			result.Status = "pass"
			return result
		}

		if time.Now().After(deadline) {
			result.Status = "fail"
			result.Message = fmt.Sprintf("deployment not available after %ds", timeoutSeconds)
			return result
		}

		remaining := time.Until(deadline).Seconds()
		r.logVerbose(fmt.Sprintf("Deployment not ready yet, waiting... (%.0fs remaining)", remaining))
		time.Sleep(2 * time.Second)
	}
}

// runK8sResourceExists checks if a K8s resource exists
func (r *Runner) runK8sResourceExists(check scenario.Check) CheckResult {
	result := CheckResult{
		Type:        check.Type,
		Description: check.Description,
	}

	if check.Resource == nil {
		result.Status = "fail"
		result.Message = "missing resource field"
		return result
	}

	kind := check.Resource.Kind
	name := check.Resource.Name
	namespace := r.getNamespace()

	r.logVerbose(fmt.Sprintf("Checking if %s/%s exists", kind, name))

	cmd := exec.Command("kubectl", "--context="+r.opts.KubeContext, "-n", namespace, "get", kind, name)
	if err := cmd.Run(); err != nil {
		result.Status = "fail"
		result.Message = fmt.Sprintf("%s/%s not found", kind, name)
		return result
	}

	result.Status = "pass"
	return result
}

// runK8sServiceMissingPort checks if a service is missing a port
func (r *Runner) runK8sServiceMissingPort(check scenario.Check) CheckResult {
	result := CheckResult{
		Type:        check.Type,
		Description: check.Description,
	}

	name := check.Name
	portName := check.PortName
	namespace := r.getNamespace()

	r.logVerbose(fmt.Sprintf("Checking if service %s is missing port %s", name, portName))

	// Get service port names
	cmd := exec.Command("kubectl", "--context="+r.opts.KubeContext, "-n", namespace, "get", "service", name,
		"-o", "jsonpath={.spec.ports[*].name}")
	output, err := cmd.Output()
	if err != nil {
		result.Status = "fail"
		result.Message = fmt.Sprintf("failed to get service: %v", err)
		return result
	}

	ports := strings.TrimSpace(string(output))
	portsList := strings.Fields(ports)
	
	// Check if port exists
	for _, p := range portsList {
		if p == portName {
			result.Status = "fail"
			result.Message = fmt.Sprintf("port %s exists, expected missing", portName)
			return result
		}
	}

	result.Status = "pass"
	return result
}
