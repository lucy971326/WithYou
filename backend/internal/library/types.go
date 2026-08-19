package library

// OpenedFile 是当前打开的本机视频。Path 只给本机能力用；浏览器播同一份原片。
type OpenedFile struct {
	Path         string
	PlayablePath string
	Name         string
	Size         int64
	Codec        string
	PixelFormat  string
	BrowserSafe  bool
}

// OpenResponse 是 POST /api/open 成功体。
type OpenResponse struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	Codec       string `json:"codec"`
	PixelFormat string `json:"pixelFormat"`
	BrowserSafe bool   `json:"browserSafe"`
}

// CurrentResponse 是 GET /api/current 成功体。
type CurrentResponse struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	Codec       string `json:"codec"`
	PixelFormat string `json:"pixelFormat"`
	BrowserSafe bool   `json:"browserSafe"`
}

// ErrorResponse 是 HTTP 错误体。
type ErrorResponse struct {
	Error string `json:"error"`
}
