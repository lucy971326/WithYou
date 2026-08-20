package realtime

import (
	"strings"
	"testing"
)

func TestCurrentPromptHasPlotSlot(t *testing.T) {
	prompt := currentPrompt()
	if prompt.Version == "" {
		t.Fatal("prompt version is empty")
	}
	if !strings.Contains(prompt.Instructions, "{{PLOT_JSON}}") {
		t.Fatal("prompt must expose the plot placeholder")
	}
	if strings.Contains(prompt.Instructions, "长期陪看者") == false {
		t.Fatal("prompt must contain the companion persona")
	}
}
