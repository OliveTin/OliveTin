#!/usr/bin/env python3
"""Prepare the execution results screenshot for a dropdown action."""

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


def _wait_for_logs_page(driver, timeout=15):
    WebDriverWait(driver, timeout).until(
        lambda d: "/logs/" in d.current_url and not d.current_url.rstrip("/").endswith("/logs")
    )


def _wait_for_execution_complete(driver, timeout=15):
    def finished(d):
        try:
            status = d.find_element(By.CSS_SELECTOR, ".execution-dialog-status").text
        except Exception:
            return False
        return "Still running" not in status and "Queued" not in status

    WebDriverWait(driver, timeout).until(finished)


def _wait_for_terminal_output(driver, expected, timeout=15):
    WebDriverWait(driver, timeout).until(
        lambda d: expected
        in d.execute_script(
            """
            if (!window.terminal || !window.terminal.getBufferAsString) {
              return '';
            }
            return window.terminal.getBufferAsString();
            """
        )
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
        }).then((response) => {
          window.location.href = '/logs/' + response.executionTrackingId;
          done(true);
        }).catch((err) => done(String(err)));
        """
    )

    _wait_for_logs_page(driver)
    _wait_for_execution_complete(driver)
    _wait_for_terminal_output(driver, "Hello there!")
    time.sleep(0.2)
