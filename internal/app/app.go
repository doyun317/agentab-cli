package app

import (
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/agentab/agentab-cli/internal/daemon"
	"github.com/agentab/agentab-cli/internal/install"
	"github.com/agentab/agentab-cli/internal/pinchtab"
	"github.com/agentab/agentab-cli/internal/response"
	"github.com/agentab/agentab-cli/internal/state"
)

type GlobalOptions struct {
	Session string
	Tab     string
	Profile string
	Mode    string
	Owner   string
	Timeout time.Duration
	Output  string
	Debug   bool
}

func Run(ctx context.Context, args []string) int {
	store, err := state.NewStore("")
	if err != nil {
		return print(response.Fail("dependency_error", err.Error(), nil, nil), "json")
	}

	global, rest, err := parseGlobal(args)
	if err != nil {
		return print(response.Fail("usage_error", err.Error(), nil, nil), global.Output)
	}
	if len(rest) == 0 {
		return print(response.Fail("usage_error", usage(), nil, nil), global.Output)
	}

	env := response.Fail("usage_error", "unknown command", nil, nil)
	switch rest[0] {
	case "doctor":
		env = runDoctor(ctx, store)
	case "daemon":
		env = runDaemon(ctx, store, global, rest[1:])
	case "session":
		env = runSession(ctx, store, global, rest[1:])
	case "tab":
		env = runTab(ctx, store, global, rest[1:])
	default:
		env = response.Fail("usage_error", usage(), nil, nil)
	}
	return print(env, global.Output)
}

func parseGlobal(args []string) (GlobalOptions, []string, error) {
	opts := GlobalOptions{Timeout: 30 * time.Second, Output: "json"}
	fs := flag.NewFlagSet("agentab", flag.ContinueOnError)
	fs.SetOutput(ioDiscard{})
	fs.StringVar(&opts.Session, "session", "", "session name")
	fs.StringVar(&opts.Tab, "tab", "", "tab id")
	fs.StringVar(&opts.Profile, "profile", "", "profile id")
	fs.StringVar(&opts.Mode, "mode", "", "mode")
	fs.StringVar(&opts.Owner, "owner", "", "owner")
	fs.DurationVar(&opts.Timeout, "timeout", 30*time.Second, "timeout")
	fs.StringVar(&opts.Output, "output", "json", "json or text")
	fs.BoolVar(&opts.Debug, "debug", false, "debug")
	if err := fs.Parse(args); err != nil {
		return opts, nil, err
	}
	return opts, fs.Args(), nil
}

func runDoctor(ctx context.Context, store *state.Store) response.Envelope {
	inst := install.New(store.Root())
	path, managed, err := inst.ResolveBinaryPath()
	manager := pinchtab.NewManager(store)
	pinchURL := manager.BaseURL()
	ptClient := manager.Client(1500 * time.Millisecond)
	_, healthErr := ptClient.Health(ctx)
	chrome := resolveChromeBinary()

	daemonInfo, daemonErr := store.ReadDaemonInfo()
	report := doctorReport{
		AgentabHome:     store.Root(),
		LogsDir:         store.LogsDir(),
		DaemonLogPath:   store.DaemonLogPath(),
		PinchtabLogPath: store.PinchtabLogPath(),
		ArtifactsDir:    store.ArtifactsDir(),
		ManagedBinPath:  inst.ManagedBinaryPath(),
		PinchtabURL:     pinchURL,
		PinchtabHealthy: healthErr == nil,
		ChromeBin:       chrome.Path,
		ChromeBinFound:  chrome.Found,
		ChromeBinSource: chrome.Source,
		ChromeBinError:  chrome.Error,
	}
	if err == nil {
		report.PinchtabBin = path
		report.PinchtabManaged = boolPtr(managed)
	} else {
		report.PinchtabBinError = err.Error()
	}
	if daemonErr == nil {
		report.Daemon = &daemonInfo
	}
	if info, infoErr := store.ReadPinchtabInfo(); infoErr == nil {
		report.Pinchtab = &info
	}
	return response.OK(report, nil)
}

