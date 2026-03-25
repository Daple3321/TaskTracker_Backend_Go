export type SelectionEdit = {
  next: string;
  selStart: number;
  selEnd: number;
};

export function wrapSelection(
  value: string,
  start: number,
  end: number,
  before: string,
  after: string
): SelectionEdit {
  const selected = value.slice(start, end);
  const next = value.slice(0, start) + before + selected + after + value.slice(end);
  if (start === end) {
    const cursor = start + before.length;
    return { next, selStart: cursor, selEnd: cursor };
  }
  return {
    next,
    selStart: start + before.length,
    selEnd: start + before.length + selected.length,
  };
}

function lineStart(value: string, pos: number): number {
  let i = pos;
  while (i > 0 && value[i - 1] !== "\n") i--;
  return i;
}

function lineEnd(value: string, pos: number): number {
  let i = pos;
  while (i < value.length && value[i] !== "\n") i++;
  return i;
}

export function prefixHeading(value: string, cursor: number, hashes: string): SelectionEdit {
  const ls = lineStart(value, cursor);
  const le = lineEnd(value, cursor);
  const line = value.slice(ls, le);
  const stripped = line.replace(/^#{1,6}\s/, "");
  const finalLine = stripped.length > 0 ? `${hashes} ${stripped}` : `${hashes} `;
  const next = value.slice(0, ls) + finalLine + value.slice(le);
  const delta = finalLine.length - line.length;
  const newPos = Math.min(Math.max(cursor + delta, ls), ls + finalLine.length);
  return { next, selStart: newPos, selEnd: newPos };
}

export function insertLink(value: string, start: number, end: number): SelectionEdit {
  const selected = value.slice(start, end);
  if (selected.length > 0) {
    const insert = `[${selected}](https://)`;
    const next = value.slice(0, start) + insert + value.slice(end);
    const urlStart = start + `[${selected}](`.length;
    const urlEnd = urlStart + `https://`.length;
    return { next, selStart: urlStart, selEnd: urlEnd };
  }
  const placeholder = `[link text](https://)`;
  const next = value.slice(0, start) + placeholder + value.slice(end);
  const labelStart = start + "[".length;
  const labelEnd = labelStart + "link text".length;
  return { next, selStart: labelStart, selEnd: labelEnd };
}

export function insertCodeBlock(value: string, start: number, end: number): SelectionEdit {
  const selected = value.slice(start, end);
  if (selected.length > 0) {
    const block = `\n\`\`\`\n${selected}\n\`\`\`\n`;
    const next = value.slice(0, start) + block + value.slice(end);
    const innerStart = start + "\n```\n".length;
    const innerEnd = innerStart + selected.length;
    return { next, selStart: innerStart, selEnd: innerEnd };
  }
  const block = `\n\`\`\`\n\n\`\`\`\n`;
  const next = value.slice(0, start) + block + value.slice(end);
  const cursor = start + "\n```\n".length;
  return { next, selStart: cursor, selEnd: cursor };
}
