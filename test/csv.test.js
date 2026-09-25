import test from "node:test";
import assert from "node:assert/strict";
import { parseCsv } from "../src/csv.js";

test("parses headers, quoted values, and special characters", () => {
  assert.deepEqual(parseCsv('name,note\nAda,"hello, ""world"" & goodbye"\n'), [
    { name: "Ada", note: 'hello, "world" & goodbye' },
  ]);
});

test("rejects rows with a different field count", () => {
  assert.throws(() => parseCsv("name,age\nAda\n"), /row 2 has 1 fields/);
});

test("rejects duplicate headers", () => {
  assert.throws(() => parseCsv("name,name\nAda,Ada\n"), /duplicate header/);
});