type chromeBinaryInfo struct {
	Path   string
	Source string
	Found  bool
	Error  string
}

func resolveChromeBinary() chromeBinaryInfo {
	if override := strings.TrimSpace(os.Getenv("CHROME_BIN")); override != "" {
		resolved, err := resolveChromeCandidate(override)
		if err != nil {
			return chromeBinaryInfo{
				Path:   override,
				Source: "env",
				Found:  false,
				Error:  err.Error(),
			}
		}
		return chromeBinaryInfo{
			Path:   resolved,
			Source: "env",
			Found:  true,
		}
	}

	chromeBins := []string{"google-chrome", "chromium", "chromium-browser", "chrome"}
	for _, name := range chromeBins {
		if bin, err := exec.LookPath(name); err == nil {
			return chromeBinaryInfo{
				Path:   bin,
				Source: "path",
				Found:  true,
			}
		}
	}
	return chromeBinaryInfo{Source: "path", Found: false}
}

func resolveChromeCandidate(candidate string) (string, error) {
	if candidate == "" {
		return "", errors.New("empty chrome binary candidate")
	}
	if info, err := os.Stat(candidate); err == nil {
		if info.IsDir() {
			return "", fmt.Errorf("chrome binary path is a directory: %s", candidate)
		}
		abs, absErr := filepath.Abs(candidate)
		if absErr != nil {
			return candidate, nil
		}
		return abs, nil
	}
	if resolved, err := exec.LookPath(candidate); err == nil {
		return resolved, nil
	}
	return "", fmt.Errorf("chrome binary not found: %s", candidate)
}

func runDaemon(ctx context.Context, store *state.Store, global GlobalOptions, args []string) response.Envelope {
	if len(args) == 0 {
		return response.Fail("usage_error", "expected daemon subcommand", nil, nil)
	}
	client := daemon.NewClient(store, global.Timeout)
	switch args[0] {
	case "start":
		if err := client.Ensure(ctx); err != nil {
			return response.Fail("dependency_error", err.Error(), nil, nil)
		}
		status, err := client.Status(ctx)
		if err != nil {
			return response.Fail("dependency_error", err.Error(), nil, nil)
		}
		return status
	case "status":
		env, err := client.Status(ctx)
		if err != nil {
			return env
		}
		return env
	case "stop":
		env, err := client.Stop(ctx)
		if err != nil {
			return env
		}
		return env
	case "serve":
		return runDaemonServe(ctx, store, args[1:])
	default:
		return response.Fail("usage_error", "unknown daemon subcommand", nil, nil)
	}
}

func runDaemonServe(ctx context.Context, store *state.Store, args []string) response.Envelope {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(ioDiscard{})
	port := daemon.DefaultPort
	token := ""
	fs.IntVar(&port, "port", daemon.DefaultPort, "port")
	fs.StringVar(&token, "token", "", "token")
	if err := fs.Parse(args); err != nil {
		return response.Fail("usage_error", err.Error(), nil, nil)
	}
	if token == "" {
		return response.Fail("usage_error", "daemon serve requires --token", nil, nil)
	}

	server := daemon.NewServer(store, pinchtab.NewManager(store), token, port)
	signalCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := server.Serve(signalCtx); err != nil {
		return response.Fail("dependency_error", err.Error(), nil, nil)
	}
	return response.OK(map[string]any{"status": "stopped"}, nil)
}

