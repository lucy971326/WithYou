package library

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ncruces/zenity"
)

// ErrCancelled 用户关掉了系统选框。
var ErrCancelled = errors.New("file pick cancelled")

// Picker 弹出系统选框，只返回本机路径，不写入播放状态。
type Picker struct{}

// Pick 阻塞直到用户选定或取消。ctx 取消不关系统窗口。
func (p *Picker) Pick(ctx context.Context) (OpenedFile, error) {
	path, err := zenity.SelectFile(
		zenity.Title("打开视频"),
		zenity.FileFilters{
			{Name: "视频", Patterns: []string{"*.mkv", "*.mp4"}, CaseFold: true},
		},
	)
	if err != nil {
		if errors.Is(err, zenity.ErrCanceled) {
			return OpenedFile{}, ErrCancelled
		}
		return OpenedFile{}, fmt.Errorf("library: native file dialog: %w", err)
	}
	err = ctx.Err()
	if err != nil {
		return OpenedFile{}, err
	}

	info, err := os.Stat(path)
	if err != nil {
		return OpenedFile{}, fmt.Errorf("library: stat picked file: %w", err)
	}
	if info.IsDir() {
		return OpenedFile{}, fmt.Errorf("library: picked path is a directory")
	}

	return OpenedFile{
		Path: path,
		Name: filepath.Base(path),
		Size: info.Size(),
	}, nil
}
