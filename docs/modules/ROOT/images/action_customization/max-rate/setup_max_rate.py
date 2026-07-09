#!/usr/bin/env python3
"""Show blocked rate-limit log entries for the date action."""

import time

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait

_START_ACTION_REPEATEDLY_JS = """
const done = arguments[arguments.length - 1];
const title = arguments[0];
const count = arguments[1];

function bindingIdForTitle(actionTitle) {
  const button = document.querySelector('[title="' + actionTitle + '"]');
  if (!button) {
    throw new Error('Action button not found: ' + actionTitle);
  }
  return button.closest('.action-button').id.replace('actionButton-', '');
}

function uniqueTrackingId() {
  if (window.isSecureContext && window.crypto?.randomUUID) {
    return window.crypto.randomUUID();
  }
  return 'doc-screenshot-' + Date.now() + '-' + Math.random();
}

async function startAndWait(actionTitle) {
  const bindingId = bindingIdForTitle(actionTitle);
  const trackingId = uniqueTrackingId();
  const response = await window.client.startAction({
    bindingId,
    arguments: [],
    uniqueTrackingId: trackingId,
  });
  const executionTrackingId = response.executionTrackingId || trackingId;

  while (true) {
    const result = await window.client.executionStatus({
      executionTrackingId,
    });
    if (result.logEntry?.executionFinished) {
      return result.logEntry;
    }
    await new Promise((resolve) => setTimeout(resolve, 100));
  }
}

async function startRepeatedly(actionTitle, runs) {
  for (let i = 0; i < runs; i++) {
    await startAndWait(actionTitle);
  }
}

startRepeatedly(title, count).then(() => done(true)).catch((err) => done(String(err)));
"""


def _wait_for_dashboard(driver, timeout=15):
    WebDriverWait(driver, timeout).until(
        lambda d: bool(d.find_element(By.TAG_NAME, "body").get_attribute("loaded-dashboard"))
    )
    WebDriverWait(driver, timeout).until(
        lambda d: d.execute_script("return !!window.client")
    )


def _wait_for_rate_limited_logs(driver, timeout=20):
    def ready(d):
        try:
            rows = d.find_elements(By.CSS_SELECTOR, ".logs-table tbody tr")
            blocked = d.find_elements(By.CSS_SELECTOR, ".logs-table .status-blocked")
            completed = d.find_elements(By.CSS_SELECTOR, ".logs-table .status-success")
        except Exception:
            return False
        return len(rows) >= 5 and len(blocked) >= 2 and len(completed) >= 3

    WebDriverWait(driver, timeout).until(ready)


def run(driver):
    _wait_for_dashboard(driver)

    driver.execute_async_script(_START_ACTION_REPEATEDLY_JS, "date", 5)

    time.sleep(1)

    driver.execute_script("window.location.href = '/logs'")
    _wait_for_rate_limited_logs(driver)

    time.sleep(0.2)
