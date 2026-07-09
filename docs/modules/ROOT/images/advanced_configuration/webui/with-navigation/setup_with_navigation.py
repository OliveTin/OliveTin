#!/usr/bin/env python3
"""Capture the default Actions view with sidebar navigation visible."""

import time

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait


def run(driver):
    WebDriverWait(driver, 15).until(
        lambda d: bool(d.find_element(By.TAG_NAME, "body").get_attribute("loaded-dashboard"))
    )
    WebDriverWait(driver, 15).until(
        lambda d: len(d.find_elements(By.CSS_SELECTOR, ".action-button button")) >= 8
    )

    driver.find_element(By.ID, "sidebar-toggler-button").click()

    WebDriverWait(driver, 15).until(
        lambda d: d.find_element(By.ID, "mainnav").is_displayed()
    )
    time.sleep(0.2)
