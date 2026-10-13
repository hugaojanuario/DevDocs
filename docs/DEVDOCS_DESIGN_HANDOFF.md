# DevDocs — Handoff de design para implementação

> Documento para o agente/dev que vai implementar o frontend em `web/`.
> Fonte da verdade visual: design system **DevDocs** (artifact de Design System). Fonte da verdade de dados: `api/openapi.yaml`.
> Este arquivo é autocontido: tokens, mapeamento Tailwind/shadcn, specs de componentes, estados, rotas e checklist de aceite.

---

## 0. Como usar este handoff

1. Copie a seção **2 (tokens.css)** literalmente para `web/src/styles/tokens.css`. Não reescreva valores à mão.
2. Configure Tailwind conforme a seção **3**, apontando tudo para as variáveis CSS (nunca hex no JSX).
3. Instale os primitivos shadcn/ui listados na seção **4** e aplique os overrides indicados.
4. Implemente os componentes na ordem da seção **6**, começando pela página de resultado (núcleo do produto).
5. Valide contra o checklist da seção **11** antes de considerar pronto.

**Regras que não podem ser quebradas**

- Resultado é **página de documentação**, nunca bolha de chat.
- Selo **Fonte oficial** só com `official === true` vindo da API. Nunca inferir pelo domínio.
- Seção sem dados **não renderiza**. Conteúdo parcial é declarado com `<Notice tone="partial">`.
- `mode: "ai"` sempre rotulado com cor `ai` + texto "Síntese por IA". Nunca usar `accent` para IA.
- **Zero persistência** de pesquisas: nada em localStorage/IndexedDB/sessionStorage, sem persistir cache do TanStack Query. Só o tema pode ir para localStorage.
- Nenhum hex, px mágico ou fonte fora dos tokens no código de componente.

---

## 1. Stack e dependências

```bash
npm create vite@latest web -- --template react-ts
cd web
npm i react-router-dom @tanstack/react-query lucide-react shiki clsx tailwind-merge class-variance-authority sonner
npm i -D tailwindcss@3.4.17 postcss autoprefixer @types/node
npx tailwindcss init -p
npx shadcn@latest init        # style: default, base color: neutral, CSS variables: yes
npx shadcn@latest add button dialog alert-dialog sheet tooltip input dropdown-menu sonner skeleton
```

Fontes (no `index.html`):

```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@400;500;600&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
```

> Alternativa offline: `@fontsource/ibm-plex-sans` e `@fontsource/jetbrains-mono` (mesmos pesos).

---

## 2. Tokens — `src/styles/tokens.css`

Gerado do `tokens.json` do design system. **Dark é o tema padrão.** O tema é aplicado via atributo `data-theme` no `<html>`.

```css
:root,
[data-theme="dark"] {
  color-scheme: dark;
  --surface: #0d0f12;
  --surface-raised: #14171b;
  --surface-sunken: #090a0c;
  --surface-hover: #1b1f24;
  --line: #252a30;
  --line-control: #5f6873;
  --ink: #e6e8eb;
  --ink-muted: #9ba3ad;
  --ink-subtle: #7f8893;
  --accent: #5ad6a6;
  --accent-soft: #0f2a21;
  --on-accent: #04140e;
  --ai: #f0b955;
  --ai-soft: #2b2210;
  --danger: #ff8a80;
  --danger-soft: #36151a;
  --on-danger: #1a0506;
  --link: var(--accent);
  --focus: var(--accent);
  --syn-keyword: #ff9e7a;
  --syn-string: #a6e3a1;
  --syn-function: #82c8ff;
  --syn-comment: #7f8893;
  --syn-number: #f5c46b;
  --syn-type: #d0a8ff;
  --syn-plain: #d4d8dd;
  --syn-punct: #aeb6bf;
  --shadow-popover: 0 8px 24px rgba(0,0,0,0.5);
  --shadow-dialog: 0 24px 64px rgba(0,0,0,0.6);
}

[data-theme="light"] {
  color-scheme: light;
  --surface: #fbfbfa;
  --surface-raised: #ffffff;
  --surface-sunken: #f3f4f2;
  --surface-hover: #eff0ee;
  --line: #e2e3df;
  --line-control: #878c93;
  --ink: #15171a;
  --ink-muted: #565c63;
  --ink-subtle: #646a71;
  --accent: #0b7a55;
  --accent-soft: #e2f4ec;
  --on-accent: #ffffff;
  --ai: #8a5800;
  --ai-soft: #fbf0d9;
  --danger: #c0322b;
  --danger-soft: #fcebe9;
  --on-danger: #ffffff;
  --link: var(--accent);
  --focus: var(--accent);
  --syn-keyword: #b8401c;
  --syn-string: #2f7a2a;
  --syn-function: #0b5fa8;
  --syn-comment: #646a71;
  --syn-number: #8a5800;
  --syn-type: #7a3fbf;
  --syn-plain: #24272b;
  --syn-punct: #4a5057;
  --shadow-popover: 0 8px 24px rgba(16,18,20,0.10);
  --shadow-dialog: 0 24px 64px rgba(16,18,20,0.16);
}

:root {
  --space-1: 4px;
  --space-2: 8px;
  --space-3: 12px;
  --space-4: 16px;
  --space-6: 24px;
  --space-8: 32px;
  --space-12: 48px;
  --space-16: 64px;
  --radius-sm: 4px;
  --radius-md: 6px;
  --radius-lg: 10px;
  --radius-full: 9999px;
  --content-width: 760px;
  --wide-width: 1040px;
  --sidebar-width: 240px;
  --header-height: 52px;
  --control-height: 36px;
  --font-sans: "IBM Plex Sans", ui-sans-serif, system-ui, -apple-system, "Segoe UI", sans-serif;
  --font-mono: "JetBrains Mono", ui-monospace, "SF Mono", Menlo, Consolas, monospace;
}
```

### Tema (dark / light / system)

```ts
// src/lib/theme.ts
export type ThemePref = 'dark' | 'light' | 'system';
const KEY = 'devdocs.theme'; // única chave permitida no localStorage

export function readThemePref(): ThemePref {
  try { return (localStorage.getItem(KEY) as ThemePref) ?? 'dark'; } catch { return 'dark'; }
}
export function applyTheme(pref: ThemePref) {
  const sys = window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark';
  document.documentElement.dataset.theme = pref === 'system' ? sys : pref;
  try { localStorage.setItem(KEY, pref); } catch {}
}
```

Aplique `readThemePref()` num script inline no `index.html` **antes** do bundle para evitar flash. Em `system`, escute `matchMedia(...).addEventListener('change', ...)`.

---

## 3. Tailwind — `tailwind.config.ts`

```ts
import type { Config } from 'tailwindcss';

const v = (name: string) => `var(--${name})`;

export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  darkMode: ['selector', '[data-theme="dark"]'],
  theme: {
    colors: {
      transparent: 'transparent',
      current: 'currentColor',
      surface: { DEFAULT: v('surface'), raised: v('surface-raised'), sunken: v('surface-sunken'), hover: v('surface-hover') },
      line: { DEFAULT: v('line'), control: v('line-control') },
      ink: { DEFAULT: v('ink'), muted: v('ink-muted'), subtle: v('ink-subtle') },
      accent: { DEFAULT: v('accent'), soft: v('accent-soft'), on: v('on-accent') },
      ai: { DEFAULT: v('ai'), soft: v('ai-soft') },
      danger: { DEFAULT: v('danger'), soft: v('danger-soft'), on: v('on-danger') },
      link: v('link'),
      focus: v('focus'),
      syn: {
        keyword: v('syn-keyword'), string: v('syn-string'), function: v('syn-function'), comment: v('syn-comment'),
        number: v('syn-number'), type: v('syn-type'), plain: v('syn-plain'), punct: v('syn-punct'),
      },
    },
    fontFamily: { sans: v('font-sans'), mono: v('font-mono') },
    fontSize: {
      display:      ['40px', { lineHeight: '44px', letterSpacing: '-0.02em',  fontWeight: '600' }],
      title:        ['28px', { lineHeight: '34px', letterSpacing: '-0.015em', fontWeight: '600' }],
      heading:      ['20px', { lineHeight: '28px', letterSpacing: '-0.01em',  fontWeight: '600' }],
      subheading:   ['15px', { lineHeight: '22px', fontWeight: '600' }],
      body:         ['15px', { lineHeight: '24px' }],
      'body-sm':    ['13px', { lineHeight: '20px' }],
      label:        ['12px', { lineHeight: '16px', letterSpacing: '0.01em', fontWeight: '500' }],
      code:         ['13px', { lineHeight: '21px' }],
      'code-sm':    ['12px', { lineHeight: '18px', fontWeight: '500' }],
    },
    spacing: {
      0: '0', px: '1px', 0.5: '2px',
      1: v('space-1'), 2: v('space-2'), 3: v('space-3'), 4: v('space-4'),
      6: v('space-6'), 8: v('space-8'), 12: v('space-12'), 16: v('space-16'),
      control: v('control-height'), header: v('header-height'), sidebar: v('sidebar-width'),
    },
    borderRadius: { none: '0', sm: v('radius-sm'), md: v('radius-md'), lg: v('radius-lg'), full: v('radius-full') },
    boxShadow: {
      none: 'none',
      popover: v('shadow-popover'),
      dialog: v('shadow-dialog'),
      focus: `0 0 0 2px ${v('surface')}, 0 0 0 4px ${v('focus')}`,
    },
    extend: {
      maxWidth: { content: v('content-width'), wide: v('wide-width') },
      width: { sidebar: v('sidebar-width') },
      transitionDuration: { DEFAULT: '120ms' },
    },
  },
  plugins: [],
} satisfies Config;
```

