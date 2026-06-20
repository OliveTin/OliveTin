#!/usr/bin/env python3
"""Show queued executions on the logs queue page."""

import time

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait

_QUEUE_ACTIONS_JS = """
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

async function queueActions() {
  const bindingId = bindingIdForTitle(title);
  for (let i = 0; i < count; i++) {
    await window.client.startAction({
      bindingId: bindingId,
      arguments: [],
      uniqueTrackingId: uniqueTrackingId(),
    });
  }
}

queueActions().then(() => done(true)).catch((err) => done(String(err)));
"""


def _wait_for_dashboard(driver, timeout=15):
    WebDriverWait(driver, timeout).until(
        lambda d: bool(d.find_element(By.TAG_NAME, "body").get_attribute("loaded-dashboard"))
    )
    WebDriverWait(driver, timeout).until(
        lambda d: d.execute_script("return !!window.client")
    )


def run(driver):
    _wait_for_dashboard(driver)

    driver.execute_async_script(_QUEUE_ACTIONS_JS, "Slow backup", 3)

    time.sleep(1)

    driver.execute_script("window.location.href = '/logs/queue'")

    WebDriverWait(driver, 15).until(
        lambda d: len(d.find_elements(By.CSS_SELECTOR, ".queue-action-group-section")) >= 1
    )
    WebDriverWait(driver, 15).until(
        lambda d: len(d.find_elements(By.CSS_SELECTOR, ".queue-position")) >= 1
    )

    time.sleep(0.2)
