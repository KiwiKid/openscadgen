/* Browser render worker.
 *
 * One OpenSCAD/Emscripten instance per invocation, matching the Playground's
 * worker lifecycle. The compiled WebAssembly.Module is supplied by the page so
 * it is downloaded and compiled only once.
 */

let openscadFactoryPromise = null;

function post(payload, transfer) {
	self.postMessage(payload, transfer || []);
}

function logStream(stream, text) {
	if (!text) return;
	post({ type: "log", stream, text: String(text) });
}

function errorMessage(error) {
	if (error && typeof error === "object" && "message" in error) {
		return String(error.message);
	}
	return String(error);
}

function formatScadValue(value) {
	if (value === null || value === undefined) return "undef";
	if (typeof value === "string") {
		if (value === "true" || value === "false") return value;
		return JSON.stringify(value);
	}
	if (typeof value === "boolean") return value ? "true" : "false";
	if (typeof value === "number") return Number.isFinite(value) ? String(value) : "undef";
	if (Array.isArray(value)) return "[" + value.map(formatScadValue).join(", ") + "]";
	return JSON.stringify(value);
}

function ensureDir(fs, path) {
	const parts = path.split("/").filter(Boolean);
	let current = "";
	for (const part of parts) {
		current += "/" + part;
		try {
			fs.mkdir(current);
		} catch (error) {
			/* already exists */
		}
	}
}

function writeFile(fs, path, data) {
	const dir = path.slice(0, path.lastIndexOf("/"));
	if (dir) ensureDir(fs, dir);
	fs.writeFile(path, data);
}

async function sha256Hex(bytes) {
	const digest = await crypto.subtle.digest("SHA-256", bytes);
	return Array.from(new Uint8Array(digest), (byte) => byte.toString(16).padStart(2, "0")).join("");
}

async function loadLibrary(fs, library, assetsBase) {
	if (!library || !library.archive) {
		if (library && library.missing) {
			throw new Error(`${library.name || "Library"} is required by this design but is not bundled with this report.`);
		}
		return;
	}

	const archiveUrl = new URL(library.archive, assetsBase).href;
	logStream("stderr", `[browser] Loading library ${library.name} (${library.version || "unknown version"})`);
	const response = await fetch(archiveUrl);
	if (!response.ok) {
		throw new Error(`Failed to fetch library ${library.name}: HTTP ${response.status}`);
	}
	const compressed = await response.arrayBuffer();
	if (library.sha256) {
		const actual = await sha256Hex(compressed);
		if (actual !== library.sha256) {
			throw new Error(`Library ${library.name} failed its integrity check.`);
		}
	}

	const decompressed = await new Response(
		new Blob([compressed]).stream().pipeThrough(new DecompressionStream("gzip"))
	).text();
	const archive = JSON.parse(decompressed);
	const root = `/libraries/${library.name}`;
	ensureDir(fs, root);
	for (const [relativePath, content] of Object.entries(archive.files || {})) {
		writeFile(fs, `${root}/${relativePath}`, content);
	}

	// OpenSCAD resolves `include <BOSL2/std.scad>` from OPENSCADPATH and also
	// from the working directory. Provide both forms.
	const aliases = ["/" + library.name, "/" + library.name + ".scad"];
	for (const alias of aliases) {
		try {
			fs.symlink(root, alias);
		} catch (error) {
			/* alias already exists or is unsupported */
		}
	}
}

async function instantiate(wasmModule) {
	if (!openscadFactoryPromise) {
		openscadFactoryPromise = import(new URL("./openscad.js", self.location.href).href).then(
			(module) => module.default || module.OpenSCAD
		);
	}
	const factory = await openscadFactoryPromise;
	if (typeof factory !== "function") {
		throw new Error("OpenSCAD WebAssembly module did not expose a factory.");
	}

	const options = {
		noInitialRun: true,
		noExitRuntime: true,
		print: (text) => logStream("stdout", text),
		printErr: (text) => logStream("stderr", text),
		ENV: { OPENSCADPATH: "/libraries" },
	};
	if (wasmModule) {
		options.instantiateWasm = (imports, receiveInstance) => {
			WebAssembly.instantiate(wasmModule, imports)
				.then((result) => receiveInstance(result.instance, result.module))
				.catch((error) => logStream("stderr", `[browser] WASM instantiation failed: ${errorMessage(error)}`));
			return {};
		};
	}
	return factory(options);
}

async function runRender(request) {
	const { bundle, instance, preview, wasmModule, assetsBase } = request;
	const started = performance.now();
	const module = await instantiate(wasmModule);
	const fs = module.FS;

	const sources = new Map((bundle.sources || []).map((source) => [source.path, source.text]));
	const sourceText = sources.get(instance.sourcePath);
	if (typeof sourceText !== "string") {
		throw new Error(`Source ${instance.sourcePath} is not present in the report bundle.`);
	}

	for (const source of bundle.sources || []) {
		writeFile(fs, "/" + source.path, source.text);
	}

	const libraries = bundle.libraries || [];
	const usedSources = [sourceText, ...(bundle.sources || []).map((source) => source.text)].join("\n");
	const requiredLibraries = libraries.filter((library) => {
		const marker = library.name + "/";
		return usedSources.includes(marker);
	});
	for (const library of requiredLibraries) {
		await loadLibrary(fs, library, assetsBase);
	}

	const definitions = instance.definitions || {};
	const args = [
		"/" + instance.sourcePath,
		"-o",
		"/ogs-out.stl",
		"--backend=manifold",
		"--export-format=binstl",
	];
	if (!preview && instance.render !== false) args.push("--render");
	for (const [name, value] of Object.entries(definitions)) {
		if (name === "version") continue;
		args.push("-D", `${name}=${formatScadValue(value)}`);
	}

	if (preview) {
		// Matches the Playground's preview trick. It keeps the same geometry
		// pipeline but asks OpenSCAD for preview-mode evaluation.
		writeFile(fs, "/" + instance.sourcePath, "$preview=true;\n" + sourceText);
	}

	let code;
	try {
		code = module.callMain(args);
	} catch (error) {
		const formatted = module.formatException ? module.formatException(error) : errorMessage(error);
		throw new Error(`OpenSCAD render failed: ${formatted}`);
	}
	if (code !== 0) {
		throw new Error(`OpenSCAD exited with code ${code}`);
	}

	let artifact;
	try {
		artifact = module.FS.readFile("/ogs-out.stl");
	} catch (error) {
		throw new Error(`OpenSCAD did not produce an STL: ${errorMessage(error)}`);
	}
	if (!artifact || artifact.byteLength < 84) {
		throw new Error("OpenSCAD produced an empty or invalid STL.");
	}

	// Copy to a transferable buffer: Emscripten's view may point into WASM memory.
	const bytes = new Uint8Array(artifact);
	const copy = bytes.slice();
	post(
		{
			type: "result",
			id: instance.id,
			outputPath: instance.outputPath,
			artifact: copy,
			elapsedMs: performance.now() - started,
		},
		[copy.buffer]
	);
}

self.onmessage = async (event) => {
	const message = event.data || {};
	if (message.type !== "render") return;
	try {
		await runRender(message);
	} catch (error) {
		post({ type: "error", id: message.instance && message.instance.id, message: errorMessage(error) });
	}
};
