# Wider-release plan

This plan separates release blockers from improvements that need design, review, and dedicated regression coverage. It intentionally does not add telemetry until the consent, data contract, and deletion behaviour are complete.

## Phase 0: release baseline

- [x] Run the existing Go test suite and remove the current `go vet` failure.
- [x] Run validation on pull requests and `main`, rather than only after a version tag exists.
- [x] Pin the template generator used by the Docker image and release workflow to the version declared in `go.mod`.
- [x] Correct the local-server port, Docker runtime, and bundled-install-guide inconsistencies.
- [ ] Protect a release branch/tag with the CI workflow and require a successful run before publishing.
- [ ] Add a short compatibility matrix for supported operating systems, OpenSCAD releases, and known limitations.

## Phase 1: safe local server

The application can edit files, delete exports, invoke desktop programs, and install developer tools. It should be treated as a local-development server, not an internet-facing service.

- [ ] Bind to loopback by default and make non-loopback listening an explicit, clearly labelled option.
- [ ] For remote operation, require authentication, CSRF protection for mutations, and a documented reverse-proxy deployment model.
- [ ] Enforce a canonical configured project root for every file read, edit, image/STL serve, delete, and process-launch route; reject paths outside it, including symlink escapes.
- [ ] Add route-level authorization tests and path-traversal/symlink regression tests.
- [ ] Add response security headers and a restrictive content-security policy that is compatible with the HTMX UI.
- [ ] Separate privileged tool installation from normal project-serving mode, or require a dedicated explicit flag.

## Phase 2: release engineering and project hygiene

- [ ] Pin BOSL2 to a reviewed commit in Docker and document the bundled dependency licences/notices.
- [ ] Generate checksums and an SBOM for every release; assess signing/notarisation for macOS and Windows before advertising the binaries broadly.
- [ ] Keep a curated example gallery in Git; move bulky generated render sets to release/site artefacts or Git LFS after deciding which are source-of-truth fixtures.
- [ ] Add `CONTRIBUTING.md`, `SECURITY.md`, issue templates, and a support path.
- [ ] Add a release smoke test that downloads/builds each target archive and checks `--version`, init, and a representative render where the platform permits it.

## Phase 3: first-run and support experience

- [ ] Add one guided starter project: create, change a parameter, preview, export.
- [ ] Add a diagnostics screen showing app/OpenSCAD/BOSL2 versions, project/output access, and actionable recent render failures.
- [ ] Make errors copyable and distinguish configuration, renderer, permissions, and output-path failures.
- [ ] Add a small manual release checklist covering fresh install, macOS/Windows trust prompts, Docker, and upgrade behaviour.

## Phase 4: optional, privacy-first telemetry

### Product contract

- [ ] Show a friendly consent prompt only after a successful first use (or from Settings), never during installation or before the app is useful.
- [ ] Default to off. Persist the choice locally; provide Settings controls to enable, disable, inspect, export, and delete local queued events.
- [ ] Publish `PRIVACY.md` before enabling collection. It must list every field, purpose, endpoint operator, retention period, and deletion process.
- [ ] Do not use a third-party analytics SDK, cookies, advertising identifiers, fingerprints, or stable device/user IDs.

### Minimum event schema

Each event is schema-versioned and contains only:

- application version;
- OS family and CPU architecture;
- OpenSCAD major/minor version;
- feature identifier (for example `project_init`, `render`, `export`, `tools_open`);
- outcome; and
- coarse duration and instance-count buckets.

Never collect project names, file paths, filenames, SCAD/TOML content, parameter values, generated artefacts, prompts, raw errors, IP addresses, or user identifiers. The ingestion service must discard request IPs rather than retaining them in application logs.

### Behaviour and operations

- [ ] Batch a small, rate-limited number of events over HTTPS; telemetry must never block rendering or exporting.
- [ ] Make every outbound request observable in the inspection UI and test that no request occurs without consent.
- [ ] Handle failures with a separate, user-initiated diagnostic report: show the exact scrubbed payload and require confirmation before sending.
- [ ] Aggregate on the server, restrict access, and delete raw events after a short published period (target: 30–90 days).
- [ ] Add contract tests for event allowlisting, path/error scrubbing, consent transitions, retry limits, and deletion.

Suggested consent copy:

> Would you like to share a tiny anonymous usage summary? I’m one developer, and it helps me fix bugs and focus on the features people use most. It never includes your designs, filenames, project settings, or personal details. You can review or change this any time.

Actions: **Share anonymous usage**, **Not now**, and **See exactly what is shared**.
