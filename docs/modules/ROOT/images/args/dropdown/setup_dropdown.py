#!/usr/bin/env python3
"""Prepare the dropdown argument form screenshot."""

import time

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait


def _wait_for_body_attr(driver, attr, timeout=15):
    WebDriverWait(driver, timeout).until(
        lambda d: bool(d.find_element(By.TAG_NAME, "body").get_attribute(attr))
    )


def _open_choice_list(driver, input_id):
    combobox_input = driver.find_element(By.ID, input_id)
    combobox_input.click()
    WebDriverWait(driver, 15).until(
        lambda d: len(
            d.find_elements(
                By.CSS_SELECTOR,
                f"#{input_id}-listbox li, #{input_id} + input + ul li",
            )
        )
        >= 2
        or len(d.find_elements(By.CSS_SELECTOR, ".choice-combobox-list li")) >= 2
    )


def run(driver):
    _wait_for_body_attr(driver, "loaded-dashboard")

    driver.find_element(By.CSS_SELECTOR, '[title="Print a message"]').click()
    _wait_for_body_attr(driver, "loaded-argument-form")
    _open_choice_list(driver, "message")
    time.sleep(0.2)
