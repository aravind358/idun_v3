package files

import (
	"context"
	"encoding/base64"
	"errors"
	"strconv"

	"idun/capabilities"
)

func (c *Capability) executeWrite(ctx context.Context, req capabilities.CapabilityRequest, op FileOperation) (map[string]interface{}, error) {
	path := req.Parameters["path"]

	// Destructive operations must be explicitly requested. The safety checks here
	// rely on the user/planner sending the correct parameters.
	
	switch op {
	case OperationCreateFile, OperationWriteFile, OperationAppendFile:
		// Decode base64 data if 'data_bytes' provided, otherwise use 'data_text'
		var data []byte
		if text, ok := req.Parameters["data_text"]; ok {
			data = []byte(text)
		} else if b64, ok := req.Parameters["data_bytes"]; ok {
			var err error
			data, err = base64.StdEncoding.DecodeString(b64)
			if err != nil {
				return nil, errors.New("invalid base64 encoding in 'data_bytes'")
			}
		}
		
		appendFlag := false
		if op == OperationAppendFile {
			appendFlag = true
		} else {
			if b, err := strconv.ParseBool(req.Parameters["append"]); err == nil {
				appendFlag = b
			}
		}

		err := c.provider.WriteFile(ctx, path, data, appendFlag)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"status": "success", "path": path}, nil

	case OperationCopyFile:
		dest := req.Parameters["destination"]
		err := c.provider.CopyFile(ctx, path, dest)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"status": "success", "src": path, "dest": dest}, nil

	case OperationMoveFile, OperationRenameFile:
		dest := req.Parameters["destination"]
		err := c.provider.MoveFile(ctx, path, dest)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"status": "success", "src": path, "dest": dest}, nil

	case OperationDeleteFile:
		err := c.provider.DeleteFile(ctx, path)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"status": "success", "path": path}, nil
		
	case OperationTemporaryFile:
		prefix := req.Parameters["prefix"]
		suffix := req.Parameters["suffix"]
		tempPath, err := c.provider.CreateTemporaryFile(ctx, prefix, suffix)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"status": "success", "path": tempPath}, nil
	}

	return nil, nil
}