func runSession(ctx context.Context, store *state.Store, global GlobalOptions, args []string) response.Envelope {
	if len(args) == 0 {
		return response.Fail("usage_error", "expected session subcommand", nil, nil)
	}
	client := daemon.NewClient(store, global.Timeout)
	switch args[0] {
	case "start":
		fs := flag.NewFlagSet("session-start", flag.ContinueOnError)
		fs.SetOutput(ioDiscard{})
		profile := global.Profile
		mode := global.Mode
		fs.StringVar(&profile, "profile", profile, "profile")
		fs.StringVar(&mode, "mode", mode, "mode")
		if err := fs.Parse(args[1:]); err != nil {
			return response.Fail("usage_error", err.Error(), nil, nil)
		}
		name := ""
		if rest := fs.Args(); len(rest) > 0 {
			name = rest[0]
		}
		env, _ := client.Request(ctx, "POST", "/sessions/start", map[string]any{
			"name":      name,
			"profileId": profile,
			"mode":      mode,
		})
		return env
	case "list":
		env, _ := client.Request(ctx, "GET", "/sessions", nil)
		return env
	case "resume":
		if len(args) < 2 {
			return response.Fail("usage_error", "session resume requires a name", nil, nil)
		}
		env, _ := client.Request(ctx, "POST", "/sessions/"+args[1]+"/resume", nil)
		return env
	case "stop":
		name := global.Session
		if len(args) > 1 {
			name = args[1]
		}
		if name == "" {
			current, err := store.CurrentSession()
			if err != nil {
				return response.Fail("not_found", "no current session", nil, nil)
			}
			name = current.Name
		}
		env, _ := client.Request(ctx, "POST", "/sessions/"+name+"/stop", nil)
		return env
	default:
		return response.Fail("usage_error", "unknown session subcommand", nil, nil)
	}
}

