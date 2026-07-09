#!/usr/bin/env python3
"""Prepare the multiline textarea argument screenshot."""

import time

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait


def _wait_for_body_attr(driver, attr, timeout=15):
    WebDriverWait(driver, timeout).until(
        lambda d: bool(d.find_element(By.TAG_NAME, "body").get_attribute(attr))
    )


def run(driver):
    _wait_for_body_attr(driver, "loaded-dashboard")

    driver.find_element(By.CSS_SELECTOR, '[title="Save text to file"]').click()

    WebDriverWait(driver, 15).until(
        lambda d: len(
            d.find_elements(By.CSS_SELECTOR, "#argument-popup textarea#content")
        )
        >= 1
    )
    time.sleep(0.2)