> A escala de espaçamento é **substituída** de propósito: só existem os passos do sistema (1, 2, 3, 4, 6, 8, 12, 16). Se precisar de outro valor, é sinal de que o layout está errado.

`src/styles/globals.css`:

```css
@import './tokens.css';
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  html { background: var(--surface); }
  body { @apply bg-surface text-ink font-sans text-body antialiased; }
  :focus-visible { outline: none; box-shadow: 0 0 0 2px var(--surface), 0 0 0 4px var(--focus); border-radius: var(--radius-sm); }
  ::selection { background: var(--accent-soft); color: var(--ink); }
  @media (prefers-reduced-motion: reduce) { *, *::before, *::after { animation: none !important; transition: none !important; } }
}
```

---

## 4. shadcn/ui — mapeamento e overrides

O shadcn espera variáveis próprias (`--background`, `--primary`…). **Não use a paleta dele.** Faça o alias para os tokens DevDocs no `globals.css` e remova os blocos `:root`/`.dark` que o `shadcn init` gerar:

```css
:root, [data-theme] {
  --background: var(--surface);
  --foreground: var(--ink);
  --card: var(--surface-raised);
  --card-foreground: var(--ink);
  --popover: var(--surface-raised);
  --popover-foreground: var(--ink);
  --primary: var(--accent);
  --primary-foreground: var(--on-accent);
  --secondary: var(--surface-raised);
  --secondary-foreground: var(--ink);
  --muted: var(--surface-hover);
  --muted-foreground: var(--ink-muted);
  --destructive: var(--danger);
  --destructive-foreground: var(--on-danger);
  --border: var(--line);
  --input: var(--line-control);
  --ring: var(--focus);
  --radius: var(--radius-md);
}
```

> **Conflito de nome:** o shadcn usa `--accent` como fundo de hover; no DevDocs `--accent` é o verde da marca. Não redefina `--accent`. Nos componentes gerados em `src/components/ui/`, troque `bg-accent text-accent-foreground` (hover/focus de itens de menu) por `bg-surface-hover text-ink`.
>
> Os tokens do DevDocs são hex, não triplas HSL. Ao gerar componentes shadcn, troque `hsl(var(--x))` por `var(--x)` nos arquivos em `src/components/ui/` (ou configure o `tailwind.config` como na seção 3 e use as classes DevDocs diretamente — recomendado).

| Primitivo shadcn | Uso no DevDocs | Override obrigatório |
|---|---|---|
| `Button` | base de `Button` | variants via `cva` da seção 6.1 |
| `AlertDialog` | `ConfirmDialog` | overlay `rgba(0,0,0,.55)`, `shadow-dialog`, foco inicial em **Cancelar** |
| `Sheet` (left) | nav mobile < 768px | largura `sidebar-width`, fundo `surface-raised` |
| `Sonner` | `Toast` | `position="bottom-right"`, 4s, erros `duration: Infinity` + botão fechar |
| `Tooltip` | itens da sidebar recolhida, botões icon-only | fundo `surface-raised`, borda `line`, `shadow-popover` |
| `Skeleton` | `Skeleton` | `bg-surface-hover`, pulso de opacidade 1.4s |
| `DropdownMenu` | ordenação em `/saved`, menu de sessão | mesmo estilo do Tooltip |

---

## 5. Paleta de referência (valores resolvidos)

| Token | Dark | Light | Uso |
|---|---|---|---|
| `surface` | `#0d0f12` | `#fbfbfa` | Fundo da página e da área de conteúdo. |
| `surface-raised` | `#14171b` | `#ffffff` | Sidebar, header, cards de consulta salva, diálogos e menus. |
| `surface-sunken` | `#090a0c` | `#f3f4f2` | Fundo de CodeBlock e de trechos de código inline. |
| `surface-hover` | `#1b1f24` | `#eff0ee` | Hover de linhas, itens de navegação e botões ghost; item ativo da sidebar. |
| `line` | `#252a30` | `#e2e3df` | Hairlines decorativas: divisórias de seção, bordas de card e de tabela. Não carrega significado sozinha. |
| `line-control` | `#5f6873` | `#878c93` | Borda de controles (input de busca, chips, botão secondary). ≥3:1 sobre surface e surface-raised nos dois temas. |
| `ink` | `#e6e8eb` | `#15171a` | Texto principal e títulos, sobre surface, surface-raised, surface-sunken e surface-hover. |
| `ink-muted` | `#9ba3ad` | `#565c63` | Texto secundário: descrições, pergunta original, metadados de fonte. Legível (≥4.5:1) em todas as superfícies. |
| `ink-subtle` | `#7f8893` | `#646a71` | Placeholder, datas, contadores, números de linha. ≥4.5:1 em surface, surface-raised, surface-sunken e surface-hover. |
| `accent` | `#5ad6a6` | `#0b7a55` | Ação primária, item de navegação ativo, links, anel de foco e selo de fonte oficial. Como texto, sobre qualquer superfície ou accent-soft. |
| `accent-soft` | `#0f2a21` | `#e2f4ec` | Fundo do selo Oficial, chip de tecnologia selecionado e Notice informativo; texto em accent. |
| `on-accent` | `#04140e` | `#ffffff` | Texto e ícones sobre fill accent (botão primary). Nunca branco literal no tema dark. |
| `ai` | `#f0b955` | `#8a5800` | Marca conteúdo sintetizado por IA (modo ai). Como texto, sobre as superfícies ou ai-soft. Sempre acompanhado da palavra “IA”. |
| `ai-soft` | `#2b2210` | `#fbf0d9` | Fundo do selo IA e do Notice de síntese por IA. |
| `danger` | `#ff8a80` | `#c0322b` | Exclusão permanente, erros de rede/autenticação. Como texto, sobre as superfícies ou danger-soft; sempre com ícone ou palavra. |
| `danger-soft` | `#36151a` | `#fcebe9` | Fundo de Notice de erro e hover do botão danger ghost. |
| `on-danger` | `#1a0506` | `#ffffff` | Texto sobre fill danger (botão Excluir permanentemente). |
| `link` | `{accent}` | `{accent}` | Links externos de fontes e links inline — alias de accent. |
| `focus` | `{accent}` | `{accent}` | Anel de foco: 2px de surface, depois 2px sólidos de focus. ≥3:1 sobre todas as superfícies. |
| `syn-keyword` | `#ff9e7a` | `#b8401c` | Shiki: palavras-chave (import, def, SELECT, func). Sobre surface-sunken. |
| `syn-string` | `#a6e3a1` | `#2f7a2a` | Shiki: strings e literais de caminho. Sobre surface-sunken. |
| `syn-function` | `#82c8ff` | `#0b5fa8` | Shiki: nomes de função e método. Sobre surface-sunken. |
| `syn-comment` | `#7f8893` | `#646a71` | Shiki: comentários. Sobre surface-sunken (≥4.5:1). |
| `syn-number` | `#f5c46b` | `#8a5800` | Shiki: números, booleanos e constantes. Sobre surface-sunken. |
| `syn-type` | `#d0a8ff` | `#7a3fbf` | Shiki: tipos, classes e módulos. Sobre surface-sunken. |
| `syn-plain` | `#d4d8dd` | `#24272b` | Shiki: texto padrão do código e identificadores. Sobre surface-sunken. |
| `syn-punct` | `#aeb6bf` | `#4a5057` | Shiki: pontuação e operadores. Sobre surface-sunken. |

**Contraste verificado (WCAG 2):** `ink`, `ink-muted`, `ink-subtle`, `accent`, `ai`, `danger` ≥ 4.5:1 sobre `surface`, `surface-raised`, `surface-sunken` e `surface-hover` nos dois temas. `line-control` e `focus` ≥ 3:1. Todos os `syn-*` ≥ 4.5:1 sobre `surface-sunken`. `line` é decorativa (< 3:1) — nunca use como única borda de controle.

### Tipografia

