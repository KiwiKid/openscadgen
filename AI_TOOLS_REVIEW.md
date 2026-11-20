# Review: what `openscadgen -ai` actually advertises

> **Status: superseded by the MCP refactor.** The chat UI and OpenAI tool loop
> described below have been removed. `-ai` now publishes a Model Context Protocol
> server (HTTP and stdio) plus a read-only settings page. Route gating (H1), the
> approval/safety model (H2), symlink containment (M1), bounded reads (M2), git
> staging scope (M3), credential-prompt hangs (M4), false-success payloads (L1),
> and documentation (section 6) were all addressed. See `README.md#ai--mcp` for
> current behaviour. This document is kept as the pre-change audit.

Reviewed: working tree at `71c964b` (uncommitted changes present), 2026-11-21.
Scope: the `-ai` flag, the chat surface it reveals, and every action the model can take.

## 1. Bottom line

There is **no MCP server in this codebase**, and `-ai` has never enabled one.
`-ai` turns on an **in-process OpenAI function-calling agent**: a `/chat` page that
sends prompts to OpenAI's Responses API and executes 9 tools against the local
filesystem and git repository. It is a tool-calling loop, not Model Context Protocol
(no JSON-RPC, no stdio, no SSE, no MCP handshake, no `tools/list`).

The MCP wording is probably remembered from a prototype that was deleted a year
before `-ai` existed:

- `f36007a` (2025-08-18) *"remove prototype mcp"* deleted `server/mcp_server.go`.
  It exposed `/v1/metadata`, `/v1/resources`, `/v1/resource/`, `/v1/tools`,
  `/v1/tool/<id>/invoke` with tools `list_instances`, `process_instance`, `process_all`.
  It was not real MCP either — it was ad-hoc JSON REST, hardcoded to
  `bols2otropolis/cube/config.toml`, **registered unconditionally in server mode**,
  and unauthenticated.
- `-ai` (`ShowAI`) first appears in `c5852c8`, 2026-08-30 — **after** the removal.

So any documentation, note, or memory saying "`-ai` publishes MCP actions" is wrong.
What it publishes is the 9 tools in section 4.

## 2. What the flag actually says and does

| Surface | Text / behaviour | Reference |
| --- | --- | --- |
| `-h` help | "Show AI tools in server mode (the chat reads DEEPSEEK_API_TOKEN or OPENAI_API_KEY from the environment)" | `cmd/flags.go:117` |
| Env equivalent | `OPENSCADGEN_AI_ON=true` also sets `ShowAI` | `main.go:25-29` |
| Startup log | "AI tools enabled: <TOKEN> found in the server environment" or a warning that no token is set | `server/chat.go:381-392`, `server/http_server.go:153-160` |
| UI effect | Shows the "AI Chat" / "AI tools" nav links | `pkg/templates/config_files.templ:479-481`, `pkg/templates/report.templ:3210-3212` |

Functionally, the flag is **cosmetic**. The handlers are registered for every server
run regardless of the flag (`server/http_server.go:214-217`), and nothing in
`handleChatRequest`, `handleChatRun`, or the tool loop consults `aiToolsEnabled`.
`-ai` only controls links and the startup notice. Navigating to
`/chat?config=...` on a server started without `-ai` gives the full chat with the
full toolset.

## 3. Provider and runtime model

- OpenAI paths use `POST /v1/responses` with `tools` + `tool_choice: "auto"` and
  `parallel_tool_calls: false` (`server/chat.go:904-913`).
- Tool loop: up to 12 rounds, 90 s per HTTP call, each round appends model output
  items plus tool outputs (`server/chat.go:900-971`).
- Execution is **async** through `/api/chat/run` + `/api/chat/status`, with job
  events kept 10 minutes and running jobs never garbage-collected
  (`server/chat_async.go:16-18`, `311-320`).
- **Tools exist only for the `openai` provider.** `chatProjectToolsEnabled` returns
  false for DeepSeek (`server/chat.go:993-995`), so a DeepSeek chat is text-only.
  The flag help ambiguously lists both tokens as if equivalent.
- Tokens are read from the server process env and never sent to the browser. Good.
- Limits: 24 history messages, 16 KB auto-attached context, 20 KB per read,
  24 KB tool output, 24 KB diff, 3 vision images / 4 MB
  (`server/chat.go:33-39`, `server/chat_process_tool.go:21-24`).

## 4. The advertised actions

Exactly 9 `function` tools are advertised to the model whenever a project root is
attached (`server/chat.go:1007-1136`):

