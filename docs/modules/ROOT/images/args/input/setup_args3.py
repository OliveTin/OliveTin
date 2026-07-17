#!/usr/bin/env python3
"""Prepare the execution-results screenshot."""

import time

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait


def _wait_for_body_attr(driver, attr, timeout=15):
    WebDriverWait(driver, timeout).until(
        lambda d: bool(d.find_element(By.TAG_NAME, "body").get_attribute(attr))
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


def _start_action_and_open_logs(driver, timeout=15):
    WebDriverWait(driver, timeout).until(
        lambda d: d.execute_script("return !!window.client")
    )
    WebDriverWait(driver, timeout).until(
        lambda d: d.find_element(By.CSS_SELECTOR, 'button[name="start"]').is_enabled()
    )

    driver.execute_async_script(
        """
        const done = arguments[arguments.length - 1];
        const bindingId = document.body.getAttribute('loaded-argument-form');
        window.client.startAction({
          bindingId: bindingId,
          arguments: [{ name: 'message', value: 'Hello World' }],
          uniqueTrackingId: 'doc-screenshot-' + Date.now(),
        }).then((response) => {
          window.location.href = '/logs/' + response.executionTrackingId;
          done(true);
        }).catch((err) => done('error: ' + err));
        """
    )


def run(driver):
    _wait_for_body_attr(driver, "loaded-dashboard")

    action = driver.find_element(By.CSS_SELECTOR, '[title="Print a message"]')
    action.click()

    _wait_for_body_attr(driver, "loaded-argument-form")
    _start_action_and_open_logs(driver)

    _wait_for_logs_page(driver)
    _wait_for_execution_complete(driver)
    WebDriverWait(driver, 15).until(
        lambda d: "Hello World"
        in d.execute_script(
            """
            if (!window.terminal || !window.terminal.getBufferAsString) {
              return '';
            }
            return window.terminal.getBufferAsString();
            """
        )
    )
    time.sleep(0.2)
