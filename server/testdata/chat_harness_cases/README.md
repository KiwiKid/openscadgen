The chat harness has two modes:

- `live`: calls the configured provider, then writes a local capture file to `tmp/chat-harness/<case>.json`
- `test-chat-tools`: reads that capture file back and replays any recorded tool calls through lightweight checker functions in [server/chat_harness_test.go](/Users/gregc/mine/making/3d-printing/openSCAD/openscadgen/server/chat_harness_test.go)

Use the included `just` recipes:

```bash
just chat-live stick-hinge-config-help
just chat-test-tools stick-hinge-config-help
```

Case file shape:

```json
{
  "name": "stick-hinge-config-help",
  "provider": "openai",
  "model": "gpt-4.1",
  "context_path": "examples/stick_hinge/config.toml",
  "prompt": "Review this config and suggest two concrete improvements.",
  "expect": {
    "require_text": true,
    "min_tool_calls": 0,
    "tool_names": []
  }
}
```

Notes:

- Keep `live` explicit. It requires `OPENSCADGEN_CHAT_HARNESS_CASE` so one test run does not fan out into multiple paid API calls.
- Captures stay local and are gitignored. The committed JSON files here are just reusable prompts plus expectations.
- The current `/chat` runtime executes project-scoped OpenAI file and git tools. `test-chat-tools` still only replays the captured tool calls you add checker functions for in the harness registry.
