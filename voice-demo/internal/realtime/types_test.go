package realtime

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseAudioDelta(t *testing.T) {
	raw := `{"event_id":"e1","type":"response.audio.delta","item_id":"msg_8","delta":"QUJDRA=="}`
	event, err := ParseServerEvent([]byte(raw))
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	audio, ok := event.(AudioDeltaEvent)
	if !ok {
		t.Fatalf("类型应为 AudioDeltaEvent，实际为 %T", event)
	}
	if audio.Delta != "QUJDRA==" {
		t.Errorf("Delta 解析错误: %q", audio.Delta)
	}
	if audio.ItemID != "msg_8" {
		t.Errorf("ItemID 解析错误: %q", audio.ItemID)
	}
}

func TestParseResponseDone(t *testing.T) {
	raw := `{
		"event_id":"e2",
		"type":"response.done",
		"response":{
			"id":"resp_1",
			"status":"completed",
			"output":[{
				"id":"msg_6",
				"type":"message",
				"status":"completed",
				"role":"assistant",
				"content":[{"type":"text","text":"你好"}]
			}]
		}
	}`
	event, err := ParseServerEvent([]byte(raw))
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	done, ok := event.(ResponseDoneEvent)
	if !ok {
		t.Fatalf("类型应为 ResponseDoneEvent，实际为 %T", event)
	}
	if done.Response.Status != "completed" {
		t.Errorf("状态错误: %q", done.Response.Status)
	}
	if len(done.Response.Output) != 1 {
		t.Fatalf("output 数量错误: %d", len(done.Response.Output))
	}
	if got := done.Response.Output[0].Content[0].Text; got != "你好" {
		t.Errorf("文本错误: %q", got)
	}
}

func TestParseSessionCreated(t *testing.T) {
	raw := `{"event_id":"e3","type":"session.created","session":{"id":"sess_1","model":"stepaudio-2.5-realtime","voice":"linjiajiejie"}}`
	event, err := ParseServerEvent([]byte(raw))
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	created, ok := event.(SessionCreatedEvent)
	if !ok {
		t.Fatalf("类型应为 SessionCreatedEvent，实际为 %T", event)
	}
	if created.Session.Model != "stepaudio-2.5-realtime" {
		t.Errorf("模型错误: %q", created.Session.Model)
	}
}

func TestParseError(t *testing.T) {
	raw := `{"event_id":"e4","type":"error","error":{"type":"invalid_request_error","code":"invalid_param","message":"音频内容不完整"}}`
	event, err := ParseServerEvent([]byte(raw))
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	errEvent, ok := event.(ErrorEvent)
	if !ok {
		t.Fatalf("类型应为 ErrorEvent，实际为 %T", event)
	}
	if errEvent.Error.Code != "invalid_param" {
		t.Errorf("错误码错误: %q", errEvent.Error.Code)
	}
}

func TestParseUnknownType(t *testing.T) {
	raw := `{"type":"scooby.dooby.doo"}`
	if _, err := ParseServerEvent([]byte(raw)); err == nil {
		t.Fatal("未知类型应返回错误")
	}
}

func TestSessionUpdateMarshal(t *testing.T) {
	event := SessionUpdateEvent{
		EventID: "e5",
		Type:    EventSessionUpdate,
		Session: SessionConfig{
			Modalities:        []string{"text", "audio"},
			Instructions:      "你是友好的助手",
			Voice:             "linjiajiejie",
			InputAudioFormat:  "pcm16",
			OutputAudioFormat: "pcm16",
			TurnDetection: &TurnDetection{
				Type:            "server_vad",
				PrefixPaddingMs: 500,
			},
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}
	s := string(data)
	for _, want := range []string{"session.update", "server_vad", "linjiajiejie", "pcm16"} {
		if !strings.Contains(s, want) {
			t.Errorf("序列化结果缺少 %q: %s", want, s)
		}
	}
}
