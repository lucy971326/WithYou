package plot

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Cache 按原片路径+大小把剧情 JSON 落盘到仓库根目录 cache/plot。
type Cache struct {
	dir string
}

func newCache() *Cache {
	dir := filepath.Join(projectRoot(), "cache", "plot")
	_ = os.MkdirAll(dir, 0o755)
	migrateOldCache(dir)
	log.Printf("plot cache: %s", dir)
	return &Cache{dir: dir}
}

func projectRoot() string {
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

func migrateOldCache(dstDir string) {
	home, err := os.UserCacheDir()
	if err != nil {
		return
	}
	srcDir := filepath.Join(home, "withyou", "plot")
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		dst := filepath.Join(dstDir, e.Name())
		if _, err := os.Stat(dst); err == nil {
			continue
		}
		src := filepath.Join(srcDir, e.Name())
		if err := copyFile(src, dst); err != nil {
			log.Printf("plot cache migrate %s: %v", e.Name(), err)
		}
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func cacheKey(path string, size int64) string {
	sum := sha256.Sum256(fmt.Appendf(nil, "%s|%d", path, size))
	return hex.EncodeToString(sum[:])
}

// Get 命中则返回文档。
func (c *Cache) Get(path string, size int64) (PlotDoc, bool) {
	data, err := os.ReadFile(c.file(path, size))
	if err != nil {
		return PlotDoc{}, false
	}
	var doc PlotDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return PlotDoc{}, false
	}
	return doc, true
}

// Put 写入缓存。
func (c *Cache) Put(path string, size int64, doc PlotDoc) error {
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.file(path, size), data, 0o644)
}

func (c *Cache) file(path string, size int64) string {
	return filepath.Join(c.dir, cacheKey(path, size)+".json")
}
