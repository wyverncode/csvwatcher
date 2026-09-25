export function parseCsv(input) {
  const rows = [];
  let row = [];
  let field = "";
  let quoted = false;
  for (let index = 0; index < input.length; index += 1) {
    const character = input[index];
    if (quoted) {
      if (character === '"' && input[index + 1] === '"') {
        field += '"';
        index += 1;
      } else if (character === '"') {
        quoted = false;
      } else {
        field += character;
      }
    } else if (character === '"') {
      quoted = true;
    } else if (character === ",") {
      row.push(field);
      field = "";
    } else if (character === "\n" || character === "\r") {
      if (character === "\r" && input[index + 1] === "\n") index += 1;
      row.push(field);
      if (row.some((value) => value !== "")) rows.push(row);
      row = [];
      field = "";
    } else {
      field += character;
    }
  }
  if (quoted) throw new Error("unterminated quoted field");
  if (field || row.length) {
    row.push(field);
    rows.push(row);
  }
  if (rows.length === 0) return [];
  const headers = rows[0];
  const seen = new Set();
  headers.forEach((header, index) => {
    if (!header) throw new Error(`header ${index + 1} is empty`);
    if (seen.has(header)) throw new Error(`duplicate header "${header}"`);
    seen.add(header);
  });
  return rows.slice(1).map((values, rowIndex) => {
    if (values.length !== headers.length) {
      throw new Error(
        `row ${rowIndex + 2} has ${values.length} fields; expected ${headers.length}`,
      );
    }
    return Object.fromEntries(
      headers.map((header, index) => [header, values[index]]),
    );
  });
}
