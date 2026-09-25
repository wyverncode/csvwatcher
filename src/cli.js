import { promises as fs } from "node:fs";
import path from "node:path";
import { processDirectory, directoriesOverlap, watch } from "./watcher.js";
import { startGui } from "./web.js";

const defaults = { intervalMs: 1000, stabilityTimeoutMs: 10000 };

export async function run(args) {
  const options = parseArgs(args);
  if (options.gui) return startGui();
  if (!options.input || !options.output)
    throw new Error("both --input and --output are required");
  const input = path.resolve(options.input);
  const output = path.resolve(options.output);
  const inputStats = await fs.stat(input);
  if (!inputStats.isDirectory())
    throw new Error(`input path is not a directory: ${input}`);
  if (directoriesOverlap(input, output))
    throw new Error("input and output directories must not overlap");
  await fs.mkdir(output, { recursive: true });
  const controller = new AbortController();
  process.once("SIGINT", () => controller.abort());
  process.once("SIGTERM", () => controller.abort());
  const watchOptions = {
    ...options,
    signal: controller.signal,
    log: (message) => console.log(`${new Date().toISOString()} ${message}`),
  };
  if (options.once) return processDirectory(input, output, watchOptions);
  return watch(input, output, watchOptions);
}

function parseArgs(args) {
  const options = { ...defaults };
  for (let index = 0; index < args.length; index += 1) {
    const argument = args[index];
    if (argument === "--help" || argument === "-h") {
      console.log(
        "Usage: csvwatcher --input DIR --output DIR [--once] [--interval 1s] [--stability-timeout 10s]",
      );
      console.log("       csvwatcher --gui");
      process.exit(0);
    }
    if (argument === "--gui") options.gui = true;
    else if (argument === "--once") options.once = true;
    else if (argument === "--input") options.input = args[++index];
    else if (argument === "--output") options.output = args[++index];
    else if (argument === "--interval")
      options.intervalMs = parseDuration(args[++index]);
    else if (argument === "--stability-timeout")
      options.stabilityTimeoutMs = parseDuration(args[++index]);
    else throw new Error(`unknown option: ${argument}`);
  }
  if (options.intervalMs <= 0 || options.stabilityTimeoutMs <= 0)
    throw new Error("durations must be greater than zero");
  return options;
}

function parseDuration(value) {
  const match = /^(\d+(?:\.\d+)?)(ms|s|m)?$/.exec(value ?? "");
  if (!match) throw new Error(`invalid duration: ${value}`);
  const amount = Number(match[1]);
  return amount * { ms: 1, s: 1000, m: 60000 }[match[2] ?? "ms"];
}
