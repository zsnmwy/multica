package agent

import (
	"log/slog"
	"strings"
	"testing"
)

func TestCopilotHandleMessageResponse(t *testing.T) {
	t.Parallel()

	b := &copilotBackend{cfg: Config{Logger: slog.Default()}}
	ch := make(chan Message, 10)
	var output strings.Builder

	msg := copilotMessage{
		Type:    "response",
		Content: "echo 'Hello World'",
	}

	b.handleMessage(msg, ch, &output)

	if output.String() != "echo 'Hello World'" {
		t.Fatalf("expected output 'echo 'Hello World'', got %q", output.String())
	}
	select {
	case m := <-ch:
		if m.Type != MessageText || m.Content != "echo 'Hello World'" {
			t.Fatalf("unexpected message: %+v", m)
		}
	default:
		t.Fatal("expected message on channel")
	}
}

func TestCopilotHandleMessageText(t *testing.T) {
	t.Parallel()

	b := &copilotBackend{cfg: Config{Logger: slog.Default()}}
	ch := make(chan Message, 10)
	var output strings.Builder

	msg := copilotMessage{
		Type:    "text",
		Content: "Some text content",
	}

	b.handleMessage(msg, ch, &output)

	if output.String() != "Some text content" {
		t.Fatalf("expected output 'Some text content', got %q", output.String())
	}
	select {
	case m := <-ch:
		if m.Type != MessageText || m.Content != "Some text content" {
			t.Fatalf("unexpected message: %+v", m)
		}
	default:
		t.Fatal("expected message on channel")
	}
}

func TestCopilotHandleMessageExplanation(t *testing.T) {
	t.Parallel()

	b := &copilotBackend{cfg: Config{Logger: slog.Default()}}
	ch := make(chan Message, 10)
	var output strings.Builder

	msg := copilotMessage{
		Type:    "explanation",
		Content: "This is an explanation",
	}

	b.handleMessage(msg, ch, &output)

	if output.String() != "" {
		t.Fatalf("explanation should not add to output, got %q", output.String())
	}
	select {
	case m := <-ch:
		if m.Type != MessageThinking || m.Content != "This is an explanation" {
			t.Fatalf("unexpected message: %+v", m)
		}
	default:
		t.Fatal("expected message on channel")
	}
}

func TestCopilotHandleMessageThinking(t *testing.T) {
	t.Parallel()

	b := &copilotBackend{cfg: Config{Logger: slog.Default()}}
	ch := make(chan Message, 10)
	var output strings.Builder

	msg := copilotMessage{
		Type:    "thinking",
		Content: "Thinking about the problem",
	}

	b.handleMessage(msg, ch, &output)

	if output.String() != "" {
		t.Fatalf("thinking should not add to output, got %q", output.String())
	}
	select {
	case m := <-ch:
		if m.Type != MessageThinking || m.Content != "Thinking about the problem" {
			t.Fatalf("unexpected message: %+v", m)
		}
	default:
		t.Fatal("expected message on channel")
	}
}

func TestCopilotHandleMessageCommand(t *testing.T) {
	t.Parallel()

	b := &copilotBackend{cfg: Config{Logger: slog.Default()}}
	ch := make(chan Message, 10)
	var output strings.Builder

	msg := copilotMessage{
		Type:    "command",
		Command: "ls -la",
	}

	b.handleMessage(msg, ch, &output)

	if output.String() != "" {
		t.Fatalf("command should not add to output, got %q", output.String())
	}
	select {
	case m := <-ch:
		if m.Type != MessageToolUse || m.Tool != "exec_command" {
			t.Fatalf("unexpected message: %+v", m)
		}
		if m.Input["command"] != "ls -la" {
			t.Fatalf("expected command 'ls -la', got %v", m.Input["command"])
		}
	default:
		t.Fatal("expected message on channel")
	}
}

