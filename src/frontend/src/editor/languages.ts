import type { LanguageSupport } from "@codemirror/language";

type LanguageLoader = () => Promise<LanguageSupport>;

const languageMap: Record<string, LanguageLoader> = {
  go: () => import("@codemirror/lang-go").then((m) => m.go()),
  javascript: () => import("@codemirror/lang-javascript").then((m) => m.javascript()),
  js: () => import("@codemirror/lang-javascript").then((m) => m.javascript()),
  typescript: () =>
    import("@codemirror/lang-javascript").then((m) => m.javascript({ typescript: true })),
  ts: () =>
    import("@codemirror/lang-javascript").then((m) => m.javascript({ typescript: true })),
  jsx: () =>
    import("@codemirror/lang-javascript").then((m) => m.javascript({ jsx: true })),
  tsx: () =>
    import("@codemirror/lang-javascript").then((m) =>
      m.javascript({ jsx: true, typescript: true })
    ),
  python: () => import("@codemirror/lang-python").then((m) => m.python()),
  py: () => import("@codemirror/lang-python").then((m) => m.python()),
  sql: () => import("@codemirror/lang-sql").then((m) => m.sql()),
  html: () => import("@codemirror/lang-html").then((m) => m.html()),
  css: () => import("@codemirror/lang-css").then((m) => m.css()),
  json: () => import("@codemirror/lang-json").then((m) => m.json()),
  markdown: () => import("@codemirror/lang-markdown").then((m) => m.markdown()),
  md: () => import("@codemirror/lang-markdown").then((m) => m.markdown()),
};

export async function loadLanguage(lang: string): Promise<LanguageSupport | null> {
  const loader = languageMap[lang.toLowerCase()];
  if (!loader) return null;
  return loader();
}

