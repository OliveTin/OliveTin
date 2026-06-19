#!/usr/bin/env python3
"""Show a timed-out action on the logs page."""

import time

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait


def _wait_for_body_attr(driver, attr, timeout=15):
    WebDriverWait(driver, timeout).until(
        lambda d: bool(d.find_element(By.TAG_NAME, "body").get_attribute(attr))
    )


def _wait_for_timed_out_log(driver, timeout=20):
    def has_timed_out_row(d):
        try:
            status = d.find_element(By.CSS_SELECTOR, ".logs-table .status-timeout").text
        except Exception:
            return False
        return "Timed out" in status

    WebDriverWait(driver, timeout).until(has_timed_out_row)


def run(driver):
    _wait_for_body_attr(driver, "loaded-dashboard")

    driver.find_element(By.CSS_SELECTOR, '[title="Slow action"]').click()

    # Default timeout is 3 seconds; the action sleeps for 5.
    time.sleep(5)

    driver.execute_script("window.location.href = '/logs'")

    WebDriverWait(driver, 15).until(
        lambda d: d.find_elements(By.CSS_SELECTOR, ".logs-table tbody tr")
    )
    _wait_for_timed_out_log(driver)

    time.sleep(0.2)
