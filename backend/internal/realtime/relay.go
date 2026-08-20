package realtime

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/coder/websocket"
)

// maxWSBytes 覆盖官方图上限（base64 ≤256KB）。库默认只读 32KB，不改会 1009。
const maxWSBytes = 1 << 20

// Relay 是一次会话：浏览器 WS ↔ Qwen WS。
// Key 只在这边用来拨号；事件本身不改写，只过白名单。
type Relay struct {
	apiKey string
	model  string
	base   string

	quietMu     sync.Mutex
	quietCounts map[string]int
}

// quietTypes 是高频流式事件，只抽样打日志，避免刷屏。
var quietTypes = map[string]bool{
	TypeAudioAppend:             true,
	TypeResponseAudioDelta:      true,
	TypeResponseAudioTransDelta: true,
	TypeUserTranscriptionDelta:  true,
	TypeResponseTextDelta:       true,
}

// 输入：ctx、已升级的浏览器 WS。输出：会话结束原因。
// 拨通 Qwen，起两条单向泵；任一方向先死即返回。
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

	// Read 会堵住，必须拆成两条单向泵才能同时听两边。
	// 浏览器侧不信任 → filter；Qwen 侧原样透传。
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

// 输入：ctx、方向名、源/目标连接、是否过滤客户端。输出：读或写失败。
// 从 src 读一条写到 dst；filter 时未知 type / 图塞进 item.create 则 DROP。
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
		r.logEvent(dir, name, len(data), extra)
		err = dst.Write(ctx, websocket.MessageText, data)
		if err != nil {
			return err
		}
	}
}

// logEvent 普通事件原样打；高频事件首次 + 每 50 条打一次计数。
func (r *Relay) logEvent(dir, name string, n int, extra string) {
	if !quietTypes[name] {
		log.Printf("realtime %s type=%s bytes=%d%s", dir, name, n, extra)
		return
	}
	r.quietMu.Lock()
	r.quietCounts[name]++
	count := r.quietCounts[name]
	r.quietMu.Unlock()
	if count == 1 || count%50 == 0 {
		log.Printf("realtime %s type=%s bytes=%d total=%d", dir, name, n, count)
	}
}

// 输入：本结构体的 base、model。输出：带 model 的 wss URL，或拼失败。
// 拼 Qwen 拨号地址。
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
