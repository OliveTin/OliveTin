#!/usr/bin/env python3
"""Prepare the logs list screenshot after running a dropdown action."""

import time

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait


def _wait_for_dashboard(driver, timeout=15):
    WebDriverWait(driver, timeout).until(
        lambda d: bool(d.find_element(By.TAG_NAME, "body").get_attribute("loaded-dashboard"))
    )
    WebDriverWait(driver, timeout).until(
        lambda d: d.execute_script("return !!window.client")
    )


def _wait_for_logs_table(driver, timeout=15):
    WebDriverWait(driver, timeout).until(
        lambda d: len(d.find_elements(By.CSS_SELECTOR, ".logs-table tbody tr")) >= 1
    )


def run(driver):
    _wait_for_dashboard(driver)

    driver.execute_async_script(
        """
        const done = arguments[arguments.length - 1];
        const button = document.querySelector('[title="Print a message"]');
        const bindingId = button.closest('.action-button').id.replace('actionButton-', '');

        window.client.startAction({
          bindingId: bindingId,
          arguments: [{ name: 'message', value: 'Hello there!' }],
          uniqueTrackingId: 'doc-screenshot-' + Date.now(),
        }).then(() => done(true)).catch((err) => done(String(err)));
        """
    )

    time.sleep(0.5)
    driver.execute_script("window.location.href = '/logs'")
    _wait_for_logs_table(driver)
    time.sleep(0.2)
