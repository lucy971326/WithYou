package plot

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var srtTimeRe = regexp.MustCompile(
	`(\d{1,2}):(\d{2}):(\d{2})[,.](\d{1,3})\s*-->\s*(\d{1,2}):(\d{2}):(\d{2})[,.](\d{1,3})`,
)

var tagRe = regexp.MustCompile(`\{[^}]*\}|<[^>]+>`)

// Parser 把 SRT / ASS 原文收成 Cue 列表。
type Parser struct{}

// Parse 按格式解析。format 为 "srt" 或 "ass"。
func (p *Parser) Parse(content, format string) ([]Cue, error) {
	content = strings.TrimPrefix(content, "\ufeff")
	switch strings.ToLower(format) {
	case "srt":
		return parseSRT(content), nil
	case "ass", "ssa":
		return parseASS(content), nil
	default:
		return nil, fmt.Errorf("plot: unsupported subtitle format %q", format)
	}
}

func parseSRT(content string) []Cue {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	blocks := strings.Split(strings.TrimSpace(content), "\n\n")
	cues := make([]Cue, 0, len(blocks))
	for _, block := range blocks {
		lines := splitNonEmpty(block)
		if len(lines) == 0 {
			continue
		}
		var timeLine string
		var textLines []string
		for _, ln := range lines {
			if srtTimeRe.MatchString(ln) {
				timeLine = ln
				continue
			}
			if _, err := strconv.Atoi(ln); err == nil {
				continue
			}
			textLines = append(textLines, ln)
		}
		if timeLine == "" {
			continue
		}
		m := srtTimeRe.FindStringSubmatch(timeLine)
		if m == nil {
			continue
		}
		text := cleanText(strings.Join(textLines, " "))
		if text == "" {
			continue
		}
		cues = append(cues, Cue{
			StartSec: srtStamp(m[1], m[2], m[3], m[4]),
			EndSec:   srtStamp(m[5], m[6], m[7], m[8]),
			Text:     text,
		})
	}
	return cues
}

func parseASS(content string) []Cue {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	format := []string{"Layer", "Start", "End", "Style", "Name", "MarginL", "MarginR", "MarginV", "Effect", "Text"}
	inEvents := false
	var cues []Cue
	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			inEvents = strings.EqualFold(line, "[Events]")
			continue
		}
		if !inEvents {
			continue
		}
		if strings.HasPrefix(strings.ToLower(line), "format:") {
			format = parseASSFormat(line)
			continue
		}
		if !strings.HasPrefix(strings.ToLower(line), "dialogue:") {
			continue
		}
		rest := strings.TrimSpace(line[len("Dialogue:"):])
		fields := splitASSFields(rest, len(format))
		if len(fields) != len(format) {
			continue
		}
		byName := map[string]string{}
		for i, name := range format {
			byName[strings.ToLower(strings.TrimSpace(name))] = fields[i]
		}
		start, ok1 := parseASSTime(byName["start"])
		end, ok2 := parseASSTime(byName["end"])
		text := cleanText(byName["text"])
		if !ok1 || !ok2 || text == "" {
			continue
		}
		cues = append(cues, Cue{
			StartSec: start,
			EndSec:   end,
			Speaker:  strings.TrimSpace(byName["name"]),
			Text:     text,
		})
	}
	return cues
}

func parseASSFormat(line string) []string {
	_, rest, ok := strings.Cut(line, ":")
	if !ok {
		return nil
	}
	parts := strings.Split(rest, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func splitASSFields(rest string, n int) []string {
	if n <= 0 {
		return nil
	}
	fields := make([]string, 0, n)
	remain := rest
	for i := 0; i < n-1; i++ {
		idx := strings.IndexByte(remain, ',')
		if idx < 0 {
			return nil
		}
		fields = append(fields, remain[:idx])
		remain = remain[idx+1:]
	}
	fields = append(fields, remain)
	return fields
}

func parseASSTime(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0, false
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	secParts := strings.SplitN(parts[2], ".", 2)
	if err1 != nil || err2 != nil || len(secParts) == 0 {
		return 0, false
	}
	sec, err3 := strconv.Atoi(secParts[0])
	if err3 != nil {
		return 0, false
	}
	frac := 0.0
	if len(secParts) == 2 {
		digits := secParts[1]
		n, err := strconv.Atoi(digits)
		if err != nil {
			return 0, false
		}
		if len(digits) <= 2 {
			frac = float64(n) / 100
		} else {
			frac = float64(n) / 1000
		}
	}
	return float64(h*3600+m*60+sec) + frac, true
}

func srtStamp(h, m, s, frac string) float64 {
	hh, _ := strconv.Atoi(h)
	mm, _ := strconv.Atoi(m)
	ss, _ := strconv.Atoi(s)
	ms := frac
	for len(ms) < 3 {
		ms += "0"
	}
	if len(ms) > 3 {
		ms = ms[:3]
	}
	milli, _ := strconv.Atoi(ms)
	return float64(hh*3600+mm*60+ss) + float64(milli)/1000
}

func cleanText(text string) string {
	text = strings.ReplaceAll(text, `\N`, " ")
	text = strings.ReplaceAll(text, `\n`, " ")
	text = tagRe.ReplaceAllString(text, "")
	return strings.TrimSpace(text)
}

func splitNonEmpty(block string) []string {
	var out []string
	for _, ln := range strings.Split(block, "\n") {
		ln = strings.TrimSpace(ln)
		if ln != "" {
			out = append(out, ln)
		}
	}
	return out
}
