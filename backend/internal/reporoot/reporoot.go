package reporoot

import (
	"os"
	"path/filepath"
)

// Root 返回仓库根目录（含 backend / frontend 的目录）。
// 从当前工作目录向上最多找一层，找不到就回退到 cwd。
func Root() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	for _, dir := range []string{cwd, filepath.Join(cwd, "..")} {
		if isRepoRoot(dir) {
			abs, err := filepath.Abs(dir)
			if err == nil {
				return abs
			}
			return dir
		}
	}
	abs, err := filepath.Abs(cwd)
	if err != nil {
		return cwd
	}
	return abs
}

func isRepoRoot(dir string) bool {
	_, err1 := os.Stat(filepath.Join(dir, "backend"))
	_, err2 := os.Stat(filepath.Join(dir, "frontend"))
	return err1 == nil && err2 == nil
}
