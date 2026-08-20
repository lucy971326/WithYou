package realtime

import (
	"encoding/json"
	"fmt"
)

const (
	TypeSessionUpdate           = "session.update"
	TypeAudioAppend             = "input_audio_buffer.append"
	TypeImageAppend             = "input_image_buffer.append"
	TypeResponseCancel          = "response.cancel"
	TypeItemCreate              = "conversation.item.create"
	TypeSessionFinish           = "session.finish"
	TypeAudioCommit             = "input_audio_buffer.commit"
	TypeAudioClear              = "input_audio_buffer.clear"
	TypeResponseCreate          = "response.create"
	TypeError                   = "error"
	TypeSessionCreated          = "session.created"
	TypeSessionUpdated          = "session.updated"
	TypeSpeechStarted           = "input_audio_buffer.speech_started"
	TypeSpeechStopped           = "input_audio_buffer.speech_stopped"
	TypeResponseAudioDelta      = "response.audio.delta"
	TypeResponseAudioTransDelta = "response.audio_transcript.delta"
	TypeUserTranscriptionDelta  = "conversation.item.input_audio_transcription.delta"
	TypeResponseTextDelta       = "response.text.delta"
	TypeResponseCreated         = "response.created"
	TypeResponseDone            = "response.done"
)

var clientAllow = map[string]bool{
	TypeSessionUpdate:  true,
	TypeAudioAppend:    true,
	TypeImageAppend:    true,
	TypeResponseCancel: true,
	TypeItemCreate:     true,
	TypeSessionFinish:  true,
	TypeAudioCommit:    true,
	TypeAudioClear:     true,
	TypeResponseCreate: true,
}

type envelope struct {
	Type  string `json:"type"`
	Image string `json:"image"`
	Error *struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error"`
	Session *struct {
		ID string `json:"id"`
	} `json:"session"`
	Item *struct {
		Content []struct {
			Type string `json:"type"`
		} `json:"content"`
	} `json:"item"`
}

func parseEnvelope(raw []byte) (envelope, error) {
	var env envelope
	err := json.Unmarshal(raw, &env)
	if err != nil {
		return envelope{}, fmt.Errorf("realtime: not json: %w", err)
	}
	if env.Type == "" {
		return envelope{}, fmt.Errorf("realtime: missing type")
	}
	return env, nil
}

func inspectClient(raw []byte) (string, error) {
	env, err := parseEnvelope(raw)
	if err != nil {
		return "", err
	}
	if !clientAllow[env.Type] {
		return env.Type, fmt.Errorf("type %q not in allowlist", env.Type)
	}
	if env.Type == TypeItemCreate && env.Item != nil {
		for _, c := range env.Item.Content {
			if c.Type == "input_image" {
				return env.Type, fmt.Errorf("conversation.item.create must not carry input_image")
			}
		}
	}
	if env.Type == TypeImageAppend && len(env.Image) > 256*1024 {
		return env.Type, fmt.Errorf("image base64 %d exceeds 256KB", len(env.Image))
	}
	return env.Type, nil
}
