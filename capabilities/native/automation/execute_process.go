package automation

import (
	"context"
	"strconv"

	"idun/capabilities"
)

func (c *Capability) executeProcess(ctx context.Context, req capabilities.CapabilityRequest, op AutomationOperation) (map[string]interface{}, error) {
	switch op {
	case OperationListProcesses:
		processes, err := c.provider.ListProcesses(ctx)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"processes": processes}, nil

	case OperationGetProcess:
		pid, _ := strconv.Atoi(req.Parameters["pid"])
		process, err := c.provider.GetProcess(ctx, pid)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"process": process}, nil
	}
	return nil, nil
}