| Estilo | Família | Tamanho / linha | Peso | Tracking | Uso |
|---|---|---|---|---|---|
| `display` | sans | 40px / 44px | 600 | -0.02em | Headline editorial da Home. Uma por tela. |
| `title` | sans | 28px / 34px | 600 | -0.015em | Título do resultado e títulos de página (Consultas salvas, Tecnologias). |
| `heading` | sans | 20px / 28px | 600 | -0.01em | Seções do resultado: O que é?, Como funciona?, Entendendo o código, Observações, Fontes. |
| `subheading` | sans | 15px / 22px | 600 | 0 | Títulos de card, de fonte e de item salvo. |
| `body` | sans | 15px / 24px | 400 | 0 | Texto corrido da documentação. Medida máxima: content-width. |
| `body-sm` | sans | 13px / 20px | 400 | 0 | Metadados, explicações da tabela, descrições de tecnologia, itens da sidebar. |
| `label` | sans | 12px / 16px | 500 | 0.01em | Selos, rótulos de botão sm, cabeçalho de tabela. |
| `code` | mono | 13px / 21px | 400 | 0 | CodeBlock. Nunca quebrar linhas: scroll horizontal. |
| `code-sm` | mono | 12px / 18px | 500 | 0 | Linguagem do CodeBlock, trechos na tabela Entendendo o código, Kbd, código inline em body-sm. |

### Iconografia

`lucide-react`, `strokeWidth={1.75}`, tamanho 16 (14 em botões `sm`, 20 em empty states), cor herdada. Ícone decorativo: `aria-hidden`. Ícone sozinho: `aria-label` no botão.

| Ícone | Onde |
|---|---|
| `Search` | busca, nav Pesquisa, Nova pesquisa |
| `Bookmark` / `BookmarkCheck` | Salvar consulta / Salvo, nav Consultas salvas |
| `Copy` / `Check` | copiar código (feedback) |
| `Link` | Copiar link |
| `ExternalLink` | links de fontes |
| `Trash2` | excluir |
| `Pencil` | renomear |
| `RotateCw` | Pesquisar novamente, tentar novamente, spinner |
| `Layers2` | nav Tecnologias |
| `Settings` | nav Configurações |
| `ShieldCheck` | selo Fonte oficial |
| `Sparkles` | selo/aviso de IA |
| `CircleAlert` | conteúdo parcial, erros |
| `WifiOff` | sem conexão |
| `FileSearch` | sugestões da Home |
| `PanelLeftClose` | recolher sidebar |
| `Sun` / `Moon` / `Monitor` | tema light / dark / system |
| `X` | fechar toast |

Sem emoji em lugar nenhum.

### Marca

Não há logo desenhado. A marca é tipográfica: `devdocs` em `font-mono`, 600, 15px, minúsculo, `letter-spacing: -0.02em`, seguido de um bloco 8×16px `bg-accent` com `rounded-[1px]` e `margin-left: 2px`. Sem animação. Em texto corrido escreva "DevDocs".

---

## 6. Componentes

Estrutura de pastas e nome de cada arquivo. Todos tipados, sem `any`, props espelhando o `index.d.ts` do design system. Os estilos exatos estão no **Apêndice A** (CSS de referência) — traduza para classes Tailwind mantendo os valores.

```text
src/components/
├── ui/            # primitivos shadcn (gerados)
├── layout/        AppShell.tsx  Header.tsx  SidebarNav.tsx  MobileNav.tsx  Brand.tsx  SkipLink.tsx
├── search/        SearchField.tsx  TechChip.tsx  TechChipGroup.tsx  Suggestions.tsx
├── documentation/ ResultHeader.tsx  DocSection.tsx  SourceList.tsx  Notice.tsx  ResultView.tsx  ResultSkeleton.tsx
├── code/          CodeBlock.tsx  CodeExplanation.tsx  shiki.ts
├── saved/         SavedItem.tsx  SavedList.tsx  ConfirmDeleteDialog.tsx  RenameDialog.tsx
└── common/        Badge.tsx  Kbd.tsx  EmptyState.tsx  Icon.tsx
```

### 6.1 Button

| Variant | Fundo | Texto | Borda | Hover |
|---|---|---|---|---|
| `primary` | `accent` | `on-accent` | — | `brightness(1.08)` |
| `secondary` | `surface-raised` | `ink` | 1px `line-control` | `surface-hover` |
| `ghost` | transparente | `ink-muted` | — | `surface-hover` + `ink` |
| `danger` | `danger` | `on-danger` | — | — |
| `danger-ghost` | transparente | `danger` | — | `danger-soft` |

- Tamanhos: `md` altura 36px, padding-x 12px, 13px/500, gap 8px · `sm` 28px, padding-x 8px, 12px, gap 4px · `iconOnly` quadrado (36 ou 28).
- Raio `radius-md`. `disabled` → opacidade 55%. `loading` → troca ícone por `RotateCw` girando (0.8s linear), `aria-busy`, desabilitado.
- **Um primary por tela.** Rótulos: verbo + objeto, sentence case.

```ts
export const buttonVariants = cva(
  'inline-flex items-center justify-center gap-2 h-control px-3 rounded-md border border-transparent text-[13px] font-medium whitespace-nowrap transition-colors disabled:opacity-[.55] disabled:cursor-default focus-visible:shadow-focus',
  {
    variants: {
      variant: {
        primary: 'bg-accent text-accent-on hover:brightness-110',
        secondary: 'bg-surface-raised border-line-control text-ink hover:bg-surface-hover',
        ghost: 'bg-transparent text-ink-muted hover:bg-surface-hover hover:text-ink',
        danger: 'bg-danger text-danger-on',
        'danger-ghost': 'bg-transparent text-danger hover:bg-danger-soft',
      },
      size: { md: '', sm: 'h-[28px] px-2 text-[12px] gap-1' },
      iconOnly: { true: 'px-0 aspect-square', false: '' },
    },
    defaultVariants: { variant: 'secondary', size: 'md', iconOnly: false },
  },
);
```

### 6.2 Badge

Altura 22px, padding-x 8px, `radius-sm`, `label`, gap 4px, ícone 12px stroke 2.

| tone | Estilo | Ícone padrão | Quando |
|---|---|---|---|
| `tech` | contorno `line`, texto `ink`, mono 11px | — | tecnologia do resultado/item |
| `official` | `accent` sobre `accent-soft` | `ShieldCheck` | **só** com `official === true` |
| `deterministic` | contorno `line`, `ink-muted` | — | `mode: "deterministic"` |
| `ai` | `ai` sobre `ai-soft` | `Sparkles` | `mode: "ai"` → "Síntese por IA" |
| `neutral` | `ink-muted` sobre `surface-hover` | — | "Complementar", "Salvo em …" |
| `danger` | `danger` sobre `danger-soft` | `CircleAlert` | "Sem fontes" |

### 6.3 Kbd

20px de altura, min-width 20, padding-x 5px, mono 11px/500, `surface-raised`, borda `line` com borda inferior 2px, `radius-sm`, `ink-muted`. Mostre `⌘` no macOS e `Ctrl` nos demais (`navigator.platform`/`userAgentData`).

### 6.4 SearchField

- `<form role="search">` com input `type="search"`, `aria-label="Pesquisar na documentação"`, `autoComplete="off"`, `spellCheck={false}`.
- **hero**: altura 52px, `radius-lg`, borda `line-control`, fundo `surface-raised`, padding-left 16px, ícone `Search` 18px `ink-subtle`, input 16px/24px. Vazio: mostra `Kbd /` e `Kbd ⌘K`. À direita, Button `primary` "Buscar" (desabilitado se vazio).
- **compact** (header): altura 36px, `radius-md`, 14px, só `Kbd /` quando vazio, sem botão.
- Foco: borda `focus` + `box-shadow: 0 0 0 1px var(--focus)`.
- Placeholder fixo: **"O que você quer aprender hoje?"** (`ink-subtle`).
- Atalhos globais: `/` (fora de input/textarea/contenteditable) e `Cmd/Ctrl+K` focam o campo. Enter envia. Valor aparado; vazio não envia.
- Submit → `navigate('/search?q=' + encodeURIComponent(q) + (tech ? '&tech=' + tech : ''))`. **Nunca** gravar o termo em storage.

### 6.5 TechChip / TechChipGroup

Botão `aria-pressed`, altura 28px, padding-x 12px, `radius-md`, mono 12px/500, borda `line`, `ink-muted`. Hover: borda `line-control`, `ink`. Selecionado: `accent-soft` + borda `accent` + texto `accent` + ícone `Check` 12px. Grupo: `role="group" aria-label="Tecnologias"`, `flex-wrap gap-2`. Lista vinda de `GET /technologies` (seleção única; clicar de novo desmarca).

### 6.6 Suggestions (Home)

Lista de botões de linha inteira: ícone `FileSearch` 16px, 14px/22px, `ink-muted`, padding 8×12, `radius-md`, hover `surface-hover` + `ink`. Textos:

