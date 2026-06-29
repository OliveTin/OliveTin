import os

from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait

from repo_helper.screenshot import set_viewport_size


def run(driver):
    WebDriverWait(driver, 30).until(
        lambda d: bool(d.find_element(By.TAG_NAME, "body").get_attribute("loaded-dashboard"))
    )
    width = int(os.environ["SHOT_WIDTH"])
    height = int(os.environ["SHOT_HEIGHT"])
    set_viewport_size(driver, width, height)
