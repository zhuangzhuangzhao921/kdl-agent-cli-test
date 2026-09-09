import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { mkdtempSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = fileURLToPath(new URL('../', import.meta.url));
const temp = mkdtempSync(join(tmpdir(), 'kdl-npm-'));
function command(cmd, args, options = {}) {
  const result = spawnSync(cmd, args, { cwd: root, encoding: 'utf8', ...options });
  if (result.error) throw result.error;
  assert.equal(result.status, 0, `${cmd}: ${result.stdout}\n${result.stderr}`);
  return result.stdout;
}
try {
  const [pack] = JSON.parse(command('npm', ['pack', '--json', '--pack-destination', temp]));
  for (const file of pack.files) {
    assert.ok(['package.json', 'README.md', 'LICENSE', 'THIRD_PARTY_NOTICES.txt'].includes(file.path)
      || /^(bin|native)\//.test(file.path), `Unexpected published file: ${file.path}`);
  }
  command('npm', ['install', '--prefix', join(temp, 'install'), '--ignore-scripts',
    '--no-audit', '--no-fund', join(temp, pack.filename)]);
  const entry = join(temp, 'install/node_modules/@zhuangzhuangzhao/kdl-agent-test/bin/kdl-agent-test.cjs');
  const result = command(process.execPath, ['--test', 'test/cli.test.mjs'], {
    env: { ...process.env, KDL_TEST_ENTRY: entry },
  });
  console.log(result);
  console.log(`Installed and verified ${pack.name}@${pack.version} (${pack.size} packed bytes).`);
} finally {
  rmSync(temp, { recursive: true, force: true });
}
