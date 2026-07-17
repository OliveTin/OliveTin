#!/usr/bin/env python3
"""Open My First Dashboard from dashboards/2-fieldsets.adoc."""

import time

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait


def _wait_for_fieldset_dashboard(driver, timeout=30):
    WebDriverWait(driver, timeout).until(
        lambda d: d.find_element(By.TAG_NAME, "body").get_attribute("loaded-dashboard")
        == "My First Dashboard"
    )

    def ready(d):
        try:
            folder1 = d.find_element(
                By.XPATH,
                '//button[contains(@class, "directory-button")]//span[contains(@class, "title") and text()="Folder 1"]',
            )
            folder2 = d.find_element(
                By.XPATH,
                '//button[contains(@class, "directory-button")]//span[contains(@class, "title") and text()="Folder 2"]',
            )
        except Exception:
            return False
        return all(element.is_displayed() for element in (folder1, folder2))

    WebDriverWait(driver, timeout).until(ready)


def run(driver):
    _wait_for_fieldset_dashboard(driver)
    time.sleep(0.2)
