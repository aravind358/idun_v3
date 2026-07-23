package media

import (
	"context"
	"errors"
)

type MockProvider struct {
	ShouldFail bool
}

func NewMockProvider(shouldFail bool) *MockProvider {
	return &MockProvider{ShouldFail: shouldFail}
}

func (p *MockProvider) PlayAudio(ctx context.Context, path string) (string, error) {
	if p.ShouldFail {
		return "", errors.New("mock audio play failure")
	}
	return "audio-session-1", nil
}

func (p *MockProvider) StopAudio(ctx context.Context, sessionID string) error {
	if p.ShouldFail {
		return errors.New("mock audio stop failure")
	}
	return nil
}

func (p *MockProvider) PauseAudio(ctx context.Context, sessionID string) error {
	if p.ShouldFail {
		return errors.New("mock audio pause failure")
	}
	return nil
}

func (p *MockProvider) ResumeAudio(ctx context.Context, sessionID string) error {
	if p.ShouldFail {
		return errors.New("mock audio resume failure")
	}
	return nil
}

func (p *MockProvider) RecordAudio(ctx context.Context, deviceID, destination string, durationMs int) error {
	if p.ShouldFail {
		return errors.New("mock audio record failure")
	}
	return nil
}

func (p *MockProvider) PlayVideo(ctx context.Context, path string) (string, error) {
	if p.ShouldFail {
		return "", errors.New("mock video play failure")
	}
	return "video-session-1", nil
}

func (p *MockProvider) PauseVideo(ctx context.Context, sessionID string) error {
	if p.ShouldFail {
		return errors.New("mock video pause failure")
	}
	return nil
}

func (p *MockProvider) ResumeVideo(ctx context.Context, sessionID string) error {
	if p.ShouldFail {
		return errors.New("mock video resume failure")
	}
	return nil
}

func (p *MockProvider) StopVideo(ctx context.Context, sessionID string) error {
	if p.ShouldFail {
		return errors.New("mock video stop failure")
	}
	return nil
}

func (p *MockProvider) RecordVideo(ctx context.Context, deviceID, destination string, durationMs int) error {
	if p.ShouldFail {
		return errors.New("mock video record failure")
	}
	return nil
}

func (p *MockProvider) CaptureImage(ctx context.Context, deviceID, destination string) error {
	if p.ShouldFail {
		return errors.New("mock capture failure")
	}
	return nil
}

func (p *MockProvider) LoadImage(ctx context.Context, path string) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock load image failure")
	}
	return map[string]interface{}{"width": 1920, "height": 1080, "format": "jpeg"}, nil
}

func (p *MockProvider) SaveImage(ctx context.Context, destination string, data []byte) error {
	if p.ShouldFail {
		return errors.New("mock save image failure")
	}
	return nil
}

func (p *MockProvider) GetMetadata(ctx context.Context, path string) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock metadata failure")
	}
	return map[string]interface{}{"duration": 120, "resolution": "1080p"}, nil
}

func (p *MockProvider) ListMediaDevices(ctx context.Context, deviceType string) ([]map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock list devices failure")
	}
	return []map[string]interface{}{{"id": "cam-1", "name": "Mock Camera"}}, nil
}

func (p *MockProvider) GetDevice(ctx context.Context, deviceID string) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock get device failure")
	}
	return map[string]interface{}{"id": deviceID, "name": "Mock Device"}, nil
}

func (p *MockProvider) ListCodecs(ctx context.Context) ([]string, error) {
	if p.ShouldFail {
		return nil, errors.New("mock list codecs failure")
	}
	return []string{"h264", "mp3", "aac"}, nil
}
