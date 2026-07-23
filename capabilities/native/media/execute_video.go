package media

import (
	"context"
	"strconv"

	"idun/capabilities"
)

func (c *Capability) executeVideo(ctx context.Context, req capabilities.CapabilityRequest, op MediaOperation) (map[string]interface{}, error) {
	switch op {
	case OperationPlayVideo:
		path := req.Parameters["path"]
		sessionID, err := c.provider.PlayVideo(ctx, path)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"session_id": sessionID,
			"status":     "playing",
		}, nil

	case OperationStopVideo:
		sessionID := req.Parameters["session_id"]
		err := c.provider.StopVideo(ctx, sessionID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"session_id": sessionID,
			"status":     "stopped",
		}, nil

	case OperationPauseVideo:
		sessionID := req.Parameters["session_id"]
		err := c.provider.PauseVideo(ctx, sessionID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"session_id": sessionID,
			"status":     "paused",
		}, nil

	case OperationResumeVideo:
		sessionID := req.Parameters["session_id"]
		err := c.provider.ResumeVideo(ctx, sessionID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"session_id": sessionID,
			"status":     "playing",
		}, nil

	case OperationRecordVideo:
		deviceID := req.Parameters["device_id"]
		destination := req.Parameters["destination"]
		durationMs := 0
		if val, ok := req.Parameters["duration_ms"]; ok {
			if d, err := strconv.Atoi(val); err == nil && d > 0 {
				durationMs = d
			}
		}

		err := c.provider.RecordVideo(ctx, deviceID, destination, durationMs)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"status":      "recorded",
			"destination": destination,
		}, nil
	}

	return nil, nil
}
