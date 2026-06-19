#!/usr/bin/env python3
"""Show the Say Hello action on the default dashboard."""

import time

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait


def run(driver):
    WebDriverWait(driver, 30).until(
        lambda d: d.execute_script("return !!window.client")
    )
    WebDriverWait(driver, 30).until(
        lambda d: bool(d.find_element(By.TAG_NAME, "body").get_attribute("loaded-dashboard"))
    )
    WebDriverWait(driver, 30).until(
        lambda d: d.find_element(By.CSS_SELECTOR, '[title="Say Hello"]').is_displayed()
    )
    time.sleep(0.2)
