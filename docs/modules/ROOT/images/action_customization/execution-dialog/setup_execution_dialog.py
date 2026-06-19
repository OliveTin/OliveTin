#!/usr/bin/env python3
"""Open the execution-dialog view for Check dmesg logs."""

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


def run(driver):
    _wait_for_body_attr(driver, "loaded-dashboard")

    driver.find_element(By.CSS_SELECTOR, '[title="Check dmesg logs"]').click()

    _wait_for_logs_page(driver)
    _wait_for_execution_complete(driver)

    WebDriverWait(driver, 15).until(
        lambda d: d.find_element(By.CSS_SELECTOR, "#execution-results-popup .xterm-rows").text.strip() != ""
    )

    time.sleep(0.2)