1. Como ler um CSV com Pandas?
2. Como escrever uma CTE no PostgreSQL?
3. Como criar uma DAG no Airflow?
4. Como carregar um .env no Python?
5. Como subir serviços com Docker Compose?

Clicar = submeter a busca com o texto.

### 6.7 CodeBlock

```text
┌ figure .border-line .rounded-lg .bg-surface-sunken .overflow-hidden ───────┐
│ figcaption h-40px px-4 border-b-line:  [python] [arquivo.py]     [⧉ Copiar] │
├─────────────────────────────────────────────────────────────────────────────┤
│ pre  py-4 overflow-x-auto  font-mono text-code  text-syn-plain  tab-size:4  │
│  1  import pandas as pd                                                     │
│  2                                                                          │
└─────────────────────────────────────────────────────────────────────────────┘
```

- Linguagem em `code-sm` `ink-subtle`; título opcional em `code-sm` `ink-muted`.
- `white-space: pre` **sempre**, inclusive no mobile. Nunca reformatar, trimar indentação interna ou converter tabs. Só remover `\n` final.
- `<pre tabIndex={0}>` para rolagem por teclado; foco = outline 2px `focus` inset.
- Números de linha opcionais (`showLineNumbers`): coluna de 2.5ch, `ink-subtle`, `user-select: none`, margem direita 16px.
- `highlightLines`: fundo `surface-hover` + `box-shadow: inset 2px 0 0 var(--accent)`.
- **Copiar**: Button `ghost sm`. `navigator.clipboard.writeText(code)` com o código **original**. Sucesso → ícone `Check`, texto "Copiado", cor `accent`, volta em 1.8s. Falha → "Falhou". Região `aria-live="polite"` anuncia "Código copiado". Não usar toast.
- Linguagens desconhecidas → `text` (sem highlight, sem erro).

**Shiki (`src/components/code/shiki.ts`)**

```ts
import { createHighlighter, createCssVariablesTheme, type Highlighter } from 'shiki';

const theme = createCssVariablesTheme({ name: 'devdocs', variablePrefix: '--shiki-', fontStyle: true });

let hl: Promise<Highlighter> | null = null;
const LANGS = ['python', 'go', 'sql', 'bash', 'shell', 'yaml', 'json', 'dockerfile', 'typescript', 'javascript', 'toml', 'ini'];

export function getHighlighter() {
  hl ??= createHighlighter({ themes: [theme], langs: LANGS });
  return hl;
}

export async function highlight(code: string, lang: string) {
  const h = await getHighlighter();
  const l = h.getLoadedLanguages().includes(lang) ? lang : 'text';
  return h.codeToHtml(code, { lang: l, theme: 'devdocs' });
}
```

Mapeamento das variáveis Shiki → tokens (em `globals.css`):

```css
:root, [data-theme] {
  --shiki-foreground: var(--syn-plain);
  --shiki-background: transparent;
  --shiki-token-keyword: var(--syn-keyword);
  --shiki-token-string: var(--syn-string);
  --shiki-token-string-expression: var(--syn-string);
  --shiki-token-function: var(--syn-function);
  --shiki-token-comment: var(--syn-comment);
  --shiki-token-constant: var(--syn-number);
  --shiki-token-parameter: var(--syn-plain);
  --shiki-token-punctuation: var(--syn-punct);
  --shiki-token-link: var(--link);
}
```

Carregue o Shiki **lazy** (`import()` dinâmico) para não pesar o bundle inicial; enquanto carrega, renderize o código sem cor (mesmo layout, sem layout shift).

### 6.8 CodeExplanation ("Entendendo o código")

- `<table>` com `<caption>` "Entendendo o código" (15px/600), colunas **Trecho** (42%) e **Explicação**.
- `th`: 12px/500 `ink-subtle`, padding 8×12, borda inferior `line`. `td`: 13px/20px `ink-muted`, padding 12, borda inferior `line`, `vertical-align: top`.
- Trecho em `<code>`: mono 12px/500 `ink`, `surface-sunken`, borda `line`, `radius-sm`, padding 1×6, `white-space: pre-wrap`.
- **Container query** `@container (max-width: 560px)`: esconde `thead`, cada `tr` vira bloco (trecho em cima, explicação embaixo, separados por 4px; linhas separadas por borda `line`).
- Renderiza logo após o `CodeBlock` correspondente, **somente** se `example.explanations.length > 0`.

### 6.9 ResultHeader

```text
[pandas] [🛡 Fonte oficial] [Determinístico]          ← selos, gap 8, mb 12
Como ler um arquivo CSV com Pandas                     ← title 28/34, text-wrap: balance
Pergunta: Como ler CSV usando Pandas?                  ← 14/22 ink-muted, mt 8
[Salvar consulta] [Copiar link] [Nova pesquisa]        ← mt 16, gap 8
───────────────────────────────────────────────────── ← pb 24, border-b line
```

- Ordem dos selos: tecnologia → Fonte oficial (condicional) → modo.
- **variant `result`**: `Salvar consulta` (primary, `Bookmark`) · `Copiar link` (ghost, `Link`) · `Nova pesquisa` (ghost, `Search`, foca o campo).
- `saveState`: `idle` → `saving` ("Salvando…", loading) → `saved` (secondary, `BookmarkCheck`, "Salvo", disabled) | `error` ("Tentar salvar de novo", `CircleAlert`) + toast de erro.
- **variant `saved`** (`/saved/:id`): selo neutro "Salvo em 12 out 2026" + `Pesquisar novamente` (secondary) · `Renomear` (ghost, `Pencil`) · `Excluir` (danger-ghost, `Trash2`).
- Primeira `DocSection` após o header tem `margin-top: 32px` (não 48).

### 6.10 DocSection

`<section aria-labelledby>` + `<h2 id>` (slug sem acentos). `h2`: `heading` (20/28/600), mb 12. Parágrafos `body`, mb 16. `code` inline: mono 13px/500, `surface-sunken`, borda `line`, `radius-sm`, padding 1×5. Entre seções: `margin-top: 48px`.

Ordem fixa e condicional:

| # | Título | Campo da API (confirmar no OpenAPI) | Renderiza se |
|---|---|---|---|
| 1 | O que é? | `summary` | não vazio |
| 2 | Como funciona? | `description` | não vazio |
| 3 | Exemplo / Exemplos | `examples[]` → CodeBlock + CodeExplanation | `examples.length > 0` |
| 4 | Observações | `notes[]` (lista) | `notes.length > 0` |
| 5 | Fontes e documentação | `references[]` | sempre que houver; se vazio → Notice `partial` "Nenhuma fonte foi retornada…" |

### 6.11 SourceList

- `<ul>` com borda `line`, `radius-lg`, `overflow-hidden`; itens separados por borda `line`.
- Cada item é um `<a target="_blank" rel="noopener noreferrer">`, padding 12×16, fundo `surface-raised`, hover `surface-hover`.
- Título 14px/22px/600 `ink` (hover: `link` + underline offset 3px). Meta 13px `ink-muted`: domínio (mono 12px) · Badge `official`/"Oficial" **ou** `neutral`/"Complementar" · `v{version}` se existir.
- Ícone `ExternalLink` 16px `ink-subtle` à direita + `<span class="sr-only">(abre em nova aba)</span>`.
- Ordem: oficiais primeiro, preservando a ordem da API dentro de cada grupo.

### 6.12 Notice

Bloco em linha (nunca modal), flex gap 12, padding 12×16, `radius-lg`, 14px/22px. Título 600 em bloco + corpo.

| tone | Fundo | Ícone/título | Corpo | role | Uso |
|---|---|---|---|---|---|
| `info` | `accent-soft` | `accent` `ShieldCheck` | `ink` | note | dicas neutras |
| `partial` | `surface-raised` + borda `line` | `ink-muted` `CircleAlert` | `ink-muted` | note | `status` parcial |
| `ai` | `ai-soft` | `ai` `Sparkles` | `ink` | note | logo após ResultHeader quando `mode: "ai"` |
| `error` | `danger-soft` | `danger` `CircleAlert` | `ink` | alert | erro de API, timeout, auth |
| `offline` | `danger-soft` | `danger` `WifiOff` | `ink` | alert | sem rede |

`action` opcional: um Button `sm` (Tentar novamente, Entrar).

### 6.13 SavedItem / SavedList

- Contêiner: `surface-raised`, borda `line`, `radius-lg`, `overflow-hidden`.
- Item: flex, gap 12, padding 12×16, borda inferior `line`, hover `surface-hover`.
  - Esquerda: `<a href="/saved/:id">` título 15/22/600 (hover `link`) + pergunta 13/20 `ink-muted` em uma linha com ellipsis (`title` com o texto completo).
  - Direita: Badge `tech` · `<time dateTime>` mono 12px `ink-subtle` ("12 out 2026") · Button `danger-ghost sm iconOnly` `Trash2` com `aria-label="Excluir {título}"`.
  - < 560px: meta quebra para linha inteira, `justify-between`.
