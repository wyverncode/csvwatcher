package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type guiServer struct {
	mu     sync.Mutex
	input  string
	output string
	cancel context.CancelFunc
	logs   []string
}

func runGUI() {
	server := &guiServer{}
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.index)
	mux.HandleFunc("/api/start", server.start)
	mux.HandleFunc("/api/stop", server.stop)
	mux.HandleFunc("/api/status", server.status)
	address := "127.0.0.1:8080"
	log.Printf("GUI available at http://%s", address)
	if err := http.ListenAndServe(address, mux); err != nil {
		log.Fatal(err)
	}
}

func (server *guiServer) index(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = guiTemplate.Execute(response, nil)
}

func (server *guiServer) start(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(response, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	input := strings.TrimSpace(request.FormValue("input"))
	output := strings.TrimSpace(request.FormValue("output"))
	input, output, err := validateGUIPaths(input, output)
	if err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}
	server.mu.Lock()
	if server.cancel != nil {
		server.mu.Unlock()
		http.Error(response, "watcher is already running", http.StatusConflict)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	server.cancel = cancel
	server.input = input
	server.output = output
	server.addLog("starting watcher")
	server.mu.Unlock()

	go func() {
		if err := processExisting(ctx, input, output, defaultStabilityTimeout); err != nil && ctx.Err() == nil {
			server.addLog("startup failed: " + err.Error())
		}
		if ctx.Err() == nil {
			_ = watch(ctx, input, output, defaultWatchInterval, defaultStabilityTimeout)
		}
		server.mu.Lock()
		server.cancel = nil
		server.mu.Unlock()
		server.addLog("watcher stopped")
	}()
	server.status(response, request)
}

func (server *guiServer) stop(response http.ResponseWriter, request *http.Request) {
	server.mu.Lock()
	if server.cancel != nil {
		server.cancel()
		server.addLog("stopping watcher")
	}
	server.mu.Unlock()
	server.status(response, request)
}

func (server *guiServer) status(response http.ResponseWriter, request *http.Request) {
	server.mu.Lock()
	defer server.mu.Unlock()
	response.Header().Set("Content-Type", "application/json")
	status := "stopped"
	if server.cancel != nil {
		status = "watching"
	}
	fmt.Fprintf(response, `{"status":%q,"input":%q,"output":%q,"logs":%s}`,
		status, server.input, server.output, marshalLogs(server.logs))
}

func (server *guiServer) addLog(message string) {
	server.mu.Lock()
	defer server.mu.Unlock()
	server.logs = append(server.logs, time.Now().Format("15:04:05")+" "+message)
	if len(server.logs) > 100 {
		server.logs = server.logs[len(server.logs)-100:]
	}
}

func marshalLogs(logs []string) string {
	data := "["
	for i, entry := range logs {
		if i > 0 {
			data += ","
		}
		data += fmt.Sprintf("%q", entry)
	}
	return data + "]"
}

func validateGUIPaths(input, output string) (string, string, error) {
	if input == "" || output == "" {
		return "", "", fmt.Errorf("input and output folders are required")
	}
	inputPath, err := filepath.Abs(input)
	if err != nil {
		return "", "", fmt.Errorf("resolve input folder: %w", err)
	}
	outputPath, err := filepath.Abs(output)
	if err != nil {
		return "", "", fmt.Errorf("resolve output folder: %w", err)
	}
	info, err := os.Stat(inputPath)
	if err != nil || !info.IsDir() {
		return "", "", fmt.Errorf("input folder is not available")
	}
	if directoriesOverlap(inputPath, outputPath) {
		return "", "", fmt.Errorf("input and output folders must not overlap")
	}
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return "", "", fmt.Errorf("create output folder: %w", err)
	}
	return inputPath, outputPath, nil
}

var guiTemplate = template.Must(template.New("index").Parse(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width">
<title>CSV Watcher</title><style>
body{font:16px system-ui;margin:2rem auto;max-width:800px;padding:0 1rem}
label{display:block;margin-top:1rem;font-weight:600}input{box-sizing:border-box;padding:.6rem;width:100%}
button{margin:.75rem .5rem .75rem 0;padding:.6rem 1rem}#logs{background:#111;color:#eee;
min-height:220px;padding:1rem;white-space:pre-wrap;overflow:auto}
</style></head><body><h1>CSV Watcher</h1>
<p>Choose folders, then start continuous conversion.</p>
<label>Input folder <input id="input" placeholder="C:\data\incoming"></label>
<label>Output folder <input id="output" placeholder="C:\data\converted"></label>
<button onclick="start()">Start watching</button><button onclick="stop()">Stop</button>
<strong id="status">stopped</strong><h2>Activity</h2><pre id="logs"></pre>
<script>
async function call(path, options){let r=await fetch(path,options);if(!r.ok) alert(await r.text());refresh()}
function start(){call('/api/start',{method:'POST',body:new URLSearchParams({input:input.value,output:output.value})})}
function stop(){call('/api/stop',{method:'POST'})}
async function refresh(){let d=await (await fetch('/api/status')).json();status.textContent=d.status;logs.textContent=d.logs.join('\n')}
setInterval(refresh,1000);refresh()
</script></body></html>`))
