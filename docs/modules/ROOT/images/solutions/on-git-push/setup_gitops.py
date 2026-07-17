#!/usr/bin/env python3
"""Open the GitOps dashboard from solutions/on-git-push/index.adoc."""

import time

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait


def _wait_for_gitops_dashboard(driver, timeout=30):
    WebDriverWait(driver, timeout).until(
        lambda d: d.find_element(By.TAG_NAME, "body").get_attribute("loaded-dashboard")
        == "GitOps"
    )

    def ready(d):
        required_titles = [
            "ServerConfiguration Ansible",
            "ServerConfiguration Flux",
            "ServerConfiguration Prom",
        ]
        for title in required_titles:
            try:
                button = d.find_element(By.CSS_SELECTOR, f'[title="{title}"]')
            except Exception:
                return False
            if not button.is_displayed():
                return False
        return True

    WebDriverWait(driver, timeout).until(ready)


def run(driver):
    _wait_for_gitops_dashboard(driver)
    time.sleep(0.2)
