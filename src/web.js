import http from "node:http";
import { promises as fs } from "node:fs";
import path from "node:path";
import { watch, directoriesOverlap } from "./watcher.js";

export function startGui() {
  const state = { controller: null, input: "", output: "", logs: [] };
  const server = http.createServer(async (request, response) => {
    if (request.url === "/" && request.method === "GET") {
      response.writeHead(200, { "content-type": "text/html; charset=utf-8" });
      response.end(page);
      return;
    }
    if (request.url === "/api/status" && request.method === "GET") {
      sendJson(response, {
        status: state.controller ? "watching" : "stopped",
        ...state,
      });
      return;
    }
    if (request.url === "/api/start" && request.method === "POST") {
      const body = await readBody(request);
      const input = path.resolve(body.input ?? "");
      const output = path.resolve(body.output ?? "");
      if (!body.input || !body.output || directoriesOverlap(input, output)) {
        response.writeHead(400);
        response.end(
          "Valid, non-overlapping input and output folders are required",
        );
        return;
      }
      const stats = await fs.stat(input).catch(() => null);
      if (!stats?.isDirectory()) {
        response.writeHead(400);
        response.end("Input folder is not available");
        return;
      }
      await fs.mkdir(output, { recursive: true });
      if (state.controller) {
        response.writeHead(409);
        response.end("Watcher is already running");
        return;
      }
      state.input = input;
      state.output = output;
      state.controller = new AbortController();
      const options = {
        signal: state.controller.signal,
        intervalMs: 1000,
        stabilityTimeoutMs: 10000,
        log: (message) =>
          state.logs.push(`${new Date().toISOString()} ${message}`),
      };
      watch(input, output, options)
        .catch((error) => options.log(`watch failed: ${error.message}`))
        .finally(() => {
          state.controller = null;
        });
      sendJson(response, { status: "watching" });
      return;
    }
    if (request.url === "/api/stop" && request.method === "POST") {
      state.controller?.abort();
      state.controller = null;
      sendJson(response, { status: "stopped" });
      return;
    }
    response.writeHead(404);
    response.end("Not found");
  });
  server.listen(8080, "127.0.0.1", () =>
    console.log("GUI available at http://127.0.0.1:8080"),
  );
  return server;
}

function sendJson(response, data) {
  response.writeHead(200, { "content-type": "application/json" });
  response.end(JSON.stringify(data));
}

async function readBody(request) {
  let body = "";
  for await (const chunk of request) body += chunk;
  return JSON.parse(body || "{}");
}

const page = `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width"><title>CSV Watcher</title><style>body{font:16px system-ui;max-width:800px;margin:2rem auto;padding:0 1rem}label{display:block;margin-top:1rem;font-weight:600}input{box-sizing:border-box;padding:.6rem;width:100%}button{margin:.75rem .5rem .75rem 0;padding:.6rem 1rem}pre{background:#111;color:#eee;min-height:220px;padding:1rem;white-space:pre-wrap}</style></head><body><h1>CSV Watcher</h1><label>Input folder<input id="input"></label><label>Output folder<input id="output"></label><button onclick="start()">Start watching</button><button onclick="stop()">Stop</button><strong id="status">stopped</strong><h2>Activity</h2><pre id="logs"></pre><script>async function refresh(){const d=await(await fetch('/api/status')).json();status.textContent=d.status;logs.textContent=(d.logs||[]).join('\\n')}async function start(){const r=await fetch('/api/start',{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify({input:input.value,output:output.value})});if(!r.ok)alert(await r.text());refresh()}async function stop(){await fetch('/api/stop',{method:'POST'});refresh()}setInterval(refresh,1000);refresh()</script></body></html>`;
