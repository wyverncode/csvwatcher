import { promises as fs } from "node:fs";
import path from "node:path";
import { parseCsv } from "./csv.js";

export async function convertFile(sourcePath, outputDir) {
  const content = await fs.readFile(sourcePath, "utf8");
  const records = parseCsv(content);
  const outputPath = path.join(
    outputDir,
    `${path.basename(sourcePath, path.extname(sourcePath))}.json`,
  );
  const temporaryPath = path.join(
    outputDir,
    `.csvwatcher-${process.pid}-${Date.now()}.tmp`,
  );
  await fs.writeFile(
    temporaryPath,
    `${JSON.stringify(records, null, 2)}\n`,
    "utf8",
  );
  await fs.rename(temporaryPath, outputPath);
  return { outputPath, rows: records.length };
}

export async function waitForStableFile(filePath, timeoutMs, signal) {
  const deadline = Date.now() + timeoutMs;
  let previous;
  while (Date.now() < deadline) {
    if (signal?.aborted) throw new Error("operation cancelled");
    const stats = await fs.stat(filePath);
    const current = `${stats.size}:${stats.mtimeMs}`;
    if (current === previous) return;
    previous = current;
    await new Promise((resolve, reject) => {
      const timer = setTimeout(resolve, 100);
      signal?.addEventListener(
        "abort",
        () => {
          clearTimeout(timer);
          reject(new Error("operation cancelled"));
        },
        { once: true },
      );
    });
  }
  throw new Error(`file did not become stable within ${timeoutMs}ms`);
}
