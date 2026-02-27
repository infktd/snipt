import { C } from "../styles/colors";

export const langColors: Record<string, string> = {
  go: C.sky,
  javascript: C.yellow,
  js: C.yellow,
  typescript: C.blue,
  ts: C.blue,
  python: C.green,
  py: C.green,
  bash: C.green,
  sh: C.green,
  sql: C.yellow,
  html: C.red,
  css: C.sky,
  json: C.peach,
  markdown: C.teal,
  md: C.teal,
  rust: C.peach,
  ruby: C.red,
  yaml: C.peach,
  toml: C.lavender,
  nix: C.mauve,
};

export function highlightTitle(title: string, indices: number[] | null | undefined) {
  if (!indices || indices.length === 0) return title;

  const indexSet = new Set(indices);
  const parts: JSX.Element[] = [];
  let current = "";
  let inMatch = false;

  for (let i = 0; i < title.length; i++) {
    const isMatch = indexSet.has(i);
    if (isMatch !== inMatch) {
      if (current) {
        parts.push(
          inMatch ? <mark key={i}>{current}</mark> : <span key={i}>{current}</span>
        );
      }
      current = "";
      inMatch = isMatch;
    }
    current += title[i];
  }
  if (current) {
    parts.push(
      inMatch ? <mark key="last">{current}</mark> : <span key="last">{current}</span>
    );
  }

  return <>{parts}</>;
}
