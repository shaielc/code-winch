import { parseAnsi } from "./ansi";

export function Terminal({ output }: { output: string }) {
  return (
    <pre
      className="terminal"
      aria-label="Terminal output"
      aria-live="polite"
      aria-atomic="false"
    >
      {parseAnsi(output).map((segment, index) => (
        <span className={segment.className} key={index}>
          {segment.text}
        </span>
      ))}
    </pre>
  );
}
