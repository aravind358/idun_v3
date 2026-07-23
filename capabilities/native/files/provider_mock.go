package files

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

func (p *MockProvider) ReadFile(ctx context.Context, path string) ([]byte, error) {
	if p.ShouldFail {
		return nil, errors.New("mock read error")
	}
	return []byte("mock data"), nil
}

func (p *MockProvider) ReadText(ctx context.Context, path string) (string, error) {
	if p.ShouldFail {
		return "", errors.New("mock read text error")
	}
	return "mock data", nil
}

func (p *MockProvider) FileExists(ctx context.Context, path string) (bool, error) {
	if p.ShouldFail {
		return false, errors.New("mock exists error")
	}
	return true, nil
}

func (p *MockProvider) GetMetadata(ctx context.Context, path string) (map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock metadata error")
	}
	return map[string]interface{}{"name": "mock.txt"}, nil
}

func (p *MockProvider) WriteFile(ctx context.Context, path string, data []byte, append bool) error {
	if p.ShouldFail {
		return errors.New("mock write error")
	}
	return nil
}

func (p *MockProvider) CopyFile(ctx context.Context, src, dest string) error {
	if p.ShouldFail {
		return errors.New("mock copy error")
	}
	return nil
}

func (p *MockProvider) MoveFile(ctx context.Context, src, dest string) error {
	if p.ShouldFail {
		return errors.New("mock move error")
	}
	return nil
}

func (p *MockProvider) DeleteFile(ctx context.Context, path string) error {
	if p.ShouldFail {
		return errors.New("mock delete error")
	}
	return nil
}

func (p *MockProvider) ListDirectory(ctx context.Context, path string, recursive bool) ([]map[string]interface{}, error) {
	if p.ShouldFail {
		return nil, errors.New("mock list error")
	}
	return []map[string]interface{}{{"name": "mock.txt"}}, nil
}

func (p *MockProvider) CreateDirectory(ctx context.Context, path string) error {
	if p.ShouldFail {
		return errors.New("mock create dir error")
	}
	return nil
}

func (p *MockProvider) DeleteDirectory(ctx context.Context, path string) error {
	if p.ShouldFail {
		return errors.New("mock delete dir error")
	}
	return nil
}

func (p *MockProvider) SearchFiles(ctx context.Context, root, pattern string, recursive, caseSensitive bool) ([]string, error) {
	if p.ShouldFail {
		return nil, errors.New("mock search error")
	}
	return []string{"mock.txt"}, nil
}

func (p *MockProvider) CalculateHash(ctx context.Context, path, algorithm string) (string, error) {
	if p.ShouldFail {
		return "", errors.New("mock hash error")
	}
	return "mockhash123", nil
}

func (p *MockProvider) CreateTemporaryFile(ctx context.Context, prefix, suffix string) (string, error) {
	if p.ShouldFail {
		return "", errors.New("mock temp file error")
	}
	return "/tmp/mock_temp_file", nil
}

func (p *MockProvider) CreateTemporaryDirectory(ctx context.Context, prefix string) (string, error) {
	if p.ShouldFail {
		return "", errors.New("mock temp dir error")
	}
	return "/tmp/mock_temp_dir", nil
}
