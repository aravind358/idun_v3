package media

import "context"

// MediaProvider abstracts native multimedia operations.
// This isolates the capability from host APIs and enables seamless mock testing.
type MediaProvider interface {
	// Audio
	PlayAudio(ctx context.Context, path string) (string, error)
	StopAudio(ctx context.Context, sessionID string) error
	PauseAudio(ctx context.Context, sessionID string) error
	ResumeAudio(ctx context.Context, sessionID string) error
	RecordAudio(ctx context.Context, deviceID, destination string, durationMs int) error

	// Video
	PlayVideo(ctx context.Context, path string) (string, error)
	PauseVideo(ctx context.Context, sessionID string) error
	ResumeVideo(ctx context.Context, sessionID string) error
	StopVideo(ctx context.Context, sessionID string) error
	RecordVideo(ctx context.Context, deviceID, destination string, durationMs int) error

	// Image
	CaptureImage(ctx context.Context, deviceID, destination string) error
	LoadImage(ctx context.Context, path string) (map[string]interface{}, error)
	SaveImage(ctx context.Context, destination string, data []byte) error

	// Metadata
	GetMetadata(ctx context.Context, path string) (map[string]interface{}, error)

	// Devices
	ListMediaDevices(ctx context.Context, deviceType string) ([]map[string]interface{}, error)
	GetDevice(ctx context.Context, deviceID string) (map[string]interface{}, error)
	ListCodecs(ctx context.Context) ([]string, error)
}
