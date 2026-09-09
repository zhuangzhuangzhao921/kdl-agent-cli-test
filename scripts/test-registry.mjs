import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { mkdtempSync, readFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = fileURLToPath(new URL('../', import.meta.url));
const pkg = JSON.parse(readFileSync(join(root, 'package.json'), 'utf8'));
const temp = mkdtempSync(join(tmpdir(), 'kdl-registry-'));
try {
  const install = spawnSync('npm', ['install', '--prefix', temp,
    '--cache', join(temp, 'cache'), '--ignore-scripts', '--no-audit', '--no-fund',
    '--registry=https://registry.npmjs.org/', `${pkg.name}@${pkg.version}`],
  { cwd: root, encoding: 'utf8' });
  assert.equal(install.status, 0, install.stderr);
  const entry = join(temp, 'node_modules', pkg.name, 'bin/kdl-agent-test.cjs');
  const result = spawnSync(process.execPath, ['--test', 'test/cli.test.mjs'], {
    cwd: root, stdio: 'inherit', env: { ...process.env, KDL_TEST_ENTRY: entry },
  });
  assert.equal(result.status, 0);
  console.log(`Verified public npm installation: ${pkg.name}@${pkg.version} on ${process.platform}-${process.arch}.`);
} finally {
  rmSync(temp, { recursive: true, force: true });
}
