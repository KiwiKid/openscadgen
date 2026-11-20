# Browser-rendered report plan

Status: first implementation landed.

Implemented:

- `-browser-report` / `-br`, `pkg/reportbundle` data contract (`pkg/browser_report.go`),
  `browser/` sidecar assets, `bundle.json`.
- `static/browser/og-browser.js` + `og-browser-worker.js`: worker-based
  OpenSCAD-WASM rendering, one instance per worker, shared compiled
  `WebAssembly.Module`, Manifold backend, preview mode, cancel, download.
- A small worker pool (up to 4, or CPU count - 1) for parallel instance
  renders.
- WebGL preview with lazy-loaded three.js; the viewer never blocks a render.
- `#og=` gzip/base64 share links; `?instance=` selection only.
- BOSL2 is packaged on demand as a gzipped JSON archive and loaded only when a
  source refers to `BOSL2/`.
- Tests for bundle construction, path normalization, asset copying and library
  packaging.

Still planned: in-browser parameter editing, general dependency discovery, and
a browser smoke test in CI. Query-string overrides for small
`?config=`/`?design=` payloads are implemented; the fragment remains the
canonical form for complete designs.

## Answer to "do we have a plan?"

No documented plan covers this. `PLAN.md` is about release hardening, and the
README TODO list does not mention browser rendering. What does exist is a
deliberate rendering seam that was written for this kind of adapter.

## Groundwork already in the code

- `pkg/renderer.go` defines the portable contract: `RenderRequest` carries
  `Source []byte`, definitions, output path and camera; `RenderResponse`
  carries an optional `Artifact []byte`. `materializeRenderArtifact` writes an
  in-memory artifact when an engine has no filesystem.
- `RenderEngine` is the execution seam. The only current implementation is
  `OpenSCADRenderer`, which shells out to the local OpenSCAD binary.
- `ProcessWithRenderer` in `pkg/processor.go` is already documented as the
  entrypoint for embedded, remote, or browser/WASM renderers.
- `Config.RawConfigFile` retains the exact TOML that was loaded.
- `pkg/source_archive.go` already collects the config plus each configured
  `.scad` input into one replaceable `source.zip`.
- `pkg/templates/report.templ` is the report UI, and `reportMode` already
  supports static (`pages`), server, view and complete variants.

So the host-side boundary is in place. The missing pieces are a browser bundle,
an OpenSCAD WASM adapter, and a report UI mode that does not need the server.

## Performance model (GPU, workers, dynamic imports)

- **WebGL cannot accelerate OpenSCAD's geometry evaluation.** Manifold and CGAL
  are CPU libraries compiled to WebAssembly. WebGL is used for the interactive
  preview, camera movement and (future) PNG rasterisation of a finished mesh.
  Do not market GPU acceleration for CSG; the honest win is Manifold plus
  worker parallelism.
- **Manifold first.** The worker passes `--backend=manifold`, matching the
  Playground. It is usually the largest single speed-up over CGAL.
- **Share the compiled module.** The page downloads and compiles
  `openscad.wasm` once (`WebAssembly.compileStreaming`) and posts the
  `WebAssembly.Module` to each worker via structured clone. Workers then only
  instantiate, not re-download or re-compile.
- **Parallel instances.** Instances are independent, so a bounded worker pool
  runs up to four at once (or CPU count - 1). This is the best use of browser
  hardware for a multi-instance report. Workers are short-lived because one
  Emscripten instance is used per render.
- **Dynamic imports, deliberately.** `og-browser.js` is tiny and deferred; the
  page does not load the 9.6 MB WASM or three.js until the user opens the panel
  (WASM is fetched on pointer hover or first render, three.js only after a
  successful render). The worker dynamically imports `openscad.js`, so a user
  who never uses browser rendering pays no WASM cost. Do not split the core UI
  into more dynamic chunks; that adds latency without reducing the dominant
  WASM download.
- **Preview before final.** Quick preview adds `$preview=true;` and omits
  `--render`; uncheck the option for a full render. The Playground uses the
  same preview trick.
- **Future:** SIMD or pthreads builds of OpenSCAD-WASM could help, but threads
  require cross-origin isolation (COOP/COEP), which file:// and many static
  hosts cannot provide. WebGPU is not a fit for Manifold today. OffscreenCanvas
  would help rasterising previews, not geometry.

## Constraints that shape the design