| # | Tool | What it does | Blast radius | Notes |
| --- | --- | --- | --- | --- |
| 1 | `list_project_files` | Walks the project tree, depth ≤ 4, 40 entries | Read | Multi-project scope lists every selected root |
| 2 | `read_project_file` | Reads UTF-8 text, offset/max_bytes | Read | Whole file loaded then sliced |
| 3 | `edit_project_file` | Exact `old_text` → `new_text`, optional replace-all | **Write** | No backup, no approval |
| 4 | `write_project_file` | Create or fully overwrite any file under root | **Write** | Unbounded content; creates dirs |
| 5 | `process_config_and_review` | Loads a `config.toml`, runs OpenSCAD, optionally asks OpenAI vision about up to 3 PNGs | **Executes renderer + egress of images** | Same process as a normal run |
| 6 | `git_status` | `git status --short --branch -- .` | Read | Scoped to project root |
| 7 | `git_diff` | `git diff [--cached] -- <path>` | Read | Truncated to 24 KB |
| 8 | `git_commit` | `git add -A -- .` then `git commit -m` | **Write to history** | Stages everything under root |
| 9 | `git_push` | `git push [-u] [remote] [branch]` | **Remote egress** | No credential-prompt suppression |

Notable: **every tool call executes immediately.** There is no confirmation step, no
dry-run, no diff preview, and no user-visible "apply" gate anywhere in
`executeChatToolCalls` (`server/chat.go:1138-1172`). The only guardrails are the
system prompt ("Never commit or push unless the user explicitly asks",
`server/chat.go:604`) and the lexical path checks.

## 5. Findings

### High

**H1. The flag is not a security boundary.** `-ai` hides links; it does not gate
`/chat`, `/api/chat/run`, or any tool. A user who believes AI tools are "off"
still has a reachable page that can overwrite files and push commits if
`OPENAI_API_KEY` happens to be in the server's environment.
`server/http_server.go:214-217`, `30`, `153`.

**H2. No approval gate on destructive actions.** `edit`, `write`, `git_commit`, and
`git_push` are all reachable in a single model round-trip with no human check.
The async design makes this worse: the browser polls for progress but has no chance
to veto anything (`server/chat_async.go:70-104`).

**H3. Prompt injection has a direct write/push path.** `config.toml` content and a
project listing are auto-attached to every request (`server/chat.go:589-631`), and
the model can read any project file. A malicious `.scad`, README, or config comment
can instruct the model to edit files or push. The tool descriptions repeat "only
when the user explicitly asks", but that is prompt text, not enforcement.

**H4. LAN exposure is documented as safe but is not.** `-server-address` help says
"non-loopback exposure requires authentication and a reverse proxy"
(`cmd/flags.go:110`), yet the binary contains **no authentication or CSRF
protection at all** (grep for BasicAuth/Authorization/cookie finds only
`sessionStorage` UI preferences). Anyone who can reach the port can drive the agent
and spend the API budget. The help text describes a control that the code does not
have.

### Medium

**M1. Symlink/junction escape.** Containment is purely lexical
(`filepath.Rel` in `resolveProjectPath`, `server/chat.go:1706-1738`;
`pathWithinRoot`, `server/chat_scope.go:251-259`). `os.ReadFile`/`os.WriteFile`
follow links, so a symlink inside the project root (`link -> ~/.ssh`) lets
`read_project_file` / `write_project_file` reach outside it. No `EvalSymlinks`
call exists in the chat path (only in `server/http_server.go:1821`, unrelated).

**M2. `read_project_file` loads the entire file into memory.** `os.ReadFile`
happens before the `max_bytes` slice (`server/chat.go:1465-1480`). Pointing it at a
multi-GB STL or `models.test` (3.6 MB here, far larger in the wild) is an
unbounded allocation in a long-lived server.

**M3. `git_commit` commits more than the agent changed.** `git add -A -- .`
(`server/chat.go:1628-1641`) stages every modification, deletion, and untracked file
under the project root, then commits. Any WIP the user had is swept into an
agent-authored commit with a model-written message. There is no diff confirmation.
`ensureNoStagedChangesOutsideProject` (`1769-1799`) only checks that the *staged
set* stays inside the root — it does not stop the sweeping.

**M4. `git_push` can hang on credentials.** `runProjectCommand` inherits the
environment but never sets `GIT_TERMINAL_PROMPT=0` / `GIT_ASKPASS`
(`server/chat.go:1740-1760`). Without a usable credential helper the push blocks
until the 2-minute timeout, holding a chat job open.

### Low

- **L1.** `process_config_and_review` returns `"ok": true` when the config fails to
  load (`server/chat_process_tool.go:136-143`) and when processing fails
  (`151-160`). Only `config_valid` / `process_succeeded` are false, so the model
  can easily report success after a failed run.
- **L2.** `write_project_file` has no size cap; `content` is bounded only by what
  the model emits.
