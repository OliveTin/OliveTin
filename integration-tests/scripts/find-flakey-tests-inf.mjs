#!/usr/bin/env node
import { spawn } from 'node:child_process'
import {
  appendFileSync,
  writeFileSync,
  readFileSync,
  unlinkSync,
} from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { tmpdir } from 'node:os'
import { randomUUID } from 'node:crypto'

const rootDir = join(dirname(fileURLToPath(import.meta.url)), '..')
const logFile = process.env.FLAKEY_LOG_FILE || join(rootDir, 'flakey-test-runs.log')
const jsonlFile = process.env.FLAKEY_JSONL_FILE || join(rootDir, 'flakey-test-runs.jsonl')

function formatFailure (failure) {
  const err = failure.err || {}
  const lines = [
    `FAILURE: ${failure.fullTitle || failure.title}`,
    `  file: ${failure.file || 'unknown'}`,
    `  message: ${(err.message || 'unknown').trim()}`,
  ]

  if (err.stack) {
    lines.push('  stack:')
    for (const line of err.stack.split('\n').slice(0, 8)) {
      lines.push(`    ${line}`)
    }
  }

  return lines.join('\n')
}

function appendRunLog (run, exitCode, report, durationMs, spawnError) {
  const timestamp = new Date().toISOString()
  const stats = report?.stats || {}
  const passes = stats.passes ?? '?'
  const failures = stats.failures ?? '?'
  const pending = stats.pending ?? 0
  const passed = exitCode === 0 && !spawnError
  const result = passed ? 'PASS' : 'FAIL'
  const durationSec = (durationMs / 1000).toFixed(1)

  const block = [
    `=== RUN ${run} | ${timestamp} | ${result} | ${passes} pass ${failures} fail ${pending} pending | ${durationSec}s ===`,
  ]

  if (spawnError) {
    block.push(`SPAWN_ERROR: ${spawnError}`)
  }

  if (report?.failures?.length) {
    for (const failure of report.failures) {
      block.push(formatFailure(failure))
    }
  } else if (!passed && !report) {
    block.push('No JSON report captured (mocha may have crashed before writing results)')
  }

  block.push('')
  appendFileSync(logFile, `${block.join('\n')}\n`)

  const jsonl = {
    run,
    timestamp,
    exitCode,
    durationMs,
    passes,
    failures,
    pending,
    failureDetails: (report?.failures || []).map((failure) => ({
      fullTitle: failure.fullTitle || failure.title,
      file: failure.file,
      message: failure.err?.message,
      stack: failure.err?.stack,
    })),
  }
  appendFileSync(jsonlFile, `${JSON.stringify(jsonl)}\n`)
}

function runMochaOnce () {
  const reportPath = join(tmpdir(), `mocha-flakey-${randomUUID()}.json`)

  return new Promise((resolve) => {
    const proc = spawn('npx', [
      'mocha',
      'tests',
      '--recursive',
      '-t',
      '10000',
      '--reporter',
      'json',
      '--reporter-option',
      `output=${reportPath}`,
    ], {
      cwd: rootDir,
      stdio: ['ignore', 'inherit', 'inherit'],
    })

    proc.on('close', (exitCode) => {
      let report = null
      try {
        report = JSON.parse(readFileSync(reportPath, 'utf8'))
      } catch {
        report = null
      }

      try {
        unlinkSync(reportPath)
      } catch {
        // ignore missing temp report
      }

      resolve({ exitCode: exitCode ?? 1, report })
    })

    proc.on('error', (spawnError) => {
      resolve({ exitCode: 1, report: null, spawnError: spawnError.message })
    })
  })
}

async function main () {
  const header = [
    `# Flaky test run log started ${new Date().toISOString()}`,
    `# Log file: ${logFile}`,
    `# JSONL file: ${jsonlFile}`,
    '',
  ].join('\n')

  writeFileSync(logFile, `${header}\n`)
  writeFileSync(jsonlFile, '')

  console.log(`Logging flaky test runs to ${logFile}`)
  console.log(`Structured run data: ${jsonlFile}`)

  let run = 0

  while (true) {
    run += 1
    console.log(`\n--- Starting run ${run} ---`)

    const start = Date.now()
    const { exitCode, report, spawnError } = await runMochaOnce()
    const durationMs = Date.now() - start

    appendRunLog(run, exitCode, report, durationMs, spawnError)

    const summary = exitCode === 0 ? 'PASS' : 'FAIL'
    console.log(`Run ${run}: ${summary} (${(durationMs / 1000).toFixed(1)}s) — logged`)

    if (exitCode !== 0) {
      console.log(`Failure on run ${run}, stopping. See ${logFile}`)
      process.exit(exitCode)
    }
  }
}

main().catch((err) => {
  console.error(err)
  process.exit(1)
})
