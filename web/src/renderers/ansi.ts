export interface AnsiSegment {
  text: string;
  className?: string;
}

/* ANSI necessarily uses ASCII control bytes in these bounded expressions. */
/* eslint-disable no-control-regex */

const colors: Record<number, string> = {
  30: "black",
  31: "red",
  32: "green",
  33: "yellow",
  34: "blue",
  35: "magenta",
  36: "cyan",
  37: "white",
};

/** A deliberately small ANSI projection. OSC links and unsupported controls are discarded. */
export function parseAnsi(value: string): AnsiSegment[] {
  const clean = value
    .replace(/\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)/g, "")
    .replace(/[\x00-\x08\x0b\x0c\x0e-\x1a\x1c-\x1f\x7f]/g, "");
  const result: AnsiSegment[] = [];
  let position = 0;
  let className: string | undefined;
  const pattern = /\x1b\[([0-9;]*)m/g;
  for (
    let match = pattern.exec(clean);
    match !== null;
    match = pattern.exec(clean)
  ) {
    if (match.index > position)
      result.push({ text: clean.slice(position, match.index), className });
    for (const raw of (match[1] || "0").split(";")) {
      const code = Number(raw);
      if (code === 0 || code === 39) className = undefined;
      else if (colors[code]) className = `ansi-${colors[code]}`;
    }
    position = pattern.lastIndex;
  }
  const tail = clean
    .slice(position)
    .replace(/\x1b(?:\[[0-?]*[ -/]*[@-~]|.)/g, "");
  if (tail) result.push({ text: tail, className });
  return result;
}
