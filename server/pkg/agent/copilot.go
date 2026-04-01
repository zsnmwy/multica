package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// copilotBackend implements Backend by spawning the GitHub Copilot CLI
// with --jsonl for streaming JSON output.
type copilotBackend struct {
	cfg Config
}

func (b *copilotBackend) Execute(ctx context.Context, prompt string, opts ExecOptions) (*Session, error) {
	execPath := b.cfg.ExecutablePath
	if execPath == "" {
		execPath = "copilot"
	}
	if _, err := exec.LookPath(execPath); err != nil {
		return nil, fmt.Errorf("copilot executable not found at %q: %w", execPath, err)
	}

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 20 * time.Minute
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)

	args := []string{"-sp", prompt, "--jsonl"}
	if opts.Model != "" {
		args = append(args, "--model", opts.Model)
	}

	cmd := exec.CommandContext(runCtx, execPath, args...)
	if opts.Cwd != "" {
		cmd.Dir = opts.Cwd
	}
	cmd.Env = buildEnv(b.cfg.Env)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("copilot stdout pipe: %w", err)
	}
	cmd.Stderr = newLogWriter(b.cfg.Logger, "[copilot:stderr] ")

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("start copilot: %w", err)
	}

	b.cfg.Logger.Info("copilot started", "pid", cmd.Process.Pid, "cwd", opts.Cwd, "model", opts.Model)

	msgCh := make(chan Message, 256)
	resCh := make(chan Result, 1)

	go func() {
		defer cancel()
		defer close(msgCh)
		defer close(resCh)

		startTime := time.Now()
		var output strings.Builder
		finalStatus := "completed"
		var finalError string

		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			var msg copilotMessage
			if err := json.Unmarshal([]byte(line), &msg); err != nil {
				continue
			}

			b.handleMessage(msg, msgCh, &output)
		}

		// Wait for process exit
		exitErr := cmd.Wait()
		duration := time.Since(startTime)

		if runCtx.Err() == context.DeadlineExceeded {
			finalStatus = "timeout"
			finalError = fmt.Sprintf("copilot timed out after %s", timeout)
		} else if runCtx.Err() == context.Canceled {
			finalStatus = "aborted"
			finalError = "execution cancelled"
		} else if exitErr != nil && finalStatus == "completed" {
			finalStatus = "failed"
			finalError = fmt.Sprintf("copilot exited with error: %v", exitErr)
		}

		b.cfg.Logger.Info("copilot finished", "pid", cmd.Process.Pid, "status", finalStatus, "duration", duration.Round(time.Millisecond).String())

		resCh <- Result{
			Status:     finalStatus,
			Output:     output.String(),
			Error:      finalError,
			DurationMs: duration.Milliseconds(),
		}
	}()

	return &Session{Messages: msgCh, Result: resCh}, nil
}

func (b *copilotBackend) handleMessage(msg copilotMessage, ch chan<- Message, output *strings.Builder) {
	switch msg.Type {
	case "response", "text":
		if msg.Content != "" {
			output.WriteString(msg.Content)
			trySend(ch, Message{Type: MessageText, Content: msg.Content})
		}
	case "explanation", "thinking":
		if msg.Content != "" {
			trySend(ch, Message{Type: MessageThinking, Content: msg.Content})
		}
	case "command", "tool_use":
		if msg.Command != "" {
			trySend(ch, Message{
				Type:  MessageToolUse,
				Tool:  "exec_command",
				Input: map[string]any{"command": msg.Command},
			})
		}
	case "result", "tool_result":
		if msg.Output != "" {
			trySend(ch, Message{
				Type:   MessageToolResult,
				Tool:   "exec_command",
				Output: msg.Output,
			})
		}
	case "error":
		if msg.Content != "" {
			trySend(ch, Message{Type: MessageError, Content: msg.Content})
		}
	case "status":
		if msg.Status != "" {
			trySend(ch, Message{Type: MessageStatus, Status: msg.Status})
		}
	}
}

// copilotMessage represents a JSON line from Copilot CLI --jsonl output.
type copilotMessage struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Command string `json:"command"`
	Output  string `json:"output"`
	Status  string `json:"status"`
}