func runTab(ctx context.Context, store *state.Store, global GlobalOptions, args []string) response.Envelope {
	if len(args) == 0 {
		return response.Fail("usage_error", "expected tab subcommand", nil, nil)
	}
	client := daemon.NewClient(store, global.Timeout)
	switch args[0] {
	case "open":
		fs := flag.NewFlagSet("tab-open", flag.ContinueOnError)
		fs.SetOutput(ioDiscard{})
		sessionName := global.Session
		fs.StringVar(&sessionName, "session", sessionName, "session")
		if err := fs.Parse(args[1:]); err != nil {
			return response.Fail("usage_error", err.Error(), nil, nil)
		}
		sessionName, err := resolveSession(store, sessionName)
		if err != nil {
			return response.Fail("not_found", err.Error(), nil, nil)
		}
		rawURL := "about:blank"
		if rest := fs.Args(); len(rest) > 0 {
			rawURL = rest[0]
		}
		env, _ := client.Request(ctx, "POST", "/sessions/"+sessionName+"/tabs/open", map[string]any{"url": rawURL})
		return env
	case "list":
		fs := flag.NewFlagSet("tab-list", flag.ContinueOnError)
		fs.SetOutput(ioDiscard{})
		sessionName := global.Session
		fs.StringVar(&sessionName, "session", sessionName, "session")
		if err := fs.Parse(args[1:]); err != nil {
			return response.Fail("usage_error", err.Error(), nil, nil)
		}
		sessionName, err := resolveSession(store, sessionName)
		if err != nil {
			return response.Fail("not_found", err.Error(), nil, nil)
		}
		env, _ := client.Request(ctx, "GET", "/sessions/"+sessionName+"/tabs", nil)
		return env
	case "close":
		fs := flag.NewFlagSet("tab-close", flag.ContinueOnError)
		fs.SetOutput(ioDiscard{})
		tabID := global.Tab
		fs.StringVar(&tabID, "tab", tabID, "tab")
		if err := fs.Parse(args[1:]); err != nil {
			return response.Fail("usage_error", err.Error(), nil, nil)
		}
		tabID, err := resolveTab(store, global.Session, tabID)
		if err != nil {
			return response.Fail("not_found", err.Error(), nil, nil)
		}
		env, _ := client.Request(ctx, "POST", "/tabs/"+tabID+"/close", nil)
		return env
	case "focus":
		fs := flag.NewFlagSet("tab-focus", flag.ContinueOnError)
		fs.SetOutput(ioDiscard{})
		tabID := global.Tab
		sessionName := global.Session
		fs.StringVar(&tabID, "tab", tabID, "tab")
		fs.StringVar(&sessionName, "session", sessionName, "session")
		if err := fs.Parse(args[1:]); err != nil {
			return response.Fail("usage_error", err.Error(), nil, nil)
		}
		sessionName, err := resolveSession(store, sessionName)
		if err != nil {
			return response.Fail("not_found", err.Error(), nil, nil)
		}
		if tabID == "" && len(fs.Args()) > 0 {
			tabID = fs.Args()[0]
		}
		if tabID == "" {
			return response.Fail("usage_error", "tab focus requires --tab or a tab id argument", nil, nil)
		}
		env, _ := client.Request(ctx, "POST", "/tabs/"+tabID+"/focus?session="+sessionName, nil)
		return env
	case "snapshot":
		fs := flag.NewFlagSet("tab-snapshot", flag.ContinueOnError)
		fs.SetOutput(ioDiscard{})
		tabID := global.Tab
		filter := "interactive"
		format := "compact"
		selector := ""
		maxTokens := ""
		depth := ""
		diff := false
		outPath := ""
		save := false
		sessionName := global.Session
		fs.StringVar(&tabID, "tab", tabID, "tab")
		fs.StringVar(&sessionName, "session", sessionName, "session")
		fs.StringVar(&filter, "filter", filter, "filter")
		fs.StringVar(&format, "format", format, "format")
		fs.StringVar(&selector, "selector", selector, "selector")
		fs.StringVar(&maxTokens, "max-tokens", maxTokens, "max tokens")
		fs.StringVar(&depth, "depth", depth, "depth")
		fs.StringVar(&outPath, "out", outPath, "output path")
		fs.BoolVar(&save, "save", save, "save artifact to the managed artifacts directory")
		fs.BoolVar(&diff, "diff", diff, "diff")
		if err := fs.Parse(args[1:]); err != nil {
			return response.Fail("usage_error", err.Error(), nil, nil)
		}
		tabID, err := resolveTab(store, sessionName, tabID)
		if err != nil {
			return response.Fail("not_found", err.Error(), nil, nil)
		}
		path := fmt.Sprintf("/tabs/%s/snapshot?filter=%s&format=%s", tabID, filter, format)
		if selector != "" {
			path += "&selector=" + selector
		}
		if maxTokens != "" {
			path += "&maxTokens=" + maxTokens
		}
		if depth != "" {
			path += "&depth=" + depth
		}
		if diff {
			path += "&diff=true"
		}
		env, _ := client.Request(ctx, "GET", path, nil)
		if !env.OK || (outPath == "" && !save) {
			return env
		}
		meta, err := saveSnapshotArtifact(
			store,
			artifactSessionName(store, sessionName),
			tabID,
			outPath,
			env.Data,
			time.Now(),
		)
		if err != nil {
			return response.Fail("dependency_error", err.Error(), nil, nil)
		}
		return response.OK(meta, nil)
	case "text":
		fs := flag.NewFlagSet("tab-text", flag.ContinueOnError)
		fs.SetOutput(ioDiscard{})
		tabID := global.Tab
		mode := "readability"
		sessionName := global.Session
		fs.StringVar(&tabID, "tab", tabID, "tab")
		fs.StringVar(&sessionName, "session", sessionName, "session")
		fs.StringVar(&mode, "mode", mode, "mode")
		if err := fs.Parse(args[1:]); err != nil {
			return response.Fail("usage_error", err.Error(), nil, nil)
		}
		tabID, err := resolveTab(store, sessionName, tabID)
		if err != nil {
			return response.Fail("not_found", err.Error(), nil, nil)
		}
		env, _ := client.Request(ctx, "GET", "/tabs/"+tabID+"/text?mode="+mode, nil)
		return env
	case "find":
		fs := flag.NewFlagSet("tab-find", flag.ContinueOnError)
		fs.SetOutput(ioDiscard{})
		tabID := global.Tab
		threshold := ""
		explain := false
		sessionName := global.Session
		fs.StringVar(&tabID, "tab", tabID, "tab")
		fs.StringVar(&sessionName, "session", sessionName, "session")
		fs.StringVar(&threshold, "threshold", threshold, "threshold")
		fs.BoolVar(&explain, "explain", explain, "explain")
		if err := fs.Parse(args[1:]); err != nil {
			return response.Fail("usage_error", err.Error(), nil, nil)
		}
		if len(fs.Args()) == 0 {
			return response.Fail("usage_error", "tab find requires a query", nil, nil)
		}
		tabID, err := resolveTab(store, sessionName, tabID)
		if err != nil {
			return response.Fail("not_found", err.Error(), nil, nil)
		}
		env, _ := client.Request(ctx, "POST", "/tabs/"+tabID+"/find", map[string]any{
			"query":     strings.Join(fs.Args(), " "),
			"threshold": threshold,
			"explain":   explain,
		})
		return env
	case "click", "type", "fill", "press", "hover", "scroll", "select":
		return runAction(ctx, store, client, global, args)
	case "eval":
		fs := flag.NewFlagSet("tab-eval", flag.ContinueOnError)
		fs.SetOutput(ioDiscard{})
		tabID := global.Tab
		sessionName := global.Session
		fs.StringVar(&tabID, "tab", tabID, "tab")
		fs.StringVar(&sessionName, "session", sessionName, "session")
		if err := fs.Parse(args[1:]); err != nil {
			return response.Fail("usage_error", err.Error(), nil, nil)
		}
		if len(fs.Args()) == 0 {
			return response.Fail("usage_error", "tab eval requires an expression", nil, nil)
		}
		tabID, err := resolveTab(store, sessionName, tabID)
		if err != nil {
			return response.Fail("not_found", err.Error(), nil, nil)
		}
		env, _ := client.Request(ctx, "POST", "/tabs/"+tabID+"/evaluate", map[string]any{
			"expression": strings.Join(fs.Args(), " "),
		})
		return env
	case "screenshot":
		return runBinaryFetch(ctx, store, client, global, args, "image/jpeg")
	case "pdf":
		return runBinaryFetch(ctx, store, client, global, args, "application/pdf")
	default:
		return response.Fail("usage_error", "unknown tab subcommand", nil, nil)
	}
}

