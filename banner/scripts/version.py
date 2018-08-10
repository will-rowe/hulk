# Retrieve the version number from the setup.py file.
# This solution was found in the ISMmapper repo and suggested on Stack Overflow:
# http://stackoverflow.com/questions/2058802/how-can-i-get-the-version-defined-in-setup-py-setuptools-in-my-package
import pkg_resources  # part of setuptools

version_number = pkg_resources.require("banner")[0].version
