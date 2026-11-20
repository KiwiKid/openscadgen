#!/usr/bin/env node
/*
 * Downloads the pinned OpenSCAD WebAssembly build used by the browser report.
 *
 * The Playground publishes a reproducible build of OpenSCAD for the web
 * (Manifold backend, headless). We pin both the archive URL and its hash so a
 * report generated today does not silently use a different engine later.
 *
 * Usage: node scripts/fetch-browser-assets.mjs [--force]
 */

import { createHash } from "node:crypto";
import { mkdir, readFile, rm, writeFile } from "node:fs/promises";
import { createWriteStream, existsSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { inflateRawSync } from "node:zlib";

const WASM_BUILD_URL =
	"https://files.openscad.org/playground/OpenSCAD-2025.03.25.wasm24456-WebAssembly-web.zip";
const WASM_BUILD_SHA256 = "0968af31b9c9b3bba68d9031de1695ccae51c32231a1aab4ef27b18c86379f3b";
const EXPECTED_FILES = {
	"openscad.js": "904a47f29e63afb597bedef747da3b457d8ea17cc793c462c6c8b444e918a62e",
	"openscad.wasm": "f72ce246c02c0e501990837102be383326b153fd761774ebfacce5c80c5ecf26",
};

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const targetDir = path.join(repoRoot, "static", "browser");
const cacheDir = path.join(targetDir, ".cache");
const force = process.argv.includes("--force");

function sha256(bytes) {
	return createHash("sha256").update(bytes).digest("hex");
}

async function fileMatches(filePath, expected) {
	try {
		return sha256(await readFile(filePath)) === expected;
	} catch {
		return false;
	}
}

async function download(url, destination) {
	const response = await fetch(url);
	if (!response.ok) throw new Error(`Download failed: HTTP ${response.status} ${response.statusText}`);
	if (!response.body) throw new Error("Download failed: empty response body");
	await mkdir(path.dirname(destination), { recursive: true });
	const file = createWriteStream(destination);
	for await (const chunk of response.body) {
		if (!file.write(chunk)) {
			await new Promise((resolve) => file.once("drain", resolve));
		}
	}
	await new Promise((resolve, reject) => {
		file.end((error) => (error ? reject(error) : resolve()));
	});
}

// Minimal ZIP reader for the two files we need. Avoids a build dependency and
// works on Windows, macOS and Linux without requiring `unzip`.
function unzipEntries(buffer) {
	const entries = new Map();
	let eocd = -1;
	for (let i = buffer.length - 22; i >= Math.max(0, buffer.length - 65557); i--) {
		if (buffer.readUInt32LE(i) === 0x06054b50) {
			eocd = i;
			break;
		}
	}
	if (eocd < 0) throw new Error("Invalid ZIP archive: end of central directory not found");

	const entryCount = buffer.readUInt16LE(eocd + 10);
	let offset = buffer.readUInt32LE(eocd + 16);
	for (let index = 0; index < entryCount; index++) {
		if (buffer.readUInt32LE(offset) !== 0x02014b50) {
			throw new Error("Invalid ZIP archive: bad central directory header");
		}
		const method = buffer.readUInt16LE(offset + 10);
		const compressedSize = buffer.readUInt32LE(offset + 20);
		const uncompressedSize = buffer.readUInt32LE(offset + 24);
		const nameLength = buffer.readUInt16LE(offset + 28);
		const extraLength = buffer.readUInt16LE(offset + 30);
		const commentLength = buffer.readUInt16LE(offset + 32);
		const localOffset = buffer.readUInt32LE(offset + 42);
		const name = buffer.toString("utf8", offset + 46, offset + 46 + nameLength);

		if (buffer.readUInt32LE(localOffset) !== 0x04034b50) {
			throw new Error(`Invalid ZIP archive: bad local header for ${name}`);
		}
		const localNameLength = buffer.readUInt16LE(localOffset + 26);
		const localExtraLength = buffer.readUInt16LE(localOffset + 28);
		const dataStart = localOffset + 30 + localNameLength + localExtraLength;
		const compressed = buffer.subarray(dataStart, dataStart + compressedSize);
		let data;
		if (method === 0) {
			data = compressed;
		} else if (method === 8) {
			data = inflateRawSync(compressed);
		} else {
			throw new Error(`Unsupported ZIP compression method ${method} for ${name}`);
		}
		if (data.length !== uncompressedSize) {
			throw new Error(`Invalid ZIP archive: size mismatch for ${name}`);
		}
		entries.set(name, data);
		offset += 46 + nameLength + extraLength + commentLength;
	}
	return entries;
}

async function main() {
	if (!force) {
		const allPresent = await Promise.all(
			Object.entries(EXPECTED_FILES).map(([name, hash]) => fileMatches(path.join(targetDir, name), hash))
		);
		if (allPresent.every(Boolean)) {
			console.log("Browser OpenSCAD assets already present.");
			return;
		}
	}

	await mkdir(cacheDir, { recursive: true });
	const archivePath = path.join(cacheDir, "openscad-playground-wasm.zip");
	console.log(`Downloading ${WASM_BUILD_URL}`);
	await download(WASM_BUILD_URL, archivePath);

	const archive = await readFile(archivePath);
	const archiveHash = sha256(archive);
	if (archiveHash !== WASM_BUILD_SHA256) {
		await rm(archivePath, { force: true });
		throw new Error(`WASM archive hash mismatch: expected ${WASM_BUILD_SHA256}, got ${archiveHash}`);
	}

	const entries = unzipEntries(archive);
	for (const [name, expectedHash] of Object.entries(EXPECTED_FILES)) {
		const data = entries.get(name);
		if (!data) throw new Error(`WASM archive is missing ${name}`);
		const actualHash = sha256(data);
		if (actualHash !== expectedHash) {
			throw new Error(`${name} hash mismatch: expected ${expectedHash}, got ${actualHash}`);
		}
		await writeFile(path.join(targetDir, name), data);
		console.log(`Wrote static/browser/${name} (${data.length.toLocaleString()} bytes)`);
	}
}

main().catch((error) => {
	console.error(error.message || error);
	process.exitCode = 1;
});
