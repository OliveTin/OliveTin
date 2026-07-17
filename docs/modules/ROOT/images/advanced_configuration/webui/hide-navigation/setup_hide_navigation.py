#!/usr/bin/env python3
"""Capture the Actions view with navigation hidden."""

import time

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait


def run(driver):
    WebDriverWait(driver, 15).until(
        lambda d: bool(d.find_element(By.TAG_NAME, "body").get_attribute("loaded-dashboard"))
    )
    WebDriverWait(driver, 15).until(
        lambda d: len(d.find_elements(By.ID, "mainnav")) == 0
    )
    WebDriverWait(driver, 15).until(
        lambda d: len(d.find_elements(By.CSS_SELECTOR, ".action-button button")) >= 8
    )
    time.sleep(0.2)
