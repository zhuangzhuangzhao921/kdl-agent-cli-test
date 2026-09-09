import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { readFileSync, readdirSync, statSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { join } from 'node:path';

const root = fileURLToPath(new URL('../', import.meta.url));
const pkg = JSON.parse(readFileSync(join(root, 'package.json'), 'utf8'));
const manifest = JSON.parse(readFileSync(join(root, 'native/manifest.json'), 'utf8'));
const expected = ['darwin-arm64', 'darwin-x64', 'linux-arm64', 'linux-x64']
  .map((target) => `native/${target}/kdl-agent-test`);
assert.equal(manifest.version, pkg.version);
assert.deepEqual(manifest.files.map((file) => file.path).sort(), expected.sort());
for (const file of manifest.files) {
  const path = join(root, file.path);
  const bytes = readFileSync(path);
  assert.equal(bytes.length, file.size);
  assert.equal(createHash('sha256').update(bytes).digest('hex'), file.sha256);
  assert.ok(statSync(path).mode & 0o111, `${file.path} must be executable`);
}
const actual = readdirSync(join(root, 'native'), { recursive: true })
  .filter((path) => statSync(join(root, 'native', path)).isFile())
  .map((path) => `native/${path}`);
assert.deepEqual(actual.sort(), [...expected, 'native/manifest.json'].sort());
assert.ok(readFileSync(join(root, 'THIRD_PARTY_NOTICES.txt'), 'utf8').includes('Go runtime'));
console.log(`Verified ${pkg.name}@${pkg.version}: ${manifest.files.length} platform executables.`);