func runAction(ctx context.Context, store *state.Store, client *daemon.Client, global GlobalOptions, args []string) response.Envelope {
	fs := flag.NewFlagSet("tab-action", flag.ContinueOnError)
	fs.SetOutput(ioDiscard{})
	sessionName := global.Session
	tabID := global.Tab
	ref := ""
	selector := ""
	waitNav := false
	fs.StringVar(&sessionName, "session", sessionName, "session")
	fs.StringVar(&tabID, "tab", tabID, "tab")
	fs.StringVar(&ref, "ref", ref, "ref")
	fs.StringVar(&selector, "selector", selector, "selector")
	fs.BoolVar(&waitNav, "wait-nav", waitNav, "wait navigation")
	if err := fs.Parse(args[1:]); err != nil {
		return response.Fail("usage_error", err.Error(), nil, nil)
	}
	tabID, err := resolveTab(store, sessionName, tabID)
	if err != nil {
		return response.Fail("not_found", err.Error(), nil, nil)
	}

	body := map[string]any{"kind": args[0], "owner": global.Owner}
	if selector != "" {
		body["selector"] = selector
	}
	if ref != "" {
		body["ref"] = ref
	}
	switch args[0] {
	case "click", "hover":
		if ref == "" && selector == "" && len(fs.Args()) > 0 {
			body["ref"] = fs.Args()[0]
		}
	case "type":
		if ref == "" && len(fs.Args()) > 0 {
			body["ref"] = fs.Args()[0]
		}
		if len(fs.Args()) > 1 {
			body["text"] = strings.Join(fs.Args()[1:], " ")
		}
	case "fill":
		if selector == "" && ref == "" && len(fs.Args()) > 0 {
			body["selector"] = fs.Args()[0]
		}
		if len(fs.Args()) > 1 {
			body["text"] = strings.Join(fs.Args()[1:], " ")
		}
	case "press":
		if len(fs.Args()) > 0 {
			body["key"] = fs.Args()[0]
		}
	case "scroll":
		if len(fs.Args()) > 0 {
			if strings.HasPrefix(fs.Args()[0], "e") {
				body["ref"] = fs.Args()[0]
			} else {
				body["scrollY"] = atoiDefault(fs.Args()[0], 800)
			}
		}
	case "select":
		if ref == "" && len(fs.Args()) > 0 {
			body["ref"] = fs.Args()[0]
		}
		if len(fs.Args()) > 1 {
			body["value"] = fs.Args()[1]
		}
	}
	if waitNav {
		body["waitNav"] = true
	}
	env, _ := client.Request(ctx, "POST", "/tabs/"+tabID+"/action", body)
	return env
}

