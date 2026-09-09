import { spawnSync } from 'node:child_process';
import { createHash } from 'node:crypto';
import { chmodSync, existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = fileURLToPath(new URL('../', import.meta.url));
const pkg = JSON.parse(readFileSync(join(root, 'package.json'), 'utf8'));
const targets = [
  ['darwin', 'arm64', 'arm64'], ['darwin', 'x64', 'amd64'],
  ['linux', 'arm64', 'arm64'], ['linux', 'x64', 'amd64'],
];
const selected = process.argv.includes('--current')
  ? targets.filter(([os, arch]) => os === process.platform && arch === process.arch)
  : targets;
if (selected.length === 0) throw new Error(`Unsupported build host: ${process.platform}-${process.arch}`);

function go(args, env = {}) {
  const result = spawnSync('go', args, {
    cwd: root, env: { ...process.env, ...env }, encoding: 'utf8',
  });
  if (result.error) throw result.error;
  if (result.status !== 0) throw new Error(result.stderr || `go exited ${result.status}`);
  return result.stdout;
}

go(['mod', 'download']);
const notices = [];
for (const module of [
  'github.com/spf13/cobra', 'github.com/spf13/pflag',
  'github.com/inconshreveable/mousetrap', 'github.com/pelletier/go-toml/v2',
]) {
  const info = JSON.parse(go(['list', '-m', '-json', module]));
  const licenseFile = ['LICENSE', 'LICENSE.txt'].find((name) => existsSync(join(info.Dir, name)));
  if (!licenseFile) throw new Error(`License not found for ${module}`);
  const license = readFileSync(join(info.Dir, licenseFile), 'utf8');
  notices.push(`${info.Path} ${info.Version}\n\n${license}`);
}
const goroot = go(['env', 'GOROOT']).trim();
const goLicense = [join(goroot, 'LICENSE'), join(goroot, '..', 'LICENSE')].find(existsSync);
if (!goLicense) throw new Error('Go toolchain license not found');
notices.push(`Go runtime and standard library\n\n${readFileSync(goLicense, 'utf8')}`);
for (const name of ['crypto', 'net', 'sys', 'text']) {
  const path = join(goroot, 'src', 'vendor', 'golang.org', 'x', name, 'LICENSE');
  if (existsSync(path)) notices.push(`Go standard library dependency: golang.org/x/${name}\n\n${readFileSync(path, 'utf8')}`);
}
writeFileSync(join(root, 'THIRD_PARTY_NOTICES.txt'), notices.join('\n\n====================\n\n'));

const files = [];
for (const [os, nodeArch, goArch] of selected) {
  const path = `native/${os}-${nodeArch}/kdl-agent-test`;
  const output = join(root, path);
  mkdirSync(join(root, 'native', `${os}-${nodeArch}`), { recursive: true });
  go(['build', '-mod=readonly', '-trimpath', '-buildvcs=false',
    '-ldflags', `-s -w -X github.com/zhuangzhuangzhao921/kdl-agent-cli-test/cmd.version=${pkg.version}`,
    '-o', output, '.'], { GOOS: os, GOARCH: goArch, CGO_ENABLED: '0' });
  chmodSync(output, 0o755);
  const bytes = readFileSync(output);
  files.push({ path, size: bytes.length, sha256: createHash('sha256').update(bytes).digest('hex') });
  console.log(`Built ${path} (${bytes.length} bytes)`);
}
writeFileSync(join(root, 'native', 'manifest.json'), `${JSON.stringify({
  version: pkg.version, go: go(['version']).trim(), files,
}, null, 2)}\n`);
