import { createHighlighterCore, createCssVariablesTheme } from "shiki/core";
import { createJavaScriptRegexEngine } from "shiki/engine/javascript";
import python from "shiki/langs/python.mjs";
import go from "shiki/langs/go.mjs";
import sql from "shiki/langs/sql.mjs";
import bash from "shiki/langs/bash.mjs";
import yaml from "shiki/langs/yaml.mjs";
import json from "shiki/langs/json.mjs";
import dockerfile from "shiki/langs/dockerfile.mjs";
import typescript from "shiki/langs/typescript.mjs";
import javascript from "shiki/langs/javascript.mjs";
import toml from "shiki/langs/toml.mjs";
import ini from "shiki/langs/ini.mjs";
const theme = createCssVariablesTheme({
  name: "devdocs",
  variablePrefix: "--shiki-",
  fontStyle: true,
});
const hl = createHighlighterCore({
  themes: [theme],
  langs: [
    python,
    go,
    sql,
    bash,
    yaml,
    json,
    dockerfile,
    typescript,
    javascript,
    toml,
    ini,
  ],
  engine: createJavaScriptRegexEngine(),
});
export async function highlight(code: string, lang: string) {
  const h = await hl;
  const alias = lang === "shell" ? "bash" : lang;
  const l = h.getLoadedLanguages().includes(alias) ? alias : "text";
  return h.codeToTokens(code, { lang: l, theme: "devdocs" }).tokens;
}
