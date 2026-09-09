#!/usr/bin/env node
'use strict';

const { spawn } = require('node:child_process');
const { join } = require('node:path');
const { constants } = require('node:os');

const targets = new Set(['darwin-arm64', 'darwin-x64', 'linux-arm64', 'linux-x64']);
const target = `${process.platform}-${process.arch}`;

if (!targets.has(target)) {
  console.error(`kdl-agent-test: unsupported platform ${target}. Supported: ${[...targets].join(', ')}.`);
  process.exitCode = 1;
} else {
  const binary = join(__dirname, '..', 'native', target, 'kdl-agent-test');
  const child = spawn(binary, process.argv.slice(2), { stdio: 'inherit' });
  const forward = (signal) => {
    if (child.pid) child.kill(signal);
  };
  process.on('SIGINT', forward);
  process.on('SIGTERM', forward);
  child.on('error', (error) => {
    console.error(`kdl-agent-test: could not start the executable (${error.code || 'UNKNOWN'}). Reinstall the npm package.`);
    process.exitCode = 1;
  });
  child.on('close', (code, signal) => {
    process.removeListener('SIGINT', forward);
    process.removeListener('SIGTERM', forward);
    process.exitCode = signal ? 128 + (constants.signals[signal] || 1) : (code ?? 1);
  });
}
