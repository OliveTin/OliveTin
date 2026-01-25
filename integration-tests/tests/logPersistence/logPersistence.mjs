import { describe, it, before, after } from 'mocha'
import { expect } from 'chai'
import { By, Condition } from 'selenium-webdriver'
import fs from 'fs'
import path from 'path'
import {
  getRootAndWait,
  getActionButtons,
  takeScreenshotOnFailure,
} from '../../lib/elements.js'

describe('config: logPersistence', function () {
  const logsDir = '/tmp/olivetin-test-logs'
  let firstExecutionTrackingId = null

  before(async function () {
    // Clean up any existing test logs
    if (fs.existsSync(logsDir)) {
      fs.rmSync(logsDir, { recursive: true, force: true })
    }
    fs.mkdirSync(logsDir, { recursive: true })

    await runner.start('logPersistence')
  })

  after(async () => {
    await runner.stop()

    // Clean up test logs directory
    if (fs.existsSync(logsDir)) {
      fs.rmSync(logsDir, { recursive: true, force: true })
    }
  })

  afterEach(function () {
    takeScreenshotOnFailure(this.currentTest, webdriver)
  })

  it('Execute action and verify log is saved to disk', async function () {
    this.timeout(30000)
    await getRootAndWait()

    // Get initial log file count
    const initialLogCount = fs.existsSync(logsDir)
      ? fs.readdirSync(logsDir).filter(f => f.endsWith('.yaml')).length
      : 0

    // Wait for action button to be available
    await webdriver.wait(
      new Condition('wait for Echo Test button', async () => {
        const buttons = await webdriver.findElements(By.css('.action-button button'))
        for (const btn of buttons) {
          const text = await btn.getText()
          if (text.includes('Echo Test')) {
            return true
          }
        }
        return false
      }),
      10000
    )

    // Find and click the Echo Test button
    const buttons = await webdriver.findElements(By.css('.action-button button'))
    let echoButton = null
    for (const btn of buttons) {
      const text = await btn.getText()
      if (text.includes('Echo Test')) {
        echoButton = btn
        break
      }
    }
    expect(echoButton).to.not.be.null

    // Click the button to execute the action
    await echoButton.click()

    // Wait for the log file to be written to disk
    await webdriver.wait(
      new Condition('wait for log file to appear', async () => {
        if (!fs.existsSync(logsDir)) {
          return false
        }
        const logFiles = fs.readdirSync(logsDir).filter(f => f.endsWith('.yaml'))
        return logFiles.length > initialLogCount
      }),
      10000
    )

    // Wait a bit more to ensure file is fully written
    await webdriver.sleep(1000)

    // Get the newest log file
    const logFiles = fs.readdirSync(logsDir).filter(f => f.endsWith('.yaml'))
    expect(logFiles.length).to.be.greaterThan(initialLogCount, 'At least one new log file should be saved')

    // Sort by modification time to get the newest
    const logFilesWithStats = logFiles.map(f => {
      const filePath = path.join(logsDir, f)
      return {
        name: f,
        path: filePath,
        mtime: fs.statSync(filePath).mtime
      }
    }).sort((a, b) => b.mtime - a.mtime)

    const newestLogFile = logFilesWithStats[0]
    expect(newestLogFile).to.not.be.undefined

    // Read the log file to extract the tracking ID
    const logFileContent = fs.readFileSync(newestLogFile.path, 'utf8')

    // Verify the log file contains expected content (action title might be in different fields)
    expect(logFileContent.length).to.be.greaterThan(0, 'Log file should not be empty')

    // Extract tracking ID from filename first (most reliable)
    // Filename format: <title>.<timestamp>.<trackingId>.yaml
    // Tracking IDs are UUIDs, so match UUID pattern at the end before .yaml
    let uuidMatch = newestLogFile.name.match(/([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})\.yaml$/)

    if (uuidMatch) {
      firstExecutionTrackingId = uuidMatch[1]
    } else {
      // Fallback: split by dots and get the last part before .yaml
      const parts = newestLogFile.name.replace('.yaml', '').split('.')
      if (parts.length >= 3) {
        // The last part should be the tracking ID
        firstExecutionTrackingId = parts[parts.length - 1]
      }
    }

    // If still not found, try to extract from YAML content
    // Try different possible field name variations
    if (!firstExecutionTrackingId) {
      const patterns = [
        /executionTrackingID:\s*([^\s\n]+)/i,
        /execution_tracking_id:\s*([^\s\n]+)/i,
        /ExecutionTrackingID:\s*([^\s\n]+)/,
        /executionTrackingId:\s*([^\s\n]+)/i,
      ]

      for (const pattern of patterns) {
        const match = logFileContent.match(pattern)
        if (match) {
          firstExecutionTrackingId = match[1].trim()
          break
        }
      }
    }

    expect(firstExecutionTrackingId).to.not.be.null
    expect(firstExecutionTrackingId.length).to.be.greaterThan(0)

    // Verify the log file name contains the tracking ID
    expect(newestLogFile.name).to.include(firstExecutionTrackingId)

    // Verify the log file content contains the action (might be in actionTitle, actionConfigTitle, or title field)
    const hasActionReference = logFileContent.includes('Echo Test') ||
                              logFileContent.includes('echo') ||
                              logFileContent.includes('actionTitle') ||
                              logFileContent.includes('actionConfigTitle')
    expect(hasActionReference).to.be.true
  })

  it('Restart service and verify logs are loaded from disk', async function () {
    this.timeout(60000)

    // Skip if first test didn't set the tracking ID
    if (!firstExecutionTrackingId) {
      this.skip()
    }

    // Verify log file exists before restart
    const logFilesBeforeRestart = fs.readdirSync(logsDir).filter(f => f.endsWith('.yaml'))
    expect(logFilesBeforeRestart.length).to.be.greaterThan(0, 'Log file should exist before restart')

    // Find the log file for this execution
    const matchingLogFileBefore = logFilesBeforeRestart.find(f => f.includes(firstExecutionTrackingId))
    expect(matchingLogFileBefore).to.not.be.undefined

    // Stop the current service instance
    await runner.stop()

    // Wait a moment to ensure the process has fully stopped
    await new Promise((resolve) => setTimeout(resolve, 2000))

    // Verify log file still exists after stop (should not be deleted)
    const logFilesAfterStop = fs.readdirSync(logsDir).filter(f => f.endsWith('.yaml'))
    expect(logFilesAfterStop.length).to.be.greaterThan(0, 'Log file should still exist after service stop')

    const matchingLogFileAfter = logFilesAfterStop.find(f => f.includes(firstExecutionTrackingId))
    expect(matchingLogFileAfter).to.not.be.undefined

    // Start a new service instance (logs should be loaded from disk)
    await runner.start('logPersistence')

    // Wait for the service to fully start and load logs
    await new Promise((resolve) => setTimeout(resolve, 3000))

    await getRootAndWait()

    // Navigate directly to the specific log entry (this verifies the log was loaded)
    await webdriver.get(runner.baseUrl() + 'logs/' + firstExecutionTrackingId)

    // Wait for the log details page to load
    await webdriver.wait(
      new Condition('wait for log details to load', async () => {
        try {
          const body = await webdriver.findElement(By.tagName('body'))
          const text = await body.getText()
          // The log should contain the output from the echo command
          return text.includes('Hello from persisted log test') || text.includes(firstExecutionTrackingId)
        } catch (e) {
          return false
        }
      }),
      15000
    )

    // Verify the log content is displayed
    const body = await webdriver.findElement(By.tagName('body'))
    const bodyText = await body.getText()

    // The persisted log should be accessible and contain the expected output
    expect(bodyText).to.include('Hello from persisted log test')
  })

  it('Verify log file still exists after restart', async function () {
    // Skip if first test didn't set the tracking ID
    if (!firstExecutionTrackingId) {
      this.skip()
    }

    // Verify the log file still exists on disk
    const logFiles = fs.readdirSync(logsDir).filter(f => f.endsWith('.yaml'))
    expect(logFiles.length).to.be.greaterThan(0, 'Log files should still exist after restart')

    // Find the log file for the first execution
    const matchingLogFile = logFiles.find(f => f.includes(firstExecutionTrackingId))
    expect(matchingLogFile).to.not.be.undefined
    expect(matchingLogFile).to.not.be.null

    // Verify the log file content is still valid
    const logFilePath = path.join(logsDir, matchingLogFile)
    const logFileContent = fs.readFileSync(logFilePath, 'utf8')
    expect(logFileContent.length).to.be.greaterThan(0, 'Log file should not be empty')

    // The filename contains the tracking ID, so verify that
    expect(matchingLogFile).to.include(firstExecutionTrackingId)

    // Verify the file contains some expected content (action reference)
    const hasActionReference = logFileContent.includes('Echo Test') ||
                              logFileContent.includes('echo') ||
                              logFileContent.includes('actionTitle') ||
                              logFileContent.includes('actionConfigTitle')
    expect(hasActionReference).to.be.true
  })
})
