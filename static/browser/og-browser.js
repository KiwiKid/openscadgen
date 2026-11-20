/*
 * Optional browser renderer for generated reports.
 *
 * Design notes (following OpenSCAD Playground where practical):
 * - State lives in the URL fragment, gzip + base64 encoded. Query parameters
 *   only carry small selections/overrides.
 * - The WASM module is compiled once on the main thread and shared with each
 *   short-lived render worker (structured clone).
 * - OpenSCAD itself runs on the CPU. WebGL is used only for displaying and
 *   re-framing the generated mesh; it cannot accelerate CSG evaluation.
 */
(() => {
	const bundleElement = document.getElementById("og-browser-bundle");
	if (!bundleElement) return;

	const assetsBase = (window.OG_BROWSER_ASSETS_BASE || "./browser").replace(/\/$/, "");
	const elements = {
		open: document.getElementById("og-browser-open"),
		modal: document.getElementById("og-browser-modal"),
		close: document.getElementById("og-browser-close"),
		renderAll: document.getElementById("og-browser-render-all"),
		cancel: document.getElementById("og-browser-cancel"),
		preview: document.getElementById("og-browser-preview"),
		share: document.getElementById("og-browser-share"),
		status: document.getElementById("og-browser-status"),
		list: document.getElementById("og-browser-list"),
		log: document.getElementById("og-browser-log"),
		canvas: document.getElementById("og-browser-canvas"),
		viewerEmpty: document.getElementById("og-browser-viewer-empty"),
		fileWarning: document.getElementById("og-browser-file-warning"),
	};

	const state = {
		bundle: null,
		wasmModulePromise: null,
		workers: new Set(),
		rejects: new Set(),
		cancelled: false,
		viewer: null,
		viewerReady: null,
	};

	function setStatus(text) {
		if (elements.status) elements.status.textContent = text;
	}

	function appendLog(text) {
		if (!elements.log) return;
		const line = String(text || "").trimEnd();
		if (!line) return;
		elements.log.textContent += (elements.log.textContent ? "\n" : "") + line;
		elements.log.scrollTop = elements.log.scrollHeight;
	}

	function clearLog() {
		if (elements.log) elements.log.textContent = "";
	}

	function parseBundle() {
		try {
			return JSON.parse(bundleElement.textContent);
		} catch (error) {
			throw new Error(`Invalid browser bundle: ${error.message}`);
		}
	}

	function assetsUrl(path) {
		return new URL(assetsBase + "/" + path.replace(/^\//, ""), window.location.href).href;
	}

	function bytesToBase64(bytes) {
		let binary = "";
		const chunkSize = 0x8000;
		for (let i = 0; i < bytes.length; i += chunkSize) {
			binary += String.fromCharCode.apply(null, bytes.subarray(i, i + chunkSize));
		}
		return btoa(binary);
	}

	function base64ToBytes(base64) {
		const binary = atob(base64);
		const bytes = new Uint8Array(binary.length);
		for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
		return bytes;
	}

	async function gzipBase64(text) {
		if (typeof CompressionStream !== "function") {
			return "0." + bytesToBase64(new TextEncoder().encode(text));
		}
		const stream = new Blob([text]).stream().pipeThrough(new CompressionStream("gzip"));
		const buffer = await new Response(stream).arrayBuffer();
		return "1." + bytesToBase64(new Uint8Array(buffer));
	}

	async function gunzipBase64(value) {
		const dot = value.indexOf(".");
		if (dot < 0) throw new Error("Unknown state encoding");
		const version = value.slice(0, dot);
		const encoded = value.slice(dot + 1);
		if (version === "0") return new TextDecoder().decode(base64ToBytes(encoded));
		if (version !== "1") throw new Error(`Unsupported state version ${version}`);
		if (typeof DecompressionStream !== "function") throw new Error("This browser cannot decompress shared state");
		const bytes = base64ToBytes(encoded);
		const stream = new Blob([bytes]).stream().pipeThrough(new DecompressionStream("gzip"));
		return new Response(stream).text();
	}

	function stateForURL(bundle) {
		return {
			schema: bundle.schema,
			config: bundle.config,
			sources: bundle.sources,
			instances: bundle.instances,
			libraries: bundle.libraries,
			capabilities: bundle.capabilities,
		};
	}

	async function decodeURLValue(value) {
		if (value == null) return null;
		if (/^[01]\./.test(value)) return gunzipBase64(value);
		return value;
	}

	async function applyStateFromQuery() {
		const params = new URLSearchParams(window.location.search);
		const configValue = params.get("config") || params.get("cfg");
		const designValue = params.get("design") || params.get("src");
		if (!configValue && !designValue) return;

		try {
			if (configValue) {
				state.bundle.config = { ...state.bundle.config, TOML: await decodeURLValue(configValue) };
			}
			if (designValue) {
				const requestedPath = params.get("source") || params.get("path");
				const source =
					(state.bundle.sources || []).find((entry) => entry.path === requestedPath) ||
					(state.bundle.sources || [])[0];
				if (!source) throw new Error("The report has no design source to override");
				source.text = await decodeURLValue(designValue);
				source.sha256 = "";
			}
			appendLog("[browser] Loaded config/design overrides from the query string.");
		} catch (error) {
			appendLog(`[browser] Could not read query-string overrides: ${error.message}`);
			setStatus("Invalid query overrides");
		}
	}

	async function applyStateFromHash() {
		const hash = window.location.hash || "";
		if (!hash.startsWith("#og=")) return;
		try {
			const decoded = await gunzipBase64(hash.slice(4));
			const shared = JSON.parse(decoded);
			if (!shared || shared.schema !== state.bundle.schema) {
				throw new Error("Shared report state has an unsupported schema");
			}
			state.bundle = { ...state.bundle, ...shared };
			appendLog("[browser] Loaded config and design from the URL fragment.");
		} catch (error) {
			appendLog(`[browser] Could not read shared state: ${error.message}`);
			setStatus("Invalid share link");
		}
	}

	function selectedInstanceIds() {
		return Array.from(elements.list.querySelectorAll('input[type="checkbox"][data-instance-id]:checked')).map(
			(input) => input.dataset.instanceId
		);
	}

	function findInstance(id) {
		return (state.bundle.instances || []).find((instance) => instance.id === id);
	}

	function itemStatus(id, text, isError) {
		const item = elements.list.querySelector(`[data-instance-item="${CSS.escape(id)}"]`);
		if (!item) return;
		item.classList.toggle("is-error", !!isError);
		const status = item.querySelector(".og-browser-item-status");
		if (status) status.textContent = text;
	}

	function addDownload(id, url, filename) {
		const item = elements.list.querySelector(`[data-instance-item="${CSS.escape(id)}"]`);
		if (!item) return;
		const actions = item.querySelector(".og-browser-item-actions");
		if (!actions) return;
		let link = actions.querySelector("a[data-download]");
		if (!link) {
			link = document.createElement("a");
			link.dataset.download = "1";
			link.target = "_blank";
			link.rel = "noopener";
			link.textContent = "Download";
			actions.appendChild(link);
		}
		link.href = url;
		link.download = filename;
	}

	function renderList() {
		elements.list.textContent = "";
		const instances = state.bundle.instances || [];
		if (!instances.length) {
			elements.list.innerHTML = '<p class="og-browser-item-meta">No renderable instances in this report.</p>';
			return;
		}
		const query = new URLSearchParams(window.location.search);
		const requested = query.get("instance");
		for (const instance of instances) {
			const item = document.createElement("div");
			item.className = "og-browser-item";
			item.dataset.instanceItem = instance.id;

			const checkbox = document.createElement("input");
			checkbox.type = "checkbox";
			checkbox.checked = !requested || instance.name === requested || instance.autoName === requested || instance.id === requested;
			checkbox.dataset.instanceId = instance.id;

			const body = document.createElement("div");
			const name = document.createElement("div");
			name.className = "og-browser-item-name";
			name.textContent = instance.autoName || instance.name || instance.id;
			const meta = document.createElement("div");
			meta.className = "og-browser-item-meta";
			const params = Object.entries(instance.definitions || {})
				.map(([key, value]) => `${key}=${typeof value === "string" ? value : JSON.stringify(value)}`)
				.join(" · ");
			meta.textContent = [instance.sourcePath, params].filter(Boolean).join(" · ");
			body.append(name, meta);

			const actions = document.createElement("div");
			actions.className = "og-browser-item-actions";
			const status = document.createElement("span");
			status.className = "og-browser-item-status";
			status.textContent = "queued";
			actions.appendChild(status);

			item.append(checkbox, body, actions);
			elements.list.appendChild(item);
		}
	}

	function getWasmModule() {
		if (!state.wasmModulePromise) {
			setStatus("Loading OpenSCAD WASM…");
			const url = assetsUrl("openscad.wasm");
			state.wasmModulePromise = fetch(url)
				.then(async (response) => {
					if (!response.ok) throw new Error(`HTTP ${response.status} loading ${url}`);
					try {
						return await WebAssembly.compileStreaming(response.clone());
					} catch (error) {
						return WebAssembly.compile(await response.arrayBuffer());
					}
				})
				.catch((error) => {
					state.wasmModulePromise = null;
					throw new Error(
						`Could not load the OpenSCAD WebAssembly engine. Serve the report over HTTP and keep the browser/ folder beside it. (${error.message})`
					);
				});
		}
		return state.wasmModulePromise;
	}

	function renderOne(instance, preview) {
		return getWasmModule().then(
			(wasmModule) =>
				new Promise((resolve, reject) => {
					if (state.cancelled) {
						reject(new Error("cancelled"));
						return;
					}
					const worker = new Worker(assetsUrl("og-browser-worker.js"));
					state.workers.add(worker);
					state.rejects.add(reject);

					const cleanup = () => {
						worker.terminate();
						state.workers.delete(worker);
						state.rejects.delete(reject);
					};

					worker.onmessage = (event) => {
						const message = event.data || {};
						if (message.type === "log") {
							appendLog(message.text);
							return;
						}
						if (message.type === "result") {
							cleanup();
							resolve(message);
							return;
						}
						if (message.type === "error") {
							cleanup();
							reject(new Error(message.message || "Browser render failed"));
						}
					};
					worker.onerror = (event) => {
						cleanup();
						reject(new Error(event.message || "Browser worker failed"));
					};
					worker.postMessage({
						type: "render",
						bundle: state.bundle,
						instance,
						preview,
						wasmModule,
						assetsBase: new URL(assetsBase + "/", window.location.href).href,
					});
				})
		);
	}

	async function renderSelected() {
		if (location.protocol === "file:") {
			setStatus("Serve this report over HTTP to use the browser renderer");
			appendLog("[browser] Workers and WebAssembly are blocked on file:// by most browsers.");
			return;
		}
		const ids = selectedInstanceIds();
		if (!ids.length) {
			setStatus("Select at least one instance");
			return;
		}
		if (state.workers.size) return;

		const preview = !!(elements.preview && elements.preview.checked);
		elements.renderAll.disabled = true;
		elements.cancel.disabled = false;
		clearLog();
		state.cancelled = false;
		setStatus(`Rendering 0/${ids.length}…`);

		// Instances are independent. A small worker pool uses spare CPU cores
		// without multiplying the WASM memory footprint without limit. The
		// compiled module itself is shared by every worker.
		const concurrency = Math.max(
			1,
			Math.min(ids.length, 4, Math.max(1, (navigator.hardwareConcurrency || 4) - 1))
		);
		let cursor = 0;
		let completed = 0;

		const runNext = async () => {
			while (!state.cancelled) {
				const index = cursor++;
				if (index >= ids.length) return;
				const id = ids[index];
				const instance = findInstance(id);
				if (!instance) continue;

				itemStatus(id, preview ? "previewing…" : "rendering…");
				try {
					const result = await renderOne(instance, preview);
					const blob = new Blob([result.artifact], { type: "model/stl" });
					const url = URL.createObjectURL(blob);
					addDownload(id, url, instance.outputPath || "model.stl");
					itemStatus(id, `${Math.round(result.elapsedMs)} ms`);
					try {
						await showInViewer(result.artifact);
					} catch (viewerError) {
						// WebGL is a convenience; a headless or GPU-less browser must not
						// turn a successful geometry render into a failure.
						appendLog(`[browser] Preview unavailable: ${viewerError.message}`);
					}
					completed++;
					setStatus(`${completed}/${ids.length} rendered`);
				} catch (error) {
					if (state.cancelled) {
						itemStatus(id, "cancelled", true);
						return;
					}
					itemStatus(id, "error", true);
					appendLog(`[browser] ${error.message}`);
				}
			}
		};

		await Promise.all(Array.from({ length: concurrency }, () => runNext()));

		elements.renderAll.disabled = false;
		elements.cancel.disabled = true;
		if (state.cancelled) {
			setStatus(`Cancelled after ${completed}/${ids.length}`);
		} else {
			setStatus(`Done: ${completed}/${ids.length} rendered`);
		}
	}

	function cancelRender() {
		state.cancelled = true;
		for (const worker of state.workers) worker.terminate();
		state.workers.clear();
		for (const reject of state.rejects) reject(new Error("cancelled"));
		state.rejects.clear();
		setStatus("Cancelled");
	}

	async function copyShareLink() {
		try {
			const encoded = await gzipBase64(JSON.stringify(stateForURL(state.bundle)));
			const base = window.location.href.split("#")[0].split("?")[0];
			const link = `${base}#og=${encoded}`;
			if (link.length > 100000) {
				throw new Error("This design is too large for a share link; the embedded report still works.");
			}
			if (navigator.clipboard && navigator.clipboard.writeText) {
				await navigator.clipboard.writeText(link);
				setStatus(`Share link copied (${Math.round(link.length / 1024)} KB)`);
			} else {
				window.prompt("Copy this share link", link);
			}
		} catch (error) {
			setStatus("Could not create share link");
			appendLog(`[browser] ${error.message}`);
		}
	}

	function loadScript(url) {
		return new Promise((resolve, reject) => {
			const existing = document.querySelector(`script[src="${url}"]`);
			if (existing) {
				if (existing.dataset.loaded === "1") resolve();
				else {
					existing.addEventListener("load", resolve, { once: true });
					existing.addEventListener("error", reject, { once: true });
				}
				return;
			}
			const script = document.createElement("script");
			script.src = url;
			script.async = true;
			script.onload = () => {
				script.dataset.loaded = "1";
				resolve();
			};
			script.onerror = () => reject(new Error(`Failed to load ${url}`));
			document.head.appendChild(script);
		});
	}

	async function ensureViewer() {
		if (state.viewerReady) return state.viewerReady;
		state.viewerReady = (async () => {
			await loadScript("https://cdn.jsdelivr.net/npm/three@0.128.0/build/three.min.js");
			await loadScript("https://cdn.jsdelivr.net/npm/three@0.128.0/examples/js/loaders/STLLoader.js");
			await loadScript("https://cdn.jsdelivr.net/npm/three@0.128.0/examples/js/controls/OrbitControls.js");

			const THREE = window.THREE;
			if (!THREE) throw new Error("three.js did not load");

			const renderer = new THREE.WebGLRenderer({
				canvas: elements.canvas,
				antialias: true,
				powerPreference: "high-performance",
			});
			renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2));
			renderer.outputEncoding = THREE.sRGBEncoding;

			const scene = new THREE.Scene();
			scene.background = new THREE.Color(0x0b1118);
			const camera = new THREE.PerspectiveCamera(45, 1, 0.01, 10000);
			camera.position.set(40, 30, 40);

			const controls = new THREE.OrbitControls(camera, renderer.domElement);
			controls.enableDamping = true;
			controls.dampingFactor = 0.08;

			scene.add(new THREE.AmbientLight(0xffffff, 0.55));
			const key = new THREE.DirectionalLight(0xffffff, 0.9);
			key.position.set(60, 80, 50);
			scene.add(key);
			const fill = new THREE.DirectionalLight(0x93c5fd, 0.35);
			fill.position.set(-50, -30, -60);
			scene.add(fill);

			const group = new THREE.Group();
			scene.add(group);

			const viewer = {
				THREE,
				renderer,
				scene,
				camera,
				controls,
				group,
				mesh: null,
				frame: 0,
				animate() {
					viewer.frame = requestAnimationFrame(() => viewer.animate());
					controls.update();
					renderer.render(scene, camera);
				},
			};
			state.viewer = viewer;

			const resize = () => {
				const width = elements.canvas.clientWidth || 1;
				const height = elements.canvas.clientHeight || 1;
				renderer.setSize(width, height, false);
				camera.aspect = width / height;
				camera.updateProjectionMatrix();
			};
			resize();
			if (window.ResizeObserver) new ResizeObserver(resize).observe(elements.canvas.parentElement);
			else window.addEventListener("resize", resize);
			viewer.animate();
			return viewer;
		})();
		return state.viewerReady;
	}

	async function showInViewer(arrayBuffer) {
		const viewer = await ensureViewer();
		const { THREE, group, camera, controls } = viewer;

		const geometry = new THREE.STLLoader().parse(arrayBuffer);
		if (!geometry || !geometry.attributes.position || geometry.attributes.position.count === 0) {
			throw new Error("The generated STL contains no geometry");
		}
		geometry.computeVertexNormals();
		geometry.computeBoundingBox();

		if (viewer.mesh) {
			group.remove(viewer.mesh);
			viewer.mesh.geometry.dispose();
			viewer.mesh.material.dispose();
		}
		const material = new THREE.MeshStandardMaterial({
			color: 0x67d5c5,
			metalness: 0.15,
			roughness: 0.55,
			flatShading: false,
		});
		viewer.mesh = new THREE.Mesh(geometry, material);
		group.add(viewer.mesh);

		const box = geometry.boundingBox;
		const center = box.getCenter(new THREE.Vector3());
		const size = box.getSize(new THREE.Vector3());
		const maxDimension = Math.max(size.x, size.y, size.z) || 1;
		group.position.copy(center).multiplyScalar(-1);
		camera.near = Math.max(maxDimension / 1000, 0.01);
		camera.far = maxDimension * 100;
		camera.position.set(maxDimension * 1.6, maxDimension * 1.2, maxDimension * 1.6);
		camera.updateProjectionMatrix();
		controls.target.set(0, 0, 0);
		controls.update();

		if (elements.viewerEmpty) elements.viewerEmpty.style.display = "none";
	}

	function openModal() {
		elements.modal.hidden = false;
		if (location.protocol === "file:") {
			elements.fileWarning.textContent =
				"file:// blocks workers; serve this folder over HTTP (for example with openscadgen -sf).";
		}
		setStatus(`${(state.bundle.instances || []).length} instances ready`);
		if (!elements.list.childElementCount) renderList();
	}

	function closeModal() {
		elements.modal.hidden = true;
	}

	async function init() {
		try {
			state.bundle = parseBundle();
		} catch (error) {
			console.error(error);
			return;
		}
		await applyStateFromHash();
		await applyStateFromQuery();
		renderList();

		if (elements.open) elements.open.addEventListener("click", openModal);
		if (elements.close) elements.close.addEventListener("click", closeModal);
		if (elements.renderAll) elements.renderAll.addEventListener("click", renderSelected);
		if (elements.cancel) elements.cancel.addEventListener("click", cancelRender);
		if (elements.share) elements.share.addEventListener("click", copyShareLink);
		if (elements.modal) {
			elements.modal.addEventListener("click", (event) => {
				if (event.target === elements.modal) closeModal();
			});
		}
		document.addEventListener("keydown", (event) => {
			if (event.key === "Escape" && elements.modal && !elements.modal.hidden) closeModal();
		});

		// Preload only when the user shows intent; do not slow the report page.
		if (elements.open) {
			elements.open.addEventListener(
				"pointerenter",
				() => {
					getWasmModule().catch(() => {});
				},
				{ once: true }
			);
		}
	}

	if (document.readyState === "loading") {
		document.addEventListener("DOMContentLoaded", init, { once: true });
	} else {
		init();
	}
})();