- **L3.** Model name is taken verbatim from the form (`readChatModel`,
  `server/chat.go:770-776`); it is not validated against the provider. Defaults are
  hardcoded (`gpt-4.1`, `deepseek-v4-flash`, `server/chat.go:29-30`) — verify the
  DeepSeek default is a model your account/plan actually serves.
- **L4.** Only `.toml` writes are journaled (`shouldJournalProjectWrite`,
  `server/chat.go:1573-1576`); `.scad` edits leave no audit trail, and job events
  expire after 10 minutes. After the fact, there is no record of what the agent did.
- **L5.** Exceeding 12 tool rounds throws away the whole response
  (`server/chat.go:970`), so a long but useful run ends in an error with no partial
  answer.
- **L6.** A config edited by the agent can set `output_path` outside the project
  root (`pkg/models/types.go:45`, absolute paths are supported), redirecting
  generated artifacts anywhere the process can write. Pre-existing capability, but
  now reachable by model output.
- **L7.** Dead code: `chatToolCall.ServerLabel` (`server/chat.go:186`) is an unused
  MCP-style vestige; `tools.templ:131` filters a `"AI Chat"` tool entry that no
  longer exists in `pkg/openscad_tools.go`.

### What is actually good

- The multi-project scope model is well built: validated against known project
  folders, prompt text explains it, and both single- and multi-root path resolution
  are covered by tests (`server/chat_scope.go`, `server/chat_scope_test.go:273-364`).
- `pkg.Process` rejects `custom_openscad_command` / `custom_openscad_args` without
  `--dangerously-skip-permissions` (`pkg/processor.go:2260-2278`), so a
  model-authored config cannot redirect the renderer to an arbitrary binary or add
  shell-ish arguments through the chat path.
- Git and renderer invocations use `exec.Command` with an argv slice, not a shell —
  no command injection.
- API tokens stay server-side; the browser only ever sees "configured / not".
- `git status`/`diff` are path-scoped to the project root even when the repo is
  larger, and `ensureNoStagedChangesOutsideProject` genuinely blocks committing
  other projects' staged work.
- Chat tests pass (`go test ./server -run 'Chat|ProjectTool|OpenAI'`) and the build
  is clean.

## 6. Documentation gap

`-ai`, the chat, the 9 tools, and both token env vars are **completely undocumented**:
zero matches for chat/OpenAI/DeepSeek in `README.md`, `INSTALL.md`, and
`RELEASE_NOTES.md`. The only user-facing description of the action set is two lines
of in-page help (`pkg/templates/chat.templ:91`):

> "OpenAI can list, read, edit, commit, and push within this project root when you
> explicitly ask for those git actions."

That under-advertises the two most consequential tools: full-file overwrite
(`write_project_file`) and **running the OpenSCAD renderer plus uploading renders to
OpenAI** (`process_config_and_review`). The flag's own help mentions neither.

## 7. Recommendations, in priority order

1. **Gate the routes, not just the links.** Register `/chat`, `/api/chat/run`, and
   `/api/chat/status` only when `ShowAI`, or return 404 from the handlers when
   `aiToolsEnabled` is false. This makes `-ai` mean what its help says.
2. **Add an approval gate for writes.** Some token the model cannot supply (a
   one-shot per-request confirm value, or a "review changes" queue) before
   `write_project_file`, `git_commit`, and `git_push`. At minimum, require a
   diff-confirm step before commit/push.
3. **Resolve symlinks** before containment checks (`EvalSymlinks` on the root and
   the candidate, then compare), and skip reparse points while walking.
4. **Cap reads before allocation** (`io.LimitReader` / `Stat` + reject oversized
   files), and cap `content` in `write_project_file`.
5. **Fix the `-server-address` help text** or implement the authentication it
   claims. Right now the flag is a footgun with misleading documentation.
6. **Make `git_commit` narrow:** stage only paths the agent touched in this run,
   or require explicit `paths` and default to `git add -- <paths>` instead of
   `add -A -- .`.
7. **Set `GIT_TERMINAL_PROMPT=0`** for agent git commands so pushes fail fast
   instead of hanging.
8. **Fix `ok: true` on failure** in `process_config_and_review`, and surface
   `process_error` prominently so the model does not claim success.
9. **Document the feature.** A README section covering the flag, env vars, the tool
   list with write/push/render called out, the OpenAI-only limitation, the fact
   that file contents and rendered PNGs leave the machine, and how to turn it off.
10. **Add tests** for: route gating when `-ai` is absent, provider gating
    (DeepSeek has no tools), symlink escape, oversized reads, `git_commit` staging
    scope, and credential-prompt timeout.

## 8. If you want the short version

`-ai` is an undocumented, always-reachable, unauthenticated chat page that hands an
OpenAI model write access to your project and push access to your git remote, with
no confirmation step and only lexical path containment. It is not MCP. The
prototype MCP server that existed was deleted in 2025 and never shared a flag with
`-ai`.
