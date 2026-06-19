#!/usr/bin/env python3
"""Open the logs calendar view with executions on the current month."""

import time

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait

_START_ACTIONS_JS = """
const done = arguments[arguments.length - 1];
const titles = arguments[0];

function bindingIdForTitle(title) {
  const button = document.querySelector('[title="' + title + '"]');
  if (!button) {
    throw new Error('Action button not found: ' + title);
  }
  return button.closest('.action-button').id.replace('actionButton-', '');
}

function uniqueTrackingId() {
  if (window.isSecureContext && window.crypto?.randomUUID) {
    return window.crypto.randomUUID();
  }
  return 'doc-screenshot-' + Date.now() + '-' + Math.random();
}

function startByTitle(title) {
  return window.client.startAction({
    bindingId: bindingIdForTitle(title),
    arguments: [],
    uniqueTrackingId: uniqueTrackingId(),
  });
}

Promise.all(titles.map(startByTitle)).then(() => done(true)).catch((err) => done(String(err)));
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

    driver.execute_async_script(
        _START_ACTIONS_JS,
        ["Check disk space", "Restart service"],
    )

    time.sleep(2)

    driver.execute_script("window.location.href = '/logs/calendar'")

    WebDriverWait(driver, 15).until(
        lambda d: len(d.find_elements(By.CSS_SELECTOR, ".calendar-event")) >= 2
    )

    driver.execute_script(
        """
        const today = new Date();
        const key = today.getFullYear() + '-'
          + String(today.getMonth() + 1).padStart(2, '0') + '-'
          + String(today.getDate()).padStart(2, '0');
        const cell = document.querySelector('[data-calendar-date="' + key + '"]');
        if (cell) {
          cell.scrollIntoView({ block: 'center' });
        }
        """
    )

    time.sleep(0.2)
