package plot

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Cache 按原片路径+大小把剧情 JSON 落盘。
type Cache struct {
	dir string
}

func newCache() *Cache {
	root, err := os.UserCacheDir()
	if err != nil {
		root = os.TempDir()
	}
	dir := filepath.Join(root, "withyou", "plot")
	_ = os.MkdirAll(dir, 0o755)
	return &Cache{dir: dir}
}

func cacheKey(path string, size int64) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%d", path, size)))
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
