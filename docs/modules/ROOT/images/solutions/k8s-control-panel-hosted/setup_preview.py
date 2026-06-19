#!/usr/bin/env python3
"""Show the default Actions dashboard for the Kubernetes control panel."""

import time

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait

ACTION_TITLES = [
    "get pods",
    "restart postgres deployment",
    "evacuate node",
]


def _wait_for_dashboard(driver, timeout=30):
    WebDriverWait(driver, timeout).until(
        lambda d: d.execute_script("return !!window.client")
    )
    WebDriverWait(driver, timeout).until(
        lambda d: bool(d.find_element(By.TAG_NAME, "body").get_attribute("loaded-dashboard"))
    )


def _wait_for_actions(driver, timeout=30):
    def ready(d):
        for title in ACTION_TITLES:
            try:
                button = d.find_element(By.CSS_SELECTOR, f'[title="{title}"]')
            except Exception:
                return False
            if not button.is_displayed():
                return False
        return len(d.find_elements(By.CSS_SELECTOR, ".action-button button")) >= 3

    WebDriverWait(driver, timeout).until(ready)


def run(driver):
    _wait_for_dashboard(driver)
    _wait_for_actions(driver)
    time.sleep(0.2)
