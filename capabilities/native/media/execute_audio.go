package media

import (
	"context"
	"strconv"

	"idun/capabilities"
)

func (c *Capability) executeAudio(ctx context.Context, req capabilities.CapabilityRequest, op MediaOperation) (map[string]interface{}, error) {
	switch op {
	case OperationPlayAudio:
		path := req.Parameters["path"]
		sessionID, err := c.provider.PlayAudio(ctx, path)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"session_id": sessionID,
			"status":     "playing",
		}, nil

	case OperationStopAudio:
		sessionID := req.Parameters["session_id"]
		err := c.provider.StopAudio(ctx, sessionID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"session_id": sessionID,
			"status":     "stopped",
		}, nil

	case OperationPauseAudio:
		sessionID := req.Parameters["session_id"]
		err := c.provider.PauseAudio(ctx, sessionID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"session_id": sessionID,
			"status":     "paused",
		}, nil

	case OperationResumeAudio:
		sessionID := req.Parameters["session_id"]
		err := c.provider.ResumeAudio(ctx, sessionID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"session_id": sessionID,
			"status":     "playing",
		}, nil

	case OperationRecordAudio:
		deviceID := req.Parameters["device_id"]
		destination := req.Parameters["destination"]
		durationMs := 0
		if val, ok := req.Parameters["duration_ms"]; ok {
			if d, err := strconv.Atoi(val); err == nil && d > 0 {
				durationMs = d
			}
		}

		err := c.provider.RecordAudio(ctx, deviceID, destination, durationMs)
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
