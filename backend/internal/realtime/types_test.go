package realtime

import "testing"

func TestInspectClientAllowlist(t *testing.T) {
	typ, err := inspectClient([]byte(`{"type":"session.update","session":{}}`))
	if err != nil || typ != TypeSessionUpdate {
		t.Fatalf("update: typ=%s err=%v", typ, err)
	}
	if _, err := inspectClient([]byte(`{"type":"made_up"}`)); err == nil {
		t.Fatal("expected drop")
	}
	if _, err := inspectClient([]byte(`{
		"type":"conversation.item.create",
		"item":{"content":[{"type":"input_image"}]}
	}`)); err == nil {
		t.Fatal("expected drop image item")
	}
	if _, err := inspectClient([]byte(`{
		"type":"conversation.item.create",
		"item":{"content":[{"type":"input_text","text":"hi"}]}
	}`)); err != nil {
		t.Fatal(err)
	}
	big := make([]byte, 256*1024+8)
	for i := range big {
		big[i] = 'A'
	}
	if _, err := inspectClient([]byte(`{"type":"input_image_buffer.append","image":"` + string(big) + `"}`)); err == nil {
		t.Fatal("expected drop oversized image")
	}
}