1. **OpenSCAD WASM is headless.** It can export STL, 3MF, CSG, OBJ, OFF, WRL,
   SVG/DXF and text formats, but not PNG or PDF because the build has no
   OpenGL and no Cairo. Browser mode can therefore re-render geometry, but it
   cannot reproduce the report's PNG previews or image-based tests. Those stay
   host-generated or become a three.js preview of the generated mesh.
   See the [OpenSCAD WASM output-format documentation](https://loumalouomega.github.io/openscad-wasm/guide/formats)
   and the upstream [openscad-wasm README](https://github.com/openscad/openscad-wasm).
2. **A report opened with `file://` cannot reliably load sidecar ES modules or
   workers.** Chrome treats `file://` as an opaque origin and blocks module and
   worker scripts. A single portable HTML file needs an inlined classic worker
   plus an embedded WASM binary (large), or the report must be served over
   HTTP. This is the biggest packaging decision and needs a spike before the
   UI work.
3. **The report is not offline today.** `HeadLinks()` loads Bulma, Bootstrap,
   Font Awesome, HTMX and Alpine from CDNs. A browser-rendered report should
   vendor what it needs, otherwise "works from a saved file" is misleading.
4. **Direct inputs are not the whole dependency graph.** `source.zip` contains
   the config and configured inputs, but not BOSL2, `use`/`include` files,
   fonts or imported assets. A browser bundle needs an explicit, hash-pinned
   dependency manifest, or browser rendering will fail on real projects.
5. **URL length is finite.** Full SCAD source belongs in the document or in a
   fragment, not in a normal query string. Query strings are also logged by
   servers and proxies.
6. **Licensing.** This repository is AGPL-3.0. OpenSCAD and openscad-wasm are
   GPL-family. Shipping a WASM build inside report artifacts needs a licence
   notice and a source-offer review; it is probably compatible, but it should
   be a deliberate step rather than an accident.
7. **Two implementations drift.** If browser mode re-derives instances from
   TOML in TypeScript while the CLI derives them in Go, parameter expansion,
   numbering, ignoring and naming will diverge. Prefer one source of truth.

## Recommended product shape

Two layers, delivered incrementally:

1. **Re-render the generated report in the browser.** The Go run already knows
   the exact instances, applied parameters, cameras and output names. Embed that
   as a render plan, then let the browser run OpenSCAD WASM over the same plan.
   This needs no config parser in JavaScript and gives exact parity for the
   common case.
2. **Share or edit a browser report through the URL.** Serialize the config and
   design into a compact bundle for links; support small overrides in the query
   string. Full in-browser config editing is a later phase.

## Shared interface

Keep the browser adapter a direct mirror of the Go contract. The browser
implementation should not expose OpenSCAD command-line strings; it accepts the
same typed request.

```ts
type RenderKind = "stl" | "image";

type RenderRequest = {
  kind: RenderKind;
  sourcePath: string;
  source: Uint8Array;
  outputPath: string;
  definitions: Record<string, unknown>;
  render: boolean;
  debug: boolean;
  camera?: { name: string; coordinates: string; imageSize: string };
};

type RenderResponse = {
  command?: string;
  output: string;
  durationMs: number;
  artifact?: Uint8Array;
};

interface RenderEngine {
  render(request: RenderRequest, signal: AbortSignal): Promise<RenderResponse>;
}
```

`OpenSCADWasmEngine` returns `kind: "image"` as an explicit unsupported error,
the same way a renderer without camera support should. The host report and any
future `/api/render` endpoint can accept the same JSON shape.

## Report bundle

Add a versioned JSON payload to `report.html`, either inline as
`<script type="application/json" id="og-render-bundle">` or as a sidecar
`report.bundle.json` when size matters.

```json
{
  "schema": 1,
  "generator": { "name": "openscadgen", "version": "..." },
  "config": { "path": "config.toml", "toml": "..." },
  "sources": [
    { "path": "clip.scad", "sha256": "...", "text": "..." }
  ],
  "libraries": [
    { "name": "BOSL2", "version": "...", "url": "...", "sha256": "..." }
  ],
  "instances": [
    {
      "id": "...",
      "name": "standard",
      "sourcePath": "clip.scad",
      "definitions": { "width": 10, "height": 5 },
      "outputPath": "clip-10mm-5mm.stl",
      "render": true
    }
  ],
  "images": []
}
```

Rules:

- The bundle is the source of truth for browser rendering; the render plan is
  generated by the same Go code that generated the instances, so naming and
  parameter expansion cannot drift.
- `source.toml` is the exact raw TOML. Do not regenerate it from the Go struct.
- Hash every source and library. The browser should refuse to silently use a
  different file than the one that produced the report.
- Do not put PNG/STL bytes in the JSON. Those already exist beside the report.
- Keep the schema versioned and forward-compatible. Unknown fields are ignored;
  unsupported schema versions show a clear message.

## URL design

Use the query string for selection and small overrides, and the fragment for
the bundle:

```text
report.html?view=browser&instance=standard#bundle=1.<base64url(gzip(json))>
report.html?view=browser&config=<base64url(gzip(toml))>&design=<base64url(gzip(scad))>
```

Recommendation: make the fragment the canonical share form.

- Fragments are not sent to the server, so a shared report does not leak the
  design into access logs.
- Fragments avoid most proxy/query-length limits.
- `CompressionStream("gzip")` and `DecompressionStream("gzip")` are available
  in current browsers; include a raw-base64 fallback.
- Keep a conservative cap (for example 8 KB for query, 64 KB for fragment) and
  show a useful message plus a "download a copy of the report" action when the
  design is too large. Do not silently truncate.
- If `config`/`design` are present in the query, they override the embedded
  bundle for that session only. `?instance=` and `?param.width=` are allowed as
  small overrides. Never evaluate query values as code.
- Support a sidecar form when served over HTTP:
  `?config=./config.toml&design=./clip.scad`, using `fetch` and the same size
  checks. This is the clean option for a project folder that is already served.

## Implementation phases

### Phase 1: render bundle (Go only)

- Add `pkg/reportbundle.go` with `BuildReportBundle(config, instances)`.
- Include the raw config, every configured source, dependency manifest and the
  resolved render plan. Exclude image-only and skipped instances, but record
  why they were excluded.
- Embed the bundle in `report.html` behind a report option, for example
  `Operations.BrowserRendering` or a CLI flag.
- Add generated-at, openscadgen version, and source hashes.
- Tests: golden JSON for one and multiple inputs, param sets, ignored params,
  numbering, and a config with a missing include.

### Phase 2: browser engine and UI

- Add a small `web/` build (esbuild or similar) that produces one worker
  bundle: OpenSCAD WASM glue plus `OpenSCADWasmEngine`.
- Pin the OpenSCAD WASM version and record its hash. Do not load an unpinned
  latest build at runtime.
- Populate the Emscripten virtual FS from the bundle; create directories for
  nested source paths and libraries.
- Add "Render in browser" at the report level and per instance. Reuse the
  existing instance cards and progress language. Provide cancel by terminating
  the worker and starting a fresh one.
- Download generated STL/3MF artifacts individually or as a zip. Show render
  diagnostics and elapsed time in the same fields the CLI report already uses.
- Explicitly disable or hide image rendering and image tests in this mode, with
  a one-line explanation.

### Phase 3: share links

- Add encode/decode for the fragment payload and the query overrides.
- Add "Copy browser link" and "Download browser-ready report".
- Handle: absent bundle, bad base64, wrong schema, hash mismatch, oversized
  payload, `file://` restrictions.
- Test the URL codec as a pure Go/JS pair with shared fixtures, including
  Unicode parameter values and compressed edge cases.

### Phase 4: in-browser editing (optional, larger)

- Reuse the report's parameter form and instance cards.
- Decide between two implementations:
  - **Port expansion to TypeScript.** Smaller download, but must be kept in
    lockstep with Go.
  - **Compile the Go pipeline to `js/wasm`.** Preserves parity, but needs a
    filesystem shim and a larger bundle, and `ProcessWithRenderer` currently
    writes export folders and source archives.
- Do not start this until Phase 1-3 are stable. "Re-render what was generated"
  delivers most of the value with a fraction of the risk.

### Phase 5: previews and packaging

- Prefer a three.js viewer over generated STL for interactive geometry. Keep
  host-generated PNGs as the static fallback.
- Vendor CSS/JS for offline use, or publish a `report-assets/` folder beside the
  HTML. Document the HTTP requirement for the non-single-file form.
- Add licence notices for the WASM build and any bundled libraries.
- CI: build the worker, run a headless browser smoke test that loads the report,
  renders one cube, and checks the resulting STL hash.

## Error and compatibility policy

- Browser mode is additive. The existing report must remain fully usable with
  JavaScript disabled.
- If WASM, worker, or language features are unavailable, show a capability
  message and keep the static report. Do not block page load.
- Never silently fall back from a requested browser render to a different
  source. A hash mismatch is an error.
- Treat the browser as untrusted for host access: the bundle is data, not code.
  Do not enable arbitrary `custom_openscad_args` or custom executable paths in
  browser mode.

## Open decisions

1. Is the goal re-rendering existing results, arbitrary config editing, or both?
2. Is a single portable `report.html` a requirement, or is serving the report
   over HTTP acceptable? This decides the worker/WASM packaging.
3. Inline bundle versus sidecar JSON. Inline keeps one file; sidecar keeps the
   HTML readable and small.
4. How are BOSL2, fonts and imported assets resolved: bundled, fetched by hash,
   or only supported when the report is served from the project folder?
5. TS reimplementation of instance expansion versus Go compiled to WASM.
6. Are browser-generated PNG previews out of scope permanently, or should a
   mesh-to-image path be added later?

## Suggested first spike

Before building UI, prove the risky part in one afternoon:

- Load a pinned openscad-wasm build in a worker from a served page.
- Write one SCAD file and a BOSL2 `include` into the virtual FS.
- Render `cube(10);` and a BOSL2 `cuboid` to binary STL.
- Repeat from `file://` with an inlined worker and embedded WASM bytes, and
  record what the major browsers actually allow.

That result determines whether Phase 2 targets a single-file report, a report
folder, or a small local report server.
