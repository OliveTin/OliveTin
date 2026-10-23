"""Opens the argument form for the "Ping host" action."""
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait


def run(driver):
    driver.find_element(By.CSS_SELECTOR, '[title="Ping host"]').click()
    WebDriverWait(driver, 15).until(
        lambda d: d.find_element(By.TAG_NAME, "body").get_attribute("loaded-argument-form")
    )
