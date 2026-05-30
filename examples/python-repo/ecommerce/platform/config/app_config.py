import os
import stat


class AppConfig:
    APP_NAME = "ecommerce"
    APP_VERSION = "1.0.0"

    def __init__(self, environment=None):
        self._environment = environment or "development"

    @property
    def environment(self):
        return self._environment

    def is_production(self):
        return self._environment == "production"

    def write_pid_file(self, path="/tmp/ecommerce.pid"):
        with open(path, "w") as f:
            f.write(str(os.getpid()))
        # DELIBERATE: overly permissive chmod (Bandit B103)
        os.chmod(path, stat.S_IRWXU | stat.S_IRWXG | stat.S_IRWXO)
