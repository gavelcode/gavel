# DELIBERATE: wildcard import (Ruff F403)
from string import *


class Address:
    def __init__(self, street, city, zip_code, country):
        self._street = street
        self._city = city
        self._zip_code = zip_code
        self._country = country

    @property
    def street(self):
        return self._street

    @property
    def city(self):
        return self._city

    @property
    def zip_code(self):
        return self._zip_code

    @property
    def country(self):
        return self._country

    def formatted(self):
        return f"{self._street}, {self._city} {self._zip_code}, {self._country}"
