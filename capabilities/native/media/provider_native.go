package media

import (
	"context"
	"errors"
	"sync"
)

type NativeProvider struct {
	mu       sync.Mutex
	sessions int
}

func NewNativeProvider() *NativeProvider {
	return &NativeProvider{}
}

func (p *NativeProvider) PlayAudio(ctx context.Context, path string) (string, error) {
	return "", errors.New("native audio playback requires external bindings")
}

func (p *NativeProvider) StopAudio(ctx context.Context, sessionID string) error {
	return errors.New("native audio playback requires external bindings")
}

func (p *NativeProvider) PauseAudio(ctx context.Context, sessionID string) error {
	return errors.New("native audio playback requires external bindings")
}

func (p *NativeProvider) ResumeAudio(ctx context.Context, sessionID string) error {
	return errors.New("native audio playback requires external bindings")
}

func (p *NativeProvider) RecordAudio(ctx context.Context, deviceID, destination string, durationMs int) error {
	return errors.New("native audio recording requires external bindings")
}

func (p *NativeProvider) PlayVideo(ctx context.Context, path string) (string, error) {
	return "", errors.New("native video playback requires external bindings")
}

func (p *NativeProvider) PauseVideo(ctx context.Context, sessionID string) error {
	return errors.New("native video playback requires external bindings")
}

func (p *NativeProvider) ResumeVideo(ctx context.Context, sessionID string) error {
	return errors.New("native video playback requires external bindings")
}

func (p *NativeProvider) StopVideo(ctx context.Context, sessionID string) error {
	return errors.New("native video playback requires external bindings")
}

func (p *NativeProvider) RecordVideo(ctx context.Context, deviceID, destination string, durationMs int) error {
	return errors.New("native video recording requires external bindings")
}

func (p *NativeProvider) CaptureImage(ctx context.Context, deviceID, destination string) error {
	return errors.New("native camera capture requires external bindings")
}

func (p *NativeProvider) LoadImage(ctx context.Context, path string) (map[string]interface{}, error) {
	// Standard image loading could use image/jpeg, image/png but keeping it generic.
	return nil, errors.New("native image loading not fully implemented")
}

func (p *NativeProvider) SaveImage(ctx context.Context, destination string, data []byte) error {
	return errors.New("native image saving not fully implemented")
}

func (p *NativeProvider) GetMetadata(ctx context.Context, path string) (map[string]interface{}, error) {
	return nil, errors.New("native media metadata parsing requires external bindings")
}

func (p *NativeProvider) ListMediaDevices(ctx context.Context, deviceType string) ([]map[string]interface{}, error) {
	return nil, errors.New("native device enumeration requires external bindings")
}

func (p *NativeProvider) GetDevice(ctx context.Context, deviceID string) (map[string]interface{}, error) {
	return nil, errors.New("native device lookup requires external bindings")
}

func (p *NativeProvider) ListCodecs(ctx context.Context) ([]string, error) {
	return nil, errors.New("native codec enumeration requires external bindings")
}
