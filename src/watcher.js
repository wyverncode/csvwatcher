import { promises as fs } from "node:fs";
import path from "node:path";
import { convertFile, waitForStableFile } from "./converter.js";

export function isCsv(fileName) {
  return path.extname(fileName).toLowerCase() === ".csv";
}

export function directoriesOverlap(first, second) {
  const a = path.resolve(first);
  const b = path.resolve(second);
  return (
    b === a ||
    b.startsWith(`${a}${path.sep}`) ||
    a.startsWith(`${b}${path.sep}`)
  );
}

export async function processDirectory(inputDir, outputDir, options) {
  const entries = await fs.readdir(inputDir, { withFileTypes: true });
  const results = [];
  for (const entry of entries) {
    if (entry.isFile() && isCsv(entry.name)) {
      results.push(
        await processFile(path.join(inputDir, entry.name), outputDir, options),
      );
    }
  }
  return results;
}

export async function processFile(sourcePath, outputDir, options) {
  try {
    await waitForStableFile(
      sourcePath,
      options.stabilityTimeoutMs,
      options.signal,
    );
    const result = await convertFile(sourcePath, outputDir);
    options.log(
      `converted ${sourcePath} -> ${result.outputPath} (${result.rows} rows)`,
    );
    return true;
  } catch (error) {
    options.log(`conversion failed for ${sourcePath}: ${error.message}`);
    return false;
  }
}

export async function watch(inputDir, outputDir, options) {
  const known = new Map();
  const scan = async () => {
    const entries = await fs.readdir(inputDir, { withFileTypes: true });
    for (const entry of entries) {
      if (!entry.isFile() || !isCsv(entry.name)) continue;
      const filePath = path.join(inputDir, entry.name);
      const stats = await fs.stat(filePath);
      const state = `${stats.size}:${stats.mtimeMs}`;
      if (known.get(filePath) === state) continue;
      if (await processFile(filePath, outputDir, options))
        known.set(filePath, state);
    }
  };
  options.log(`watching ${inputDir}; output directory is ${outputDir}`);
  await scan();
  while (!options.signal?.aborted) {
    await new Promise((resolve) => setTimeout(resolve, options.intervalMs));
    await scan();
  }
}
