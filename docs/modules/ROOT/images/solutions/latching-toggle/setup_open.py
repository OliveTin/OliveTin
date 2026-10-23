#!/usr/bin/env python3
"""Click Open valve and wait for the dashboard to show the valve open."""

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
    driver.find_element(By.CSS_SELECTOR, 'button[title="Open valve"]').click()
    wait_for_state(driver, "open")
    WebDriverWait(driver, 10).until(
        lambda d: d.find_element(By.CSS_SELECTOR, 'button[title="Close valve"]').is_enabled()
    )
    time.sleep(0.5)
