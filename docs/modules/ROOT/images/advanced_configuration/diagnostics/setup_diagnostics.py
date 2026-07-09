#!/usr/bin/env python3
"""Open the diagnostics page from advanced_configuration/diagnostics.adoc."""

import time

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait


def _wait_for_diagnostics_page(driver, timeout=30):
    WebDriverWait(driver, timeout).until(
        lambda d: d.execute_script("return !!window.client")
    )

    def ready(d):
        try:
            ssh_heading = d.find_element(
                By.XPATH,
                '//*[self::h2 or self::h3][contains(normalize-space(), "SSH")]',
            )
            server_diagnostics_heading = d.find_element(
                By.XPATH,
                '//*[self::h2 or self::h3][contains(normalize-space(), "Server Diagnostics")]',
            )
            key_value = d.find_element(
                By.XPATH,
                '//dt[contains(normalize-space(), "Found Key")]/following-sibling::dd[1]',
            )
            config_value = d.find_element(
                By.XPATH,
                '//dt[contains(normalize-space(), "Found Config")]/following-sibling::dd[1]',
            )
        except Exception:
            return False

        return all(
            element.is_displayed()
            for element in (
                ssh_heading,
                server_diagnostics_heading,
                key_value,
                config_value,
            )
        ) and all(
            element.text.strip() not in ("", "?")
            for element in (key_value, config_value)
        )

    WebDriverWait(driver, timeout).until(ready)


def run(driver):
    _wait_for_diagnostics_page(driver)
    time.sleep(0.2)