func runBinaryFetch(ctx context.Context, store *state.Store, client *daemon.Client, global GlobalOptions, args []string, mime string) response.Envelope {
	fs := flag.NewFlagSet("tab-binary", flag.ContinueOnError)
	fs.SetOutput(ioDiscard{})
	sessionName := global.Session
	tabID := global.Tab
	outPath := ""
	save := false
	quality := ""
	scale := ""
	landscape := false
	fs.StringVar(&sessionName, "session", sessionName, "session")
	fs.StringVar(&tabID, "tab", tabID, "tab")
	fs.StringVar(&outPath, "out", "", "output path")
	fs.BoolVar(&save, "save", false, "save artifact to the managed artifacts directory")
	fs.StringVar(&quality, "quality", "", "quality")
	fs.StringVar(&scale, "scale", "", "scale")
	fs.BoolVar(&landscape, "landscape", false, "landscape")
	if err := fs.Parse(args[1:]); err != nil {
		return response.Fail("usage_error", err.Error(), nil, nil)
	}
	tabID, err := resolveTab(store, sessionName, tabID)
	if err != nil {
		return response.Fail("not_found", err.Error(), nil, nil)
	}
	path := "/tabs/" + tabID
	if mime == "image/jpeg" {
		path += "/screenshot"
		if quality != "" {
			path += "?quality=" + quality
		}
	} else {
		path += "/pdf"
		query := []string{}
		if scale != "" {
			query = append(query, "scale="+scale)
		}
		if landscape {
			query = append(query, "landscape=true")
		}
		if len(query) > 0 {
			path += "?" + strings.Join(query, "&")
		}
	}
	env, _ := client.Request(ctx, "GET", path, nil)
	if !env.OK || (outPath == "" && !save) {
		return env
	}

	data, err := decodeBinaryEnvelope(env, mime)
	if err != nil {
		return response.Fail("upstream_error", err.Error(), nil, nil)
	}
	kind := "pdf"
	ext := "pdf"
	if mime == "image/jpeg" {
		kind = "screenshot"
		ext = "jpg"
	}
	meta, err := saveBinaryArtifact(
		store,
		artifactSessionName(store, sessionName),
		tabID,
		kind,
		outPath,
		ext,
		mime,
		data,
		time.Now(),
	)
	if err != nil {
		return response.Fail("dependency_error", err.Error(), nil, nil)
	}
	return response.OK(meta, nil)
}

func resolveSession(store *state.Store, sessionName string) (string, error) {
	if sessionName != "" {
		return sessionName, nil
	}
	current, err := store.CurrentSession()
	if err != nil {
		return "", errors.New("no current session; pass --session")
	}
	return current.Name, nil
}

func resolveTab(store *state.Store, sessionName, tabID string) (string, error) {
	if tabID != "" {
		return tabID, nil
	}
	sessionName, err := resolveSession(store, sessionName)
	if err != nil {
		return "", err
	}
	session, err := store.GetSession(sessionName)
	if err != nil {
		return "", err
	}
	if session.CurrentTabID == "" {
		return "", errors.New("no current tab; pass --tab or open/focus a tab first")
	}
	return session.CurrentTabID, nil
}

func decodeBinaryEnvelope(env response.Envelope, mime string) ([]byte, error) {
	payload, ok := env.Data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected map payload")
	}
	rawData, ok := payload["data"].(string)
	if !ok {
		return nil, fmt.Errorf("binary data missing")
	}
	return base64.StdEncoding.DecodeString(rawData)
}

func atoiDefault(raw string, fallback int) int {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func print(env response.Envelope, format string) int {
	if err := response.Print(env, format); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 7
	}
	return response.ExitCode(env)
}

func usage() string {
	return "usage: agentab [--output json|text] <doctor|daemon|session|tab> ..."
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }
