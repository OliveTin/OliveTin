"""Pins prefers-color-scheme; OliveTin's dark mode follows the browser.

repo-helper reuses one browser session for every section, so each section
must set the scheme explicitly or it inherits the previous one.
"""


def set_color_scheme(driver, scheme):
    driver.execute_cdp_cmd(
        "Emulation.setEmulatedMedia",
        {"features": [{"name": "prefers-color-scheme", "value": scheme}]},
    )
