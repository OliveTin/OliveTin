#!/usr/bin/env python3
"""Show action buttons in idle, running, and queued states."""

import time

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait

_START_ACTION_JS = """
const done = arguments[arguments.length - 1];
const title = arguments[0];

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

window.client.startAction({
  bindingId: bindingIdForTitle(title),
  arguments: [],
  uniqueTrackingId: uniqueTrackingId(),
}).then(() => done(true)).catch((err) => done(String(err)));
"""


def _wait_for_dashboard(driver, timeout=30):
    WebDriverWait(driver, timeout).until(
        lambda d: d.execute_script("return !!window.client")
    )
    WebDriverWait(driver, timeout).until(
        lambda d: bool(d.find_element(By.TAG_NAME, "body").get_attribute("loaded-dashboard"))
    )


def _start_action(driver, title):
    driver.execute_async_script(_START_ACTION_JS, title)


def _wait_for_layout_states(driver, timeout=20):
    def ready(d):
        try:
            restart = d.find_element(By.CSS_SELECTOR, '[title="Restart service"]')
            running = d.find_element(
                By.CSS_SELECTOR,
                '[title="Long task"]'
            ).find_element(By.XPATH, './ancestor::div[contains(@class, "action-button")]//span[contains(@class, "execution-indicator-running")]')
            queued = d.find_element(
                By.CSS_SELECTOR,
                '[title="Backup job"]'
            ).find_element(By.XPATH, './ancestor::div[contains(@class, "action-button")]//span[contains(@class, "execution-indicator-queued")]')
            onclick = d.find_element(
                By.CSS_SELECTOR,
                '[title="Restart service"] .navigate-on-start',
            )
        except Exception:
            return False
        return all(
            element.is_displayed()
            for element in (restart, running, queued, onclick)
        )

    WebDriverWait(driver, timeout).until(ready)


def run(driver):
    _wait_for_dashboard(driver)

    _start_action(driver, "Long task")
    time.sleep(0.3)
    _start_action(driver, "Backup job")

    _wait_for_layout_states(driver)
    time.sleep(0.2)
