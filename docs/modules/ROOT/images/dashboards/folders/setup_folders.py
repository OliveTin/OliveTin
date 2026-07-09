#!/usr/bin/env python3
"""Open Folder 1 from dashboards/3-folders.adoc."""

import time

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait


def _wait_for_folder_dashboard(driver, timeout=30):
    WebDriverWait(driver, timeout).until(
        lambda d: d.find_element(By.TAG_NAME, "body").get_attribute("loaded-dashboard")
        == "Folder 1"
    )

    def ready(d):
        try:
            action1 = d.find_element(By.CSS_SELECTOR, '[title="Action 1"]')
            action2 = d.find_element(By.CSS_SELECTOR, '[title="Action 2"]')
            subfolder = d.find_element(
                By.XPATH,
                '//button[contains(@class, "directory-button")]//span[contains(@class, "title") and text()="Subfolder 2"]',
            )
        except Exception:
            return False
        return all(
            element.is_displayed()
            for element in (action1, action2, subfolder)
        )

    WebDriverWait(driver, timeout).until(ready)


def run(driver):
    _wait_for_folder_dashboard(driver)
    time.sleep(0.2)
