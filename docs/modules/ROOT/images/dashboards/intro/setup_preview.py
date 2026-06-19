#!/usr/bin/env python3
"""Open the My Servers dashboard from dashboards/intro.adoc."""

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


def _wait_for_my_servers_dashboard(driver, timeout=30):
    def ready(d):
        try:
            ping_all = d.find_element(By.CSS_SELECTOR, '[title="Ping All Servers"]')
            hypervisors = d.find_element(
                By.XPATH,
                '//button[contains(@class, "directory-button")]//span[contains(@class, "title") and text()="Hypervisors"]',
            )
            server1 = d.find_element(By.CSS_SELECTOR, '[title="server1 Wake on Lan"]')
            server3 = d.find_element(By.CSS_SELECTOR, '[title="server3 Power Off"]')
        except Exception:
            return False
        return all(
            element.is_displayed()
            for element in (ping_all, hypervisors, server1, server3)
        )

    WebDriverWait(driver, timeout).until(ready)


def run(driver):
    _wait_for_dashboard(driver)
    _wait_for_my_servers_dashboard(driver)
    time.sleep(0.2)
