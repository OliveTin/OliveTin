#!/usr/bin/env python3
"""Capture the custom page title from advanced_configuration/webui.adoc."""

import time

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait


def run(driver):
    WebDriverWait(driver, 15).until(
        lambda d: bool(d.find_element(By.TAG_NAME, "body").get_attribute("loaded-dashboard"))
    )
    WebDriverWait(driver, 15).until(
        lambda d: "My OliveTin Instance" in d.find_element(By.CSS_SELECTOR, "header h1").text
    )
    WebDriverWait(driver, 15).until(
        lambda d: d.find_element(By.CSS_SELECTOR, '[title="Ping the Internet"]').is_displayed()
    )
    time.sleep(0.2)
