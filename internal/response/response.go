package response

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type Envelope struct {
	OK          bool           `json:"ok"`
	Data        any            `json:"data,omitempty"`
	Error       *ErrorBody     `json:"error,omitempty"`
	Diagnostics map[string]any `json:"diagnostics,omitempty"`
}

type TextRenderer interface {
	RenderText() string
}

func OK(data any, diagnostics map[string]any) Envelope {
	return Envelope{OK: true, Data: data, Diagnostics: diagnostics}
}

func Fail(code, message string, details any, diagnostics map[string]any) Envelope {
	return Envelope{
		OK:          false,
		Error:       &ErrorBody{Code: code, Message: message, Details: details},
		Diagnostics: diagnostics,
	}
}

func WriteJSON(w http.ResponseWriter, status int, env Envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(env)
}

func Print(env Envelope, format string) error {
	if format == "text" {
		return printText(os.Stdout, os.Stderr, env)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(env)
}

func printText(stdout, stderr io.Writer, env Envelope) error {
	if env.OK {
		if renderer, ok := env.Data.(TextRenderer); ok {
			_, err := fmt.Fprintln(stdout, renderer.RenderText())
			return err
		}
		if text, ok := env.Data.(string); ok {
			_, err := fmt.Fprintln(stdout, text)
			return err
		}
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(env.Data)
	}

	if env.Error == nil {
		_, err := fmt.Fprintln(stderr, "unknown error")
		return err
	}
	if _, err := fmt.Fprintf(stderr, "%s: %s\n", env.Error.Code, env.Error.Message); err != nil {
		return err
	}
	if env.Error.Details != nil {
		enc := json.NewEncoder(stderr)
		enc.SetIndent("", "  ")
		return enc.Encode(env.Error.Details)
	}
	return nil
}

func ExitCode(env Envelope) int {
	if env.OK || env.Error == nil {
		return 0
	}

	switch env.Error.Code {
	case "usage_error":
		return 2
	case "dependency_error":
		return 3
	case "not_found":
		return 4
	case "lock_conflict":
		return 5
	case "timeout":
		return 6
	default:
		return 7
	}
}
