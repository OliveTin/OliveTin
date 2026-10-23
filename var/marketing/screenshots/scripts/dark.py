"""Forces the dark color scheme before capture."""
import os
import sys

sys.path.insert(0, os.path.dirname(__file__))
from theme import set_color_scheme  # noqa: E402


def run(driver):
    set_color_scheme(driver, "dark")
