import os

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait

from repo_helper.screenshot import set_viewport_size


def _enable_dark_mode(driver):
    driver.execute_cdp_cmd(
        "Emulation.setEmulatedMedia",
        {"features": [{"name": "prefers-color-scheme", "value": "dark"}]},
    )


def run(driver):
    WebDriverWait(driver, 30).until(
        lambda d: bool(d.find_element(By.TAG_NAME, "body").get_attribute("loaded-dashboard"))
    )
    if os.environ.get("SHOT_DARK") == "1":
        _enable_dark_mode(driver)
    width = int(os.environ["SHOT_WIDTH"])
    height = int(os.environ["SHOT_HEIGHT"])
    set_viewport_size(driver, width, height)
