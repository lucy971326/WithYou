package realtime

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/coder/websocket"
)

// maxWSBytes 覆盖官方图片上限（base64 ≤ 256KB）再留余量。coder/websocket 默认只读 32768。
const maxWSBytes = 1 << 20

// Relay 浏览器 ↔ Qwen 的 WS 中继。过滤客户端 type，逐条打日志。
type Relay struct {
	apiKey string
	model  string
	base   string
}

func (r *Relay) Serve(ctx context.Context, browser *websocket.Conn) error {
	browser.SetReadLimit(maxWSBytes)
	qwenURL, err := r.dialURL()
	if err != nil {
		return err
	}
	log.Printf("realtime dial model=%s", r.model)

	header := http.Header{}
	header.Set("Authorization", "Bearer "+r.apiKey)
	qwen, resp, err := websocket.Dial(ctx, qwenURL, &websocket.DialOptions{HTTPHeader: header})
	if err != nil {
		status := 0
		if resp != nil {
			status = resp.StatusCode
		}
		log.Printf("realtime dial failed status=%d err=%v", status, err)
		return fmt.Errorf("realtime: dial qwen: %w", err)
	}
	qwen.SetReadLimit(maxWSBytes)
	defer qwen.Close(websocket.StatusNormalClosure, "done")
	log.Printf("realtime qwen connected read_limit=%d", maxWSBytes)

	errc := make(chan error, 2)
	go func() {
		errc <- r.pump(ctx, "browser→qwen", browser, qwen, true)
	}()
	go func() {
		errc <- r.pump(ctx, "qwen→browser", qwen, browser, false)
	}()
	err = <-errc
	log.Printf("realtime session end: %v", err)
	return err
}

func (r *Relay) pump(ctx context.Context, dir string, src, dst *websocket.Conn, filter bool) error {
	for {
		typ, data, err := src.Read(ctx)
		if err != nil {
			return err
		}
		if typ != websocket.MessageText {
			log.Printf("realtime %s DROP non-text opcode=%v bytes=%d", dir, typ, len(data))
			if filter {
				continue
			}
		}
		env, perr := parseEnvelope(data)
		name := "?"
		if perr == nil {
			name = env.Type
		}
		if filter {
			got, ierr := inspectClient(data)
			if ierr != nil {
				log.Printf("realtime %s DROP type=%s bytes=%d reason=%v", dir, got, len(data), ierr)
				continue
			}
			name = got
		}
		extra := ""
		if env.Type == TypeError && env.Error != nil {
			extra = fmt.Sprintf(" code=%s msg=%s", env.Error.Code, env.Error.Message)
		}
		if env.Type == TypeSessionCreated && env.Session != nil {
			extra = " session_id=" + env.Session.ID
		}
		log.Printf("realtime %s type=%s bytes=%d%s", dir, name, len(data), extra)
		if err := dst.Write(ctx, websocket.MessageText, data); err != nil {
			return err
		}
	}
}

func (r *Relay) dialURL() (string, error) {
	u, err := url.Parse(r.base)
	if err != nil {
		return "", fmt.Errorf("realtime: bad url: %w", err)
	}
	q := u.Query()
	if q.Get("model") == "" {
		q.Set("model", r.model)
	}
	u.RawQuery = q.Encode()
	if !strings.HasPrefix(u.Scheme, "ws") {
		return "", fmt.Errorf("realtime: url must be ws/wss")
	}
	return u.String(), nil
}
