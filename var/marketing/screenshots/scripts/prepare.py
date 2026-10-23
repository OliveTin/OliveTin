"""Waits for the SPA to render before capture."""
import time

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait


def _rendered(driver):
    if driver.find_element(By.TAG_NAME, "body").get_attribute("loaded-dashboard"):
        return True
    if driver.find_elements(By.CSS_SELECTOR, "main table tbody tr"):
        return True
    return any(rows.text.strip() for rows in driver.find_elements(By.CSS_SELECTOR, ".xterm-rows"))


def run(driver):
    WebDriverWait(driver, 30).until(_rendered)
    time.sleep(1)
    driver.execute_script("window.scrollTo(0, 0);")
