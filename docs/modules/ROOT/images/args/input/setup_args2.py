#!/usr/bin/env python3
"""Prepare the argument-form screenshot."""

import time

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait


def _wait_for_body_attr(driver, attr, timeout=15):
    WebDriverWait(driver, timeout).until(
        lambda d: bool(d.find_element(By.TAG_NAME, "body").get_attribute(attr))
    )


def _open_form(driver):
    _wait_for_body_attr(driver, "loaded-dashboard")

    action = driver.find_element(By.CSS_SELECTOR, '[title="Print a message"]')
    action.click()

    _wait_for_body_attr(driver, "loaded-argument-form")


def run(driver):
    _open_form(driver)

    driver.execute_script(
        """
        const input = document.getElementById('message');
        if (input && !input.value) {
          input.value = 'Hello World';
        }
        """
    )
    time.sleep(0.2)
