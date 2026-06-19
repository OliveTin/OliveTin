#!/usr/bin/env python3
"""Open the My Services dashboard for the systemd control panel solution."""

import time

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait


def _wait_for_dashboard(driver, timeout=30):
    WebDriverWait(driver, timeout).until(
        lambda d: d.execute_script("return !!window.client")
    )
    WebDriverWait(driver, timeout).until(
        lambda d: bool(d.find_element(By.TAG_NAME, "body").get_attribute("loaded-dashboard"))
    )


def _wait_for_my_services_dashboard(driver, timeout=30):
    def ready(d):
        required_titles = [
            "Start boot.mount",
            "Stop boot.mount",
            "Start podman.service",
            "Start upsilon-drone.service",
        ]
        for title in required_titles:
            try:
                button = d.find_element(By.CSS_SELECTOR, f'[title="{title}"]')
            except Exception:
                return False
            if not button.is_displayed():
                return False
        return True

    WebDriverWait(driver, timeout).until(ready)


def run(driver):
    _wait_for_dashboard(driver)
    _wait_for_my_services_dashboard(driver)
    time.sleep(0.2)