- Toolbar de `/saved`: título `title` + total ("12 consultas") · busca local por texto (compact) · filtro de tecnologia (chips ou select) · ordenação (Mais recentes / Mais antigas / Título A–Z).

### 6.14 TechCard (`/technologies`)

Card `surface-raised`, borda `line`, `radius-lg`, padding 16, gap 8. Cabeçalho: nome 15/22/600 + Badge `neutral` com a categoria. Descrição 13/20 `ink-muted` (se houver). Rodapé (`flex-wrap`): link oficial mono 12px `link` + `ExternalLink` 12px · Button `ghost sm` `Search` "Pesquisar em {nome}" → `/?tech={slug}` com o chip pré-selecionado e foco no campo. Grid `repeat(auto-fill, minmax(260px, 1fr))`, gap 16, agrupado por categoria (h2 `heading`) quando a API enviar.

### 6.15 EmptyState

Coluna centralizada, padding 48×24, borda **tracejada** `line`, `radius-lg`. Ícone em quadrado 40px `surface-hover` `radius-md` (ícone 20px `ink-muted`), mb 16. Título 15/22/600. Descrição 13/20 `ink-muted`, max 360px. No máximo uma ação.

| Contexto | Título | Descrição | Ação |
|---|---|---|---|
| `/saved` vazio | Você ainda não salvou consultas | Ao encontrar uma resposta útil, use Salvar consulta para guardá-la aqui. O DevDocs não guarda seu histórico de pesquisas. | Fazer uma pesquisa (primary) |
| filtro sem resultado | Nenhuma consulta corresponde ao filtro | Ajuste o texto ou a tecnologia selecionada. | Limpar filtros (secondary) |
| busca sem resposta | Nenhuma documentação encontrada | Tente reformular a pergunta ou escolher outra tecnologia. | Nova pesquisa (secondary) |
| `/technologies` vazio | Nenhuma tecnologia cadastrada | O backend ainda não retornou tecnologias. | — |

### 6.16 ConfirmDeleteDialog

Baseado em `AlertDialog` (focus trap, `Esc`, `role="alertdialog"`).

- Painel: max 420px, padding 24, `surface-raised`, borda `line`, `radius-lg`, `shadow-dialog`. Overlay `rgba(0,0,0,.55)`.
- Título (17/24/600): **"Excluir consulta permanentemente?"**
- Descrição (14/22 `ink-muted`): "Esta ação não pode ser desfeita. A consulta e todo o seu conteúdo serão removidos do DevDocs."
- Ações à direita, gap 8: **Cancelar** (secondary, foco inicial) · **Excluir permanentemente** (danger, `Trash2`).
- Fluxo: confirmar → `loading` ("Excluindo…", ambos desabilitados, `Esc` bloqueado) → `DELETE /api/v1/saved/{id}`.
  - Sucesso: fecha, `invalidateQueries(['saved'])` e `removeQueries(['saved', id])`, toast "Consulta excluída", e se estava em `/saved/:id` navega para `/saved`.
  - Falha: diálogo **continua aberto**, mensagem de erro em `danger` com `CircleAlert` abaixo da descrição, item permanece na lista.
- Sem soft delete, sem lixeira, sem "desfazer".

### 6.17 RenameDialog (`/saved/:id`)

`Dialog` com input (compact, valor atual, `maxLength` conforme OpenAPI) e ações Cancelar / Salvar (primary). `PATCH /api/v1/saved/{id}` com `{ title }`. Sucesso → atualiza cache (`setQueryData`) + toast "Título atualizado". Enter salva, Esc cancela.

### 6.18 Toast

`sonner`, canto inferior direito. Altura mínima 44, padding 8×8×8×16, `surface-raised`, borda `line`, `radius-lg`, `shadow-popover`, 13/20/500. Ícone `Check` `accent` (sucesso) ou `CircleAlert` `danger` (erro). `role="status"`. Sucesso some em 4s; erro fica até fechar.

Disparar **somente** para: consulta salva · consulta excluída · título atualizado · link copiado · falha ao salvar/excluir/renomear.

### 6.19 Skeleton (carregando resultado)

Imita o resultado: linha de selos (3 × 60px × 22px), título (60% × 22px), depois 3 linhas de 12px (100%, 92%, 78%), gap 10, um bloco 160px simulando CodeBlock. `bg-surface-hover`, `radius-sm`, pulso de opacidade 1.4s. `aria-busy="true"` + `aria-label="Carregando documentação"`. Após **8s** sem resposta → mostrar Notice de timeout com "Tentar novamente" (manter requisição viva até o `AbortController` de 30s).

---

## 7. Layout e responsividade

```text
┌───────────────────────── header 52px · surface-raised · border-b line ─────────────────────────┐
│ devdocs▌        [🔍 Nova pesquisa            /]                          [☾] [sessão]          │
├──────────────┬──────────────────────────────────────────────────────────────────────────────────┤
│ sidebar 240  │   main · padding 32 24 64                                                        │
│ surface-     │        ┌──────── article max-w 760 (content-width), mx-auto ────────┐             │
│ raised       │        │ ResultHeader                                               │             │
│ border-r     │        │ DocSection …                                               │             │
│ line         │        └────────────────────────────────────────────────────────────┘             │
└──────────────┴──────────────────────────────────────────────────────────────────────────────────┘
```

| Breakpoint | Comportamento |
|---|---|
| ≥ 1280 | sidebar 240 expandida; conteúdo centrado em 760 (lista: 1040) |
| 1024–1279 | sidebar expandida, recolhível a 56px (estado só em memória ou localStorage como preferência de UI) |
| 768–1023 | sidebar recolhida (56px, só ícones + Tooltip) por padrão |
| < 768 | sem sidebar; botão menu no header abre `Sheet` com a mesma nav; gutters 16; busca compacta vira ícone que abre a busca em tela cheia |

- Header: `Brand` à esquerda, `SearchField compact` centralizado (max 420px) **exceto na Home** (a Home tem o hero), à direita botão de tema (menu Dark / Light / System com `Moon`/`Sun`/`Monitor`) e sessão.
- Sidebar: itens 34px, padding-x 12, gap 12, 13px/500 `ink-muted`; ativo = `aria-current="page"`, fundo `surface-hover`, `ink`, ícone `accent`, `box-shadow: inset 2px 0 0 var(--accent)`. Contador opcional (total de salvos) mono 11px `ink-subtle` à direita.
- `SkipLink` "Pular para o conteúdo" como primeiro elemento focável, visível só no foco.
- Rodapé: opcional, 12px `ink-subtle`, com versão da API e link para a documentação do projeto. Não criar footer de marketing.

---

## 8. Telas

### 8.1 `/` — Pesquisa

```text
                          (space-16 do topo)
            Encontre respostas na documentação            ← display 40/44, centralizado
   Respostas a partir de documentação oficial, com         ← body ink-muted, max 520
       código explicado e links para as fontes.
   [🔍 O que você quer aprender hoje?        / ⌘K  [Buscar]]  ← SearchField hero, max 680
   [python] [go] [postgresql] [pandas] [airflow] …           ← TechChipGroup, mt 16
                                                              
   Experimente                                                ← label ink-subtle, mt 48
   ⌕ Como ler um CSV com Pandas?                              ← Suggestions, max 680
   ⌕ …
   🛡 Priorizamos fontes oficiais. Conteúdo gerado por IA é sempre sinalizado.   ← body-sm ink-muted
```

- `autoFocus` no campo. Query param `?tech=` pré-seleciona o chip.

### 8.2 `/search?q=…&tech=…` — Resultado (prioridade 1)

Fluxo: lê `q` da URL → `POST /api/v1/search` (query key `['search', q, tech, aiEnabled]`, `gcTime` curto, **sem persistência**) → renderiza `ResultView`.

Ordem de renderização: `ResultHeader` → (`Notice ai` se `mode === 'ai'`) → (`Notice partial` se `status` parcial) → O que é? → Como funciona? → para cada exemplo: `CodeBlock` + `CodeExplanation` → Observações → Fontes e documentação.

Estados obrigatórios:

| Estado | UI |
|---|---|
| sem `q` | redireciona para `/` |
| loading | `ResultSkeleton` (header compacto já mostra o termo) |
| sucesso | `ResultView` |
| parcial | `ResultView` + `Notice partial` explicando o que faltou |
| vazio (`404`/sem conteúdo) | `EmptyState` "Nenhuma documentação encontrada" |
| erro de rede | `Notice offline` + Tentar novamente |
| timeout | `Notice error` "A busca demorou mais que o esperado." + Tentar novamente |
| fontes indisponíveis (`503`) | `Notice error` "As fontes de documentação estão indisponíveis no momento." |
| `401` | redireciona para `/login?next=…` |

