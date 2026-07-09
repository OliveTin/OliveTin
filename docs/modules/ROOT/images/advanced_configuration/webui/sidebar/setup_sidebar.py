#!/usr/bin/env python3
"""Capture the default sidebar navigation style."""

import time

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait


def _wait_for_sidebar_navigation(driver, timeout=30):
    WebDriverWait(driver, timeout).until(
        lambda d: bool(d.find_element(By.TAG_NAME, "body").get_attribute("loaded-dashboard"))
    )

    driver.find_element(By.ID, "sidebar-toggler-button").click()

    WebDriverWait(driver, timeout).until(
        lambda d: d.find_element(By.ID, "mainnav").is_displayed()
    )

    driver.find_element(By.CSS_SELECTOR, "#mainnav .stick-toggle").click()

    def ready(d):
        try:
            my_servers = d.find_element(
                By.XPATH,
                '//*[@id="mainnav"]//a[contains(normalize-space(), "My Servers")]',
            )
            my_containers = d.find_element(
                By.XPATH,
                '//*[@id="mainnav"]//a[contains(normalize-space(), "My Containers")]',
            )
            disk_space = d.find_element(By.CSS_SELECTOR, '[title="Check disk space"]')
            ssh = d.find_element(By.CSS_SELECTOR, '[title="Setup easy SSH"]')
        except Exception:
            return False
        return all(element.is_displayed() for element in (my_servers, my_containers, disk_space, ssh))

    WebDriverWait(driver, timeout).until(ready)


def run(driver):
    _wait_for_sidebar_navigation(driver)
    time.sleep(0.2)
