#!/usr/bin/env node
// Rebuild the WASM artifacts shipped in this package from the surrounding
// source tree.
//
// npm/dwscript.wasm is a build output that happens to be tracked in git. It is
// authoritative only for the commit that produced it, and the release workflow
// (.github/workflows/npm-publish.yml) always rebuilds it before publishing.
// Anyone consuming npm/ directly -- `npm pack`, `npm publish` from a checkout,
// or a Git dependency -- would otherwise get whatever binary happened to be
// committed last, which silently predates the Go sources next to it.
//
// This script closes that gap: whenever the Go source tree is reachable, the
// binary is rebuilt so the packaged artifact always matches the checkout.
// Installs from the npm registry never run it (npm runs "prepare" only for Git
// dependencies and for packing/publishing), and it no-ops if the sources are
// not there.
//
// Escape hatch: set DWSCRIPT_SKIP_WASM_BUILD=1 to keep the tracked binary.

import { execFileSync } from 'node:child_process';
import { existsSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const packageDir = dirname(dirname(fileURLToPath(import.meta.url)));
const repoRoot = dirname(packageDir);

const prefix = '[dwscript] ';
const log = (message) => console.log(prefix + message);

if (process.env.DWSCRIPT_SKIP_WASM_BUILD === '1') {
  log('DWSCRIPT_SKIP_WASM_BUILD=1 — keeping the existing dwscript.wasm.');
  process.exit(0);
}

const buildScript = join(repoRoot, 'build', 'wasm', 'build.sh');
const wasmMain = join(repoRoot, 'cmd', 'dwscript-wasm');

// No source tree: this is an installed tarball, so the bundled binary is all
// there is and it is the right one.
if (!existsSync(join(repoRoot, 'go.mod')) || !existsSync(buildScript) || !existsSync(wasmMain)) {
  log('No Go source tree next to the package — using the bundled dwscript.wasm.');
  process.exit(0);
}

// From here on the sources are present, so shipping the tracked binary would
// mean shipping a stale one. Fail loudly rather than do that silently.
try {
  execFileSync('go', ['version'], { stdio: 'ignore' });
} catch {
  console.error(
    prefix +
      'Go is required to package @cwbudde/dwscript from a source checkout: ' +
      'the tracked npm/dwscript.wasm is a build output and is generally older ' +
      'than the sources beside it. Install Go 1.24+ (https://go.dev/dl/), or set ' +
      'DWSCRIPT_SKIP_WASM_BUILD=1 to accept the tracked binary as-is.',
  );
  process.exit(1);
}

log('Building dwscript.wasm from source...');
execFileSync('bash', [buildScript, 'monolithic'], { cwd: repoRoot, stdio: 'inherit' });

const dist = join(repoRoot, 'build', 'wasm', 'dist');
for (const name of ['dwscript.wasm', 'wasm_exec.js']) {
  const source = join(dist, name);
  if (!existsSync(source)) {
    console.error(prefix + `Build did not produce ${source}.`);
    process.exit(1);
  }
  // Plain read/write rather than copyFileSync: the destination is a tracked
  // file and copy-on-write copies fail on some sandboxes and network mounts.
  const bytes = readFileSync(source);
  writeFileSync(join(packageDir, name), bytes);
  log(`Updated ${name} (${bytes.length} bytes).`);
}
