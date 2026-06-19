#!/usr/bin/env python3
"""Open the Human in the Control Loop dashboard with water level output."""

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


def _wait_for_water_level(driver, timeout=30):
    def ready(d):
        try:
            output = d.find_element(By.CSS_SELECTOR, ".mre-output").text
            pump = d.find_element(By.CSS_SELECTOR, '[title="Pump ON - 5m"]')
        except Exception:
            return False
        return "Water level 47%" in output and pump.is_displayed()

    WebDriverWait(driver, timeout).until(ready)


def run(driver):
    _wait_for_dashboard(driver)
    _wait_for_water_level(driver)
    time.sleep(0.2)
