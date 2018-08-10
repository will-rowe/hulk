#!/usr/bin/env python3
__author__ = 'Will Rowe'
__mail__ = "will.rowe@stfc.ac.uk"

# Install a project in editable mode
# pip install -e . --user

from setuptools import setup

setup(
    name = "banner",
    version = '0.0.1',
    packages = ["banner"],
    author = 'Will Rowe',
    author_email = 'will.rowe@stfc.ac.uk',
    url = 'http://will-rowe.github.io/',
    description = '...',
    long_description = open('README.md').read(),
    package_dir = {'banner': 'scripts'},
    entry_points = {
        "console_scripts": ['banner = banner.banner:Banner']
        },
    install_requires = [
        "numpy == 1.15.0",
        "sklearn == 0.0",
        "pandas == 0.23.4",
    ],
)
