#!/usr/bin/env python3
"""Capture the topbar section navigation style."""

import time

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait


def _wait_for_topbar_navigation(driver, timeout=30):
    WebDriverWait(driver, timeout).until(
        lambda d: bool(d.find_element(By.TAG_NAME, "body").get_attribute("loaded-dashboard"))
    )

    def ready(d):
        try:
            topbar = d.find_element(By.CSS_SELECTOR, "nav.topbar")
            my_servers = d.find_element(
                By.XPATH,
                '//nav[contains(@class, "topbar")]//a[contains(normalize-space(), "My Servers")]',
            )
            my_containers = d.find_element(
                By.XPATH,
                '//nav[contains(@class, "topbar")]//a[contains(normalize-space(), "My Containers")]',
            )
            diagnostics = d.find_element(
                By.XPATH,
                '//nav[contains(@class, "topbar")]//a[contains(normalize-space(), "Diagnostics")]',
            )
            logs = d.find_element(
                By.XPATH,
                '//nav[contains(@class, "topbar")]//a[contains(normalize-space(), "Logs")]',
            )
            ping_host = d.find_element(By.CSS_SELECTOR, '[title="Ping host"]')
            restart = d.find_element(By.CSS_SELECTOR, '[title="Restart Docker Container"]')
            delete_backups = d.find_element(By.CSS_SELECTOR, '[title="Delete old backups"]')
        except Exception:
            return False

        return topbar.is_displayed() and all(
            element.is_displayed()
            for element in (
                my_servers,
                my_containers,
                diagnostics,
                logs,
                ping_host,
                restart,
                delete_backups,
            )
        )

    WebDriverWait(driver, timeout).until(ready)


def run(driver):
    _wait_for_topbar_navigation(driver)
    time.sleep(0.2)