func TestCopilotHandleMessageToolUse(t *testing.T) {
	t.Parallel()

	b := &copilotBackend{cfg: Config{Logger: slog.Default()}}
	ch := make(chan Message, 10)
	var output strings.Builder

	msg := copilotMessage{
		Type:    "tool_use",
		Command: "git status",
	}

	b.handleMessage(msg, ch, &output)

	if output.String() != "" {
		t.Fatalf("tool_use should not add to output, got %q", output.String())
	}
	select {
	case m := <-ch:
		if m.Type != MessageToolUse || m.Tool != "exec_command" {
			t.Fatalf("unexpected message: %+v", m)
		}
		if m.Input["command"] != "git status" {
			t.Fatalf("expected command 'git status', got %v", m.Input["command"])
		}
	default:
		t.Fatal("expected message on channel")
	}
}

func TestCopilotHandleMessageResult(t *testing.T) {
	t.Parallel()

	b := &copilotBackend{cfg: Config{Logger: slog.Default()}}
	ch := make(chan Message, 10)
	var output strings.Builder

	msg := copilotMessage{
		Type:   "result",
		Output: "command output here",
	}

	b.handleMessage(msg, ch, &output)

	if output.String() != "" {
		t.Fatalf("result should not add to output, got %q", output.String())
	}
	select {
	case m := <-ch:
		if m.Type != MessageToolResult || m.Tool != "exec_command" {
			t.Fatalf("unexpected message: %+v", m)
		}
		if m.Output != "command output here" {
			t.Fatalf("expected output 'command output here', got %v", m.Output)
		}
	default:
		t.Fatal("expected message on channel")
	}
}

func TestCopilotHandleMessageToolResult(t *testing.T) {
	t.Parallel()

	b := &copilotBackend{cfg: Config{Logger: slog.Default()}}
	ch := make(chan Message, 10)
	var output strings.Builder

	msg := copilotMessage{
		Type:   "tool_result",
		Output: "tool execution result",
	}

	b.handleMessage(msg, ch, &output)

	if output.String() != "" {
		t.Fatalf("tool_result should not add to output, got %q", output.String())
	}
	select {
	case m := <-ch:
		if m.Type != MessageToolResult || m.Tool != "exec_command" {
			t.Fatalf("unexpected message: %+v", m)
		}
		if m.Output != "tool execution result" {
			t.Fatalf("expected output 'tool execution result', got %v", m.Output)
		}
	default:
		t.Fatal("expected message on channel")
	}
}

func TestCopilotHandleMessageError(t *testing.T) {
	t.Parallel()

	b := &copilotBackend{cfg: Config{Logger: slog.Default()}}
	ch := make(chan Message, 10)
	var output strings.Builder

	msg := copilotMessage{
		Type:    "error",
		Content: "Something went wrong",
	}

	b.handleMessage(msg, ch, &output)

	if output.String() != "" {
		t.Fatalf("error should not add to output, got %q", output.String())
	}
	select {
	case m := <-ch:
		if m.Type != MessageError || m.Content != "Something went wrong" {
			t.Fatalf("unexpected message: %+v", m)
		}
	default:
		t.Fatal("expected message on channel")
	}
}

func TestCopilotHandleMessageStatus(t *testing.T) {
	t.Parallel()

	b := &copilotBackend{cfg: Config{Logger: slog.Default()}}
	ch := make(chan Message, 10)
	var output strings.Builder

	msg := copilotMessage{
		Type:   "status",
		Status: "running",
	}

	b.handleMessage(msg, ch, &output)

	if output.String() != "" {
		t.Fatalf("status should not add to output, got %q", output.String())
	}
	select {
	case m := <-ch:
		if m.Type != MessageStatus || m.Status != "running" {
			t.Fatalf("unexpected message: %+v", m)
		}
	default:
		t.Fatal("expected message on channel")
	}
}

func TestCopilotHandleMessageEmpty(t *testing.T) {
	t.Parallel()

	b := &copilotBackend{cfg: Config{Logger: slog.Default()}}
	ch := make(chan Message, 10)
	var output strings.Builder

	msg := copilotMessage{
		Type:    "response",
		Content: "",
	}

	b.handleMessage(msg, ch, &output)

	if output.String() != "" {
		t.Fatalf("expected empty output, got %q", output.String())
	}
	select {
	case m := <-ch:
		t.Fatalf("expected no message for empty content, got %+v", m)
	default:
	}
}
