#!/usr/bin/env python3
"""Open the Water Valve dashboard with the valve closed."""

import time

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait


def wait_for_dashboard(driver, timeout=30):
    WebDriverWait(driver, timeout).until(
        lambda d: bool(d.find_element(By.TAG_NAME, "body").get_attribute("loaded-dashboard"))
    )


def wait_for_state(driver, state, timeout=30):
    def ready(d):
        try:
            return f"State: {state}" in d.find_element(By.TAG_NAME, "main").text
        except Exception:
            return False

    WebDriverWait(driver, timeout).until(ready)


def run(driver):
    wait_for_dashboard(driver)
    wait_for_state(driver, "closed")
    time.sleep(0.2)
