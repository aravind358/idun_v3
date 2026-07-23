package files

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type NativeProvider struct{}

func NewNativeProvider() *NativeProvider {
	return &NativeProvider{}
}

func (p *NativeProvider) ReadFile(ctx context.Context, path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (p *NativeProvider) ReadText(ctx context.Context, path string) (string, error) {
	data, err := p.ReadFile(ctx, path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (p *NativeProvider) FileExists(ctx context.Context, path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (p *NativeProvider) GetMetadata(ctx context.Context, path string) (map[string]interface{}, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	
	// Normalize path
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	return map[string]interface{}{
		"name":          info.Name(),
		"path":          absPath,
		"size":          info.Size(),
		"is_directory":  info.IsDir(),
		"modified_time": info.ModTime().Format(time.RFC3339),
		"permissions":   info.Mode().String(),
	}, nil
}

func (p *NativeProvider) WriteFile(ctx context.Context, path string, data []byte, append bool) error {
	flag := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if append {
		flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	}
	f, err := os.OpenFile(path, flag, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

func (p *NativeProvider) CopyFile(ctx context.Context, src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func (p *NativeProvider) MoveFile(ctx context.Context, src, dest string) error {
	return os.Rename(src, dest)
}

func (p *NativeProvider) DeleteFile(ctx context.Context, path string) error {
	return os.Remove(path)
}

func (p *NativeProvider) ListDirectory(ctx context.Context, path string, recursive bool) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	
	if !recursive {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			info, err := e.Info()
			if err != nil {
				continue
			}
			results = append(results, map[string]interface{}{
				"name":         e.Name(),
				"is_directory": e.IsDir(),
				"size":         info.Size(),
			})
		}
		return results, nil
	}

	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if p == path {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		results = append(results, map[string]interface{}{
			"path":         p,
			"name":         d.Name(),
			"is_directory": d.IsDir(),
			"size":         info.Size(),
		})
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (p *NativeProvider) CreateDirectory(ctx context.Context, path string) error {
	return os.MkdirAll(path, 0755)
}

func (p *NativeProvider) DeleteDirectory(ctx context.Context, path string) error {
	return os.RemoveAll(path)
}

func (p *NativeProvider) SearchFiles(ctx context.Context, root, pattern string, recursive, caseSensitive bool) ([]string, error) {
	var matches []string
	if !caseSensitive {
		pattern = strings.ToLower(pattern)
	}
	
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		
		name := d.Name()
		if !caseSensitive {
			name = strings.ToLower(name)
		}
		
		matched, _ := filepath.Match(pattern, name)
		if matched || strings.Contains(name, pattern) {
			matches = append(matches, path)
		}
		
		if !recursive && d.IsDir() && path != root {
			return filepath.SkipDir
		}
		return nil
	})
	
	return matches, err
}

func (p *NativeProvider) CalculateHash(ctx context.Context, path, algorithm string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var result []byte

	switch strings.ToLower(algorithm) {
	case "md5":
		hash := md5.New()
		if _, err := io.Copy(hash, f); err != nil {
			return "", err
		}
		result = hash.Sum(nil)
	case "sha1":
		hash := sha1.New()
		if _, err := io.Copy(hash, f); err != nil {
			return "", err
		}
		result = hash.Sum(nil)
	case "sha256":
		hash := sha256.New()
		if _, err := io.Copy(hash, f); err != nil {
			return "", err
		}
		result = hash.Sum(nil)
	default:
		return "", errors.New("unsupported hash algorithm")
	}

	return hex.EncodeToString(result), nil
}

func (p *NativeProvider) CreateTemporaryFile(ctx context.Context, prefix, suffix string) (string, error) {
	f, err := os.CreateTemp("", prefix+"*"+suffix)
	if err != nil {
		return "", err
	}
	f.Close()
	return f.Name(), nil
}

func (p *NativeProvider) CreateTemporaryDirectory(ctx context.Context, prefix string) (string, error) {
	return os.MkdirTemp("", prefix+"*")
}
