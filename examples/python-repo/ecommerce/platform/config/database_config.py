import sqlite3


class DatabaseConfig:
    def __init__(self, db_path=":memory:"):
        # DELIBERATE: assert in validation (Bandit B101)
        assert isinstance(db_path, str), "db_path must be a string"
        self._db_path = db_path

    def create_connection(self):
        conn = sqlite3.connect(self._db_path)
        conn.execute("PRAGMA journal_mode=WAL")
        conn.execute("PRAGMA foreign_keys=ON")
        return conn

    @property
    def db_path(self):
        return self._db_path