**Copiar link** copia `location.href` e mostra toast "Link copiado". Texto auxiliar (tooltip): "O link contém só a pergunta, não a resposta salva."

**Salvar** → `POST /api/v1/saved` com o resultado **atual** (sem nova busca). Após sucesso, guardar o `id` retornado em estado de memória para manter "Salvo" e oferecer o link "Ver em Consultas salvas" no toast. Se a API retornar `409`, tratar como já salvo.

### 8.3 `/saved`

`GET /api/v1/saved` → toolbar (seção 6.13) + `SavedList`. Filtro/busca/ordenação são **client-side** sobre a lista carregada (ou query params da API se o OpenAPI suportar). Estados: loading (5 linhas skeleton), vazio, filtro vazio, erro.

### 8.4 `/saved/:id`

`GET /api/v1/saved/{id}` → mesma `ResultView` com `ResultHeader variant="saved"`. Conteúdo é o **original salvo**, nunca refeito. "Pesquisar novamente" navega para `/search?q={query}&tech={tech}`. `404` → EmptyState "Consulta não encontrada" + link para `/saved`.

### 8.5 `/technologies`

`GET /api/v1/technologies` (+ `GET /api/v1/sources` se trouxer os links oficiais). Agrupar por categoria na ordem: Linguagens, Bancos, Engenharia de Dados, Infraestrutura, Frameworks (categorias desconhecidas vão no fim, ordem alfabética). Só itens retornados pela API.

### 8.6 `/settings`

Seções em lista vertical com divisórias `line`, max 760:

| Seção | Controle | Persistência |
|---|---|---|
| Aparência | segmented Dark / Light / System | localStorage `devdocs.theme` |
| Idioma | pt-BR (único, desabilitado com nota "Mais idiomas em breve" — opcional) | — |
| Fontes | toggle "Priorizar fontes oficiais" | localStorage `devdocs.preferOfficial` (não sensível) |
| IA | toggle "Permitir síntese por IA" — **desabilitado** com `Notice info` "Disponível quando o backend habilitar síntese por IA" se a API não expuser a capacidade | localStorage só se habilitado |
| Sistema | status da API (`GET /health` + `/ready`): bolinha 8px `radius-full` `accent` "Conectado" / `danger` "Indisponível", versão da API, link para docs | — |

### 8.7 `/login`

Tela centrada, card 360px `surface-raised`: `Brand` + "Entrar no DevDocs" + campo(s) definidos no OpenAPI + Button primary "Entrar". Erro: `Notice error` inline. Sessão por **cookie HttpOnly**; enviar `credentials: 'include'` e o header CSRF conforme o OpenAPI. Sem cadastro, sem "esqueci a senha", sem multiusuário.

---

## 9. Camada de API

```text
src/services/
├── http.ts            # fetch wrapper: base '/api/v1', credentials include, CSRF, timeout, ApiError tipado
├── search.ts          # postSearch(body, signal)
├── saved.ts           # listSaved, getSaved, createSaved, renameSaved, deleteSaved
├── technologies.ts    # listTechnologies, listSources
├── system.ts          # health, ready
└── fixtures/          # SÓ dev/storybook, importadas via import.meta.env.VITE_USE_FIXTURES
src/types/api.ts       # gerado: npx openapi-typescript ../api/openapi.yaml -o src/types/api.ts
src/hooks/             # useSearch, useSavedList, useSaved, useCreateSaved, useRenameSaved, useDeleteSaved, useTechnologies, useHealth
```

- Tipos **gerados** do `api/openapi.yaml` (`openapi-typescript`). Não escrever interfaces de contrato à mão; se o OpenAPI não existir ainda, criar `src/types/api.draft.ts` marcado `// TEMPORÁRIO — substituir pelo gerado`.
- `ApiError { status, code, message }` → mapear para os estados da seção 8.2.
- Timeout: `AbortController` 30s; aviso visual em 8s.
- TanStack Query: `QueryClient` padrão, **sem** `persistQueryClient`. `search` com `staleTime: 0`, `gcTime: 5 * 60_000`, `retry: 1` (não repetir 4xx). `saved` invalida em create/rename/delete.
- Vite proxy de dev:

```ts
// vite.config.ts
server: { proxy: { '/api': { target: 'http://localhost:8080', changeOrigin: true } } }
```

Payload de busca (ilustrativo, confirmar no OpenAPI):

```json
{ "query": "Como ler CSV usando Pandas?", "language": "pt-BR", "preferred_sources": ["official"], "ai_enabled": false, "technology": "pandas" }
```

---

## 10. Acessibilidade e interação

- Foco visível em **todo** interativo: 2px `surface` + 2px `focus` (`shadow-focus`).
- Teclado: `/` e Cmd/Ctrl+K → busca · Enter → buscar · Esc → fecha diálogo/sheet · Tab segue header → sidebar → conteúdo.
- Landmarks: `header`, `nav[aria-label="Principal"]`, `main#conteudo`, `form[role=search]`.
- Um `h1` por página (título do resultado ou da tela). Seções em `h2`.
- `aria-live="polite"` para: copiar código, estados de salvar.
- Links externos: `rel="noopener noreferrer"` + texto oculto "(abre em nova aba)".
- Cor nunca é o único sinal: todo selo/aviso tem palavra e/ou ícone.
- Transições ≤ 120ms, só de cor. Sem animação de entrada. `prefers-reduced-motion` desliga spinner e pulso.
- Alvos de toque ≥ 28px (36px recomendado no mobile).

---

## 11. Checklist de aceite

**Visual**
- [ ] Nenhum hex/px/fonte fora de `tokens.css` em `src/components/**` (grep `#[0-9a-f]{3,6}` deve dar zero fora de `styles/`).
- [ ] Dark padrão, light e system funcionando sem flash no carregamento.
- [ ] CodeBlock mantém indentação e rola horizontalmente em 360px de largura.
- [ ] CodeExplanation empilha abaixo de 560px de container.
- [ ] Sidebar recolhe; mobile usa Sheet.

**Produto**
- [ ] Resultado nunca parece chat (sem bolhas, avatares, histórico).
- [ ] Selo Fonte oficial só com `official === true`.
- [ ] Seções vazias não aparecem; parcial é sinalizado.
- [ ] `mode: "ai"` com Badge + Notice âmbar.
- [ ] Salvar não refaz busca; estado "Salvo" impede duplicata.
- [ ] Excluir sempre passa pelo diálogo; falha mantém o item.
- [ ] `/saved/:id` exibe o conteúdo original; "Pesquisar novamente" não sobrescreve.
- [ ] `/technologies` só mostra o que a API retornou.

**Privacidade/segurança**
- [ ] DevTools → Application: nenhuma pesquisa/resultado em localStorage, sessionStorage, IndexedDB ou Cache Storage.
- [ ] Sem `persistQueryClient`.
- [ ] Nenhum segredo/token no bundle (`grep -ri "token\|secret\|password" dist/` só em strings de UI).
- [ ] Cookies de sessão HttpOnly; CSRF enviado conforme OpenAPI.

**Acessibilidade**
- [ ] Navegação completa só com teclado, foco sempre visível.
- [ ] axe/Lighthouse a11y ≥ 95 em `/`, `/search`, `/saved`.
- [ ] Contraste conferido nos dois temas (tokens já validados — não alterar valores).

**Entrega**
- [ ] `README.md` com `npm i`, `npm run dev`, `npm run build`, `npm run preview`, variáveis (`VITE_USE_FIXTURES`), integração com o OpenAPI e aviso de que fixtures são só para desenvolvimento.

---

## 12. Fora do escopo

Chatbot, histórico de conversas ou de buscas, playground de código, IDE web, dashboard de métricas, planos/pagamento, multiusuário, cadastro público, comentários sociais, painel admin, lixeira/soft delete.

---

## Apêndice A — CSS de referência dos componentes

CSS que roda os previews do design system (`components/bundle.css`). É a **especificação exata** de cada componente; traduza para classes Tailwind/cva mantendo valores e estados. Prefixo `dd-` = componente DevDocs.

