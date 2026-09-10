package cli

import "testing"

func TestNewPrompt(t *testing.T) {
	prompt, err := NewPrompt()
	if err != nil {
		t.Fatalf("expected prompt creation to succeed, got %v", err)
	}

	if prompt == nil {
		t.Fatal("expected prompt, got nil")
	}

	if err := prompt.Close(); err != nil {
		t.Fatalf("expected prompt to close successfully, got %v", err)
	}
}
