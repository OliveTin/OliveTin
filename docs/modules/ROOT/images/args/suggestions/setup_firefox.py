#!/usr/bin/env python3
"""Prepare the Firefox-style suggestions screenshot."""

import time

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait


def _wait_for_body_attr(driver, attr, timeout=15):
    WebDriverWait(driver, timeout).until(
        lambda d: bool(d.find_element(By.TAG_NAME, "body").get_attribute(attr))
    )


def _open_form(driver):
    _wait_for_body_attr(driver, "loaded-dashboard")

    action = driver.find_element(
        By.CSS_SELECTOR, '[title="Restart Docker Container"]'
    )
    action.click()

    _wait_for_body_attr(driver, "loaded-argument-form")

    driver.execute_script(
        """
        const form = document.getElementById('argument-popup');
        if (form) {
          form.style.margin = '2rem auto';
          form.style.maxWidth = '520px';
        }
        """
    )


def run(driver):
    _open_form(driver)

    driver.execute_script(
        """
        const input = document.getElementById('container');
        input.focus();
        input.value = '';

        document.getElementById('doc-suggestions-overlay')?.remove();

        const rect = input.getBoundingClientRect();
        const menu = document.createElement('div');
        menu.id = 'doc-suggestions-overlay';
        menu.style.position = 'fixed';
        menu.style.left = `${rect.left}px`;
        menu.style.top = `${rect.bottom + 2}px`;
        menu.style.width = `${rect.width}px`;
        menu.style.background = '#fff';
        menu.style.border = '1px solid #ccc';
        menu.style.boxShadow = '0 1px 4px rgba(0, 0, 0, 0.15)';
        menu.style.font = '13px sans-serif';
        menu.style.zIndex = '9999';

        for (const label of [
          'Firewall Controller',
          'graefik',
          'grafana',
          'plex',
          'WiFi Controller',
        ]) {
          const row = document.createElement('div');
          row.textContent = label;
          row.style.padding = '4px 8px';
          menu.appendChild(row);
        }

        document.body.appendChild(menu);
        """
    )
    time.sleep(0.2)
