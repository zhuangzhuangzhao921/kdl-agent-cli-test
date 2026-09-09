import assert from 'node:assert/strict';
import { spawn } from 'node:child_process';
import { once } from 'node:events';
import { mkdtemp, rm } from 'node:fs/promises';
import { createServer } from 'node:http';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import test from 'node:test';

const entry = process.env.KDL_TEST_ENTRY || fileURLToPath(new URL('../bin/kdl-agent-test.cjs', import.meta.url));
const { version } = JSON.parse(readFileSync(new URL('../package.json', import.meta.url), 'utf8'));
const sentinel = 'kdl_ag_test.not-a-real-credential';

function run(args, cwd, env = {}) {
  return new Promise((resolve, reject) => {
    const child = spawn(process.execPath, [entry, ...args], {
      cwd, env: { ...process.env, KDL_AGENT_CONFIG: join(cwd, 'test.toml'),
        KDL_AGENT_GATEWAY_URL: '', KDL_AGENT_TOKEN: '', NO_COLOR: '1', ...env },
    });
    let stdout = '', stderr = '';
    child.stdout.on('data', (data) => { stdout += data; });
    child.stderr.on('data', (data) => { stderr += data; });
    child.on('error', reject);
    child.on('close', (code) => resolve({ code, stdout, stderr }));
  });
}

test('installed command exposes help, subcommands, version, and errors', async (t) => {
  const cwd = await mkdtemp(join(tmpdir(), 'kdl-cli-'));
  t.after(() => rm(cwd, { recursive: true, force: true }));
  const result = await run(['--version'], cwd);
  assert.equal(result.code, 0);
  assert.equal(result.stdout.trim(), `kdl-agent-test ${version}`);
  for (const args of [['--help'], ['auth', '--help'], ['order', '--help'], ['proxy', '--help']]) {
    const help = await run(args, cwd);
    assert.equal(help.code, 0);
    assert.match(help.stdout, /kdl-agent-test/);
  }
  assert.notEqual((await run(['--unknown'], cwd)).code, 0);
  assert.equal((await run(['account', 'funds'], cwd)).code, 2);
});

test('installed command queries a Gateway and handles revoked credentials', async (t) => {
  const cwd = await mkdtemp(join(tmpdir(), 'kdl-http-'));
  t.after(() => rm(cwd, { recursive: true, force: true }));
  let calls = 0;
  const server = createServer((req, res) => {
    calls++;
    assert.equal(req.method, 'GET');
    assert.equal(req.url, '/v1/account/funds');
    assert.equal(req.headers.authorization, `Bearer ${sentinel}`);
    res.setHeader('Content-Type', 'application/json');
    if (calls === 1) {
      res.end(JSON.stringify({ success: true, data: { balance: '12.34' }, request_id: 'test-request' }));
    } else {
      res.statusCode = 401;
      res.end(JSON.stringify({ success: false, error: { code: 'CREDENTIAL_REVOKED', message: 'revoked' } }));
    }
  });
  server.listen(0, '127.0.0.1');
  await once(server, 'listening');
  t.after(() => new Promise((resolve) => server.close(resolve)));
  const env = { KDL_AGENT_GATEWAY_URL: `http://127.0.0.1:${server.address().port}`, KDL_AGENT_TOKEN: sentinel };
  const success = await run(['account', 'funds', '--format', 'json'], cwd, env);
  assert.equal(success.code, 0, success.stderr);
  assert.equal(JSON.parse(success.stdout).data.balance, '12.34');
  const failure = await run(['account', 'funds', '--format', 'json'], cwd, env);
  assert.equal(failure.code, 1);
  assert.match(failure.stderr, /CREDENTIAL_REVOKED/);
  assert.equal(calls, 2);
  assert.ok(!`${success.stdout}${success.stderr}${failure.stdout}${failure.stderr}`.includes(sentinel));
});