```css
@import url("https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@400;500;600&family=JetBrains+Mono:wght@400;500&display=swap");

/* DevDocs components — every value from tokens.css */
body { margin: 0; font-family: var(--font-sans); font-size: 15px; line-height: 24px; background: var(--surface); color: var(--ink); -webkit-font-smoothing: antialiased; }
.dd-stack { display: flex; flex-direction: column; gap: var(--space-4); padding: var(--space-4); }
.dd-row { display: flex; flex-wrap: wrap; gap: var(--space-2); align-items: center; padding: var(--space-4); }
.dd-focus:focus-visible, .dd-btn:focus-visible, .dd-chip:focus-visible, .dd-nav a:focus-visible, .dd-src a:focus-visible, .dd-saved a:focus-visible, .dd-search input:focus-visible { outline: none; box-shadow: 0 0 0 2px var(--surface), 0 0 0 4px var(--focus); }
.dd-icon { flex: none; display: block; }

/* Button */
.dd-btn { display: inline-flex; align-items: center; justify-content: center; gap: var(--space-2); height: var(--control-height); padding: 0 var(--space-3); border-radius: var(--radius-md); border: 1px solid transparent; font: 500 13px/1 var(--font-sans); cursor: pointer; white-space: nowrap; transition: background-color .12s ease, border-color .12s ease, color .12s ease; }
.dd-btn:disabled { cursor: default; opacity: .55; }
.dd-btn-sm { height: 28px; padding: 0 var(--space-2); font-size: 12px; gap: var(--space-1); }
.dd-btn-icon { width: var(--control-height); padding: 0; }
.dd-btn-sm.dd-btn-icon { width: 28px; }
.dd-btn-primary { background: var(--accent); color: var(--on-accent); }
.dd-btn-primary:hover:not(:disabled) { filter: brightness(1.08); }
.dd-btn-secondary { background: var(--surface-raised); border-color: var(--line-control); color: var(--ink); }
.dd-btn-secondary:hover:not(:disabled) { background: var(--surface-hover); }
.dd-btn-ghost { background: transparent; color: var(--ink-muted); }
.dd-btn-ghost:hover:not(:disabled) { background: var(--surface-hover); color: var(--ink); }
.dd-btn-danger { background: var(--danger); color: var(--on-danger); }
.dd-btn-danger-ghost { background: transparent; color: var(--danger); }
.dd-btn-danger-ghost:hover:not(:disabled) { background: var(--danger-soft); }
.dd-btn[data-state="done"] { color: var(--accent); }
.dd-spin { animation: dd-spin .8s linear infinite; }
@keyframes dd-spin { to { transform: rotate(360deg); } }
@media (prefers-reduced-motion: reduce) { .dd-spin, .dd-skel { animation: none !important; } }

/* Kbd */
.dd-kbd { display: inline-flex; align-items: center; min-width: 20px; height: 20px; padding: 0 5px; box-sizing: border-box; justify-content: center; border: 1px solid var(--line); border-bottom-width: 2px; border-radius: var(--radius-sm); background: var(--surface-raised); color: var(--ink-muted); font: 500 11px/1 var(--font-mono); }

/* Badge */
.dd-badge { display: inline-flex; align-items: center; gap: var(--space-1); height: 22px; padding: 0 var(--space-2); border-radius: var(--radius-sm); font: 500 12px/16px var(--font-sans); letter-spacing: .01em; background: var(--surface-hover); color: var(--ink-muted); white-space: nowrap; }
.dd-badge-official { background: var(--accent-soft); color: var(--accent); }
.dd-badge-ai { background: var(--ai-soft); color: var(--ai); }
.dd-badge-deterministic { background: transparent; box-shadow: inset 0 0 0 1px var(--line); color: var(--ink-muted); }
.dd-badge-tech { background: transparent; box-shadow: inset 0 0 0 1px var(--line); color: var(--ink); font-family: var(--font-mono); font-size: 11px; }
.dd-badge-danger { background: var(--danger-soft); color: var(--danger); }

/* SearchField */
.dd-search { position: relative; display: flex; align-items: center; gap: var(--space-2); background: var(--surface-raised); border: 1px solid var(--line-control); border-radius: var(--radius-lg); padding: 0 var(--space-2) 0 var(--space-4); height: 52px; transition: border-color .12s ease; }
.dd-search:focus-within { border-color: var(--focus); box-shadow: 0 0 0 1px var(--focus); }
.dd-search input { flex: 1; min-width: 0; height: 100%; border: 0; background: transparent; color: var(--ink); font: 400 16px/24px var(--font-sans); outline: none; }
.dd-search input::placeholder { color: var(--ink-subtle); }
.dd-search .dd-search-ic { color: var(--ink-subtle); }
.dd-search-compact { height: var(--control-height); border-radius: var(--radius-md); padding-left: var(--space-3); }
.dd-search-compact input { font-size: 14px; }

/* TechChip */
.dd-chips { display: flex; flex-wrap: wrap; gap: var(--space-2); }
.dd-chip { height: 28px; padding: 0 var(--space-3); border-radius: var(--radius-md); border: 1px solid var(--line); background: transparent; color: var(--ink-muted); font: 500 12px/1 var(--font-mono); cursor: pointer; display: inline-flex; align-items: center; gap: var(--space-1); }
.dd-chip:hover { border-color: var(--line-control); color: var(--ink); }
.dd-chip[aria-pressed="true"] { background: var(--accent-soft); border-color: var(--accent); color: var(--accent); }
.dd-suggest { display: flex; flex-direction: column; }
.dd-suggest button { all: unset; box-sizing: border-box; display: flex; align-items: center; gap: var(--space-3); padding: var(--space-2) var(--space-3); border-radius: var(--radius-md); color: var(--ink-muted); font: 400 14px/22px var(--font-sans); cursor: pointer; }
.dd-suggest button:hover, .dd-suggest button:focus-visible { background: var(--surface-hover); color: var(--ink); }
.dd-suggest button:focus-visible { box-shadow: 0 0 0 2px var(--focus); }

/* CodeBlock */
.dd-code { border: 1px solid var(--line); border-radius: var(--radius-lg); background: var(--surface-sunken); overflow: hidden; }
.dd-code-head { display: flex; align-items: center; justify-content: space-between; gap: var(--space-2); height: 40px; padding: 0 var(--space-2) 0 var(--space-4); border-bottom: 1px solid var(--line); }
.dd-code-lang { font: 500 12px/18px var(--font-mono); color: var(--ink-subtle); }
.dd-code-title { font: 500 12px/18px var(--font-mono); color: var(--ink-muted); }
.dd-code pre { margin: 0; padding: var(--space-4) 0; overflow-x: auto; font: 400 13px/21px var(--font-mono); color: var(--syn-plain); tab-size: 4; white-space: pre; }
.dd-code pre:focus-visible { outline: 2px solid var(--focus); outline-offset: -2px; }
.dd-code code { display: block; min-width: max-content; }
.dd-ln { display: block; padding: 0 var(--space-4); }
.dd-code[data-numbers="true"] .dd-ln::before { content: attr(data-n); display: inline-block; width: 2.5ch; margin-right: var(--space-4); text-align: right; color: var(--ink-subtle); user-select: none; }
.dd-ln[data-hl="true"] { background: var(--surface-hover); box-shadow: inset 2px 0 0 var(--accent); }
.tk-k { color: var(--syn-keyword); } .tk-s { color: var(--syn-string); } .tk-f { color: var(--syn-function); } .tk-c { color: var(--syn-comment); font-style: italic; } .tk-n { color: var(--syn-number); } .tk-t { color: var(--syn-type); } .tk-p { color: var(--syn-punct); }
.dd-sr { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0); white-space: nowrap; }

/* CodeExplanation */
.dd-explain { container-type: inline-size; }
.dd-explain table { width: 100%; border-collapse: collapse; font: 400 13px/20px var(--font-sans); }
.dd-explain caption { text-align: left; font: 600 15px/22px var(--font-sans); padding-bottom: var(--space-2); color: var(--ink); }
.dd-explain th { text-align: left; font: 500 12px/16px var(--font-sans); color: var(--ink-subtle); padding: var(--space-2) var(--space-3); border-bottom: 1px solid var(--line); }
.dd-explain td { vertical-align: top; padding: var(--space-3); border-bottom: 1px solid var(--line); color: var(--ink-muted); }
.dd-explain td:first-child { width: 42%; }
.dd-explain td code { font: 500 12px/18px var(--font-mono); color: var(--ink); background: var(--surface-sunken); border: 1px solid var(--line); border-radius: var(--radius-sm); padding: 1px 6px; white-space: pre-wrap; word-break: break-word; }
@container (max-width: 560px) {
  .dd-explain thead { display: none; }
  .dd-explain tr, .dd-explain td { display: block; width: auto !important; }
  .dd-explain tr { padding: var(--space-3) 0; border-bottom: 1px solid var(--line); }
  .dd-explain td { border: 0; padding: var(--space-1) 0; }
}

/* Doc layout */
.dd-doc { max-width: var(--content-width); }
.dd-sec { margin-top: var(--space-12); }
.dd-sec:first-child { margin-top: 0; }
.dd-rh + .dd-sec { margin-top: var(--space-8); }
.dd-sec h2 { margin: 0 0 var(--space-3); font: 600 20px/28px var(--font-sans); letter-spacing: -.01em; color: var(--ink); }
.dd-sec p { margin: 0 0 var(--space-4); color: var(--ink); }
.dd-sec p code, .dd-inline { font: 500 13px/20px var(--font-mono); background: var(--surface-sunken); border: 1px solid var(--line); border-radius: var(--radius-sm); padding: 1px 5px; }

/* ResultHeader */
.dd-rh { padding-bottom: var(--space-6); border-bottom: 1px solid var(--line); }
.dd-rh-meta { display: flex; flex-wrap: wrap; gap: var(--space-2); align-items: center; margin-bottom: var(--space-3); }
.dd-rh h1 { margin: 0; font: 600 28px/34px var(--font-sans); letter-spacing: -.015em; color: var(--ink); text-wrap: balance; }
.dd-rh-q { margin: var(--space-2) 0 0; color: var(--ink-muted); font-size: 14px; line-height: 22px; }
.dd-rh-actions { display: flex; flex-wrap: wrap; gap: var(--space-2); margin-top: var(--space-4); }

/* SourceList */
.dd-src { list-style: none; margin: 0; padding: 0; border: 1px solid var(--line); border-radius: var(--radius-lg); overflow: hidden; }
.dd-src li + li { border-top: 1px solid var(--line); }
.dd-src a { display: flex; align-items: flex-start; gap: var(--space-3); padding: var(--space-3) var(--space-4); text-decoration: none; color: inherit; background: var(--surface-raised); }
.dd-src a:hover { background: var(--surface-hover); }
.dd-src-body { flex: 1; min-width: 0; }
.dd-src-title { display: block; font: 600 14px/22px var(--font-sans); color: var(--ink); }
.dd-src a:hover .dd-src-title { color: var(--link); text-decoration: underline; text-underline-offset: 3px; }
.dd-src-meta { display: flex; flex-wrap: wrap; gap: var(--space-2); align-items: center; margin-top: 2px; font: 400 13px/20px var(--font-sans); color: var(--ink-muted); }
.dd-src-domain { font-family: var(--font-mono); font-size: 12px; }
.dd-src .dd-icon { color: var(--ink-subtle); margin-top: 3px; }

/* Notice */
.dd-notice { display: flex; gap: var(--space-3); padding: var(--space-3) var(--space-4); border-radius: var(--radius-lg); border: 1px solid var(--line); background: var(--surface-raised); color: var(--ink); font: 400 14px/22px var(--font-sans); }
.dd-notice .dd-icon { margin-top: 3px; color: var(--ink-muted); }
.dd-notice-title { font-weight: 600; display: block; }
.dd-notice-body { color: var(--ink-muted); }
.dd-notice-action { margin-top: var(--space-2); }
.dd-notice-info { background: var(--accent-soft); border-color: transparent; }
.dd-notice-info .dd-icon { color: var(--accent); }
.dd-notice-ai { background: var(--ai-soft); border-color: transparent; }
.dd-notice-ai .dd-icon, .dd-notice-ai .dd-notice-title { color: var(--ai); }
.dd-notice-error { background: var(--danger-soft); border-color: transparent; }
.dd-notice-error .dd-icon, .dd-notice-error .dd-notice-title { color: var(--danger); }
.dd-notice-ai .dd-notice-body, .dd-notice-error .dd-notice-body, .dd-notice-info .dd-notice-body { color: var(--ink); }

/* SavedItem */
.dd-saved { display: flex; align-items: center; gap: var(--space-3); padding: var(--space-3) var(--space-4); border-bottom: 1px solid var(--line); }
.dd-saved:hover { background: var(--surface-hover); }
.dd-saved-main { flex: 1; min-width: 0; }
.dd-saved a { display: block; text-decoration: none; color: var(--ink); font: 600 15px/22px var(--font-sans); border-radius: var(--radius-sm); }
.dd-saved a:hover { color: var(--link); }
.dd-saved-q { margin: 2px 0 0; color: var(--ink-muted); font: 400 13px/20px var(--font-sans); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.dd-saved-meta { display: flex; align-items: center; gap: var(--space-3); flex: none; color: var(--ink-subtle); font: 400 12px/16px var(--font-mono); }
@media (max-width: 560px) { .dd-saved { flex-wrap: wrap; } .dd-saved-meta { width: 100%; justify-content: space-between; } }

/* EmptyState */
.dd-empty { display: flex; flex-direction: column; align-items: center; text-align: center; padding: var(--space-12) var(--space-6); border: 1px dashed var(--line); border-radius: var(--radius-lg); }
.dd-empty-ic { display: grid; place-items: center; width: 40px; height: 40px; border-radius: var(--radius-md); background: var(--surface-hover); color: var(--ink-muted); margin-bottom: var(--space-4); }
.dd-empty h3 { margin: 0; font: 600 15px/22px var(--font-sans); }
.dd-empty p { margin: var(--space-1) 0 var(--space-4); max-width: 360px; color: var(--ink-muted); font: 400 13px/20px var(--font-sans); }

/* ConfirmDialog */
.dd-overlay { position: fixed; inset: 0; display: grid; place-items: center; padding: var(--space-4); background: rgba(0,0,0,.55); }
.dd-overlay[data-inline="true"] { position: static; background: transparent; padding: var(--space-4); }
.dd-dialog { width: 100%; max-width: 420px; box-sizing: border-box; padding: var(--space-6); border: 1px solid var(--line); border-radius: var(--radius-lg); background: var(--surface-raised); box-shadow: var(--shadow-dialog); }
.dd-dialog h2 { margin: 0; font: 600 17px/24px var(--font-sans); color: var(--ink); }
.dd-dialog p { margin: var(--space-2) 0 0; color: var(--ink-muted); font: 400 14px/22px var(--font-sans); }
.dd-dialog-err { margin-top: var(--space-3) !important; color: var(--danger) !important; display: flex; gap: var(--space-2); align-items: center; }
.dd-dialog-actions { display: flex; justify-content: flex-end; gap: var(--space-2); margin-top: var(--space-6); }

/* SidebarNav */
.dd-nav { width: var(--sidebar-width); box-sizing: border-box; padding: var(--space-3) var(--space-2); background: var(--surface-raised); border-right: 1px solid var(--line); display: flex; flex-direction: column; gap: 2px; }
.dd-nav[data-collapsed="true"] { width: 56px; }
.dd-nav a { display: flex; align-items: center; gap: var(--space-3); height: 34px; padding: 0 var(--space-3); border-radius: var(--radius-md); color: var(--ink-muted); text-decoration: none; font: 500 13px/1 var(--font-sans); }
.dd-nav a:hover { background: var(--surface-hover); color: var(--ink); }
.dd-nav a[aria-current="page"] { background: var(--surface-hover); color: var(--ink); box-shadow: inset 2px 0 0 var(--accent); }
.dd-nav a[aria-current="page"] .dd-icon { color: var(--accent); }
.dd-nav-count { margin-left: auto; font: 500 11px/1 var(--font-mono); color: var(--ink-subtle); }
.dd-nav[data-collapsed="true"] .dd-nav-label, .dd-nav[data-collapsed="true"] .dd-nav-count { display: none; }
.dd-brand { display: flex; align-items: center; gap: 2px; font: 600 15px/1 var(--font-mono); color: var(--ink); letter-spacing: -.02em; }
.dd-brand-caret { display: inline-block; width: 8px; height: 16px; background: var(--accent); margin-left: 2px; border-radius: 1px; }

/* Skeleton */
.dd-skel { display: block; height: 12px; border-radius: var(--radius-sm); background: var(--surface-hover); animation: dd-pulse 1.4s ease-in-out infinite; }
@keyframes dd-pulse { 50% { opacity: .5; } }

/* Toast */
.dd-toast { display: inline-flex; align-items: center; gap: var(--space-3); padding: var(--space-2) var(--space-2) var(--space-2) var(--space-4); min-height: 44px; box-sizing: border-box; border: 1px solid var(--line); border-radius: var(--radius-lg); background: var(--surface-raised); box-shadow: var(--shadow-popover); font: 500 13px/20px var(--font-sans); color: var(--ink); }
.dd-toast-success .dd-icon { color: var(--accent); }
.dd-toast-error .dd-icon { color: var(--danger); }

/* TechCard */
.dd-tech { display: flex; flex-direction: column; gap: var(--space-2); padding: var(--space-4); border: 1px solid var(--line); border-radius: var(--radius-lg); background: var(--surface-raised); }
.dd-tech-head { display: flex; align-items: center; justify-content: space-between; gap: var(--space-2); }
.dd-tech h3 { margin: 0; font: 600 15px/22px var(--font-sans); }
.dd-tech p { margin: 0; color: var(--ink-muted); font: 400 13px/20px var(--font-sans); }
.dd-tech-foot { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: var(--space-2); margin-top: var(--space-2); }
.dd-tech-link { display: inline-flex; align-items: center; gap: var(--space-1); color: var(--link); font: 400 12px/16px var(--font-mono); text-decoration: none; border-radius: var(--radius-sm); }
.dd-tech-link:hover { text-decoration: underline; text-underline-offset: 3px; }
.dd-tech-link:focus-visible { outline: none; box-shadow: 0 0 0 2px var(--focus); }
```
