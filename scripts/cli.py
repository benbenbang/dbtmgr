import tomlkit
from pathlib import Path


def verkit():
    project_root = Path(__file__).parent.parent
    pyproject_toml = tomlkit.loads((project_root / "pyproject.toml").read_text())
    version = pyproject_toml.get("tool", {}).get("poetry", {}).get("version", "")
    Path(project_root / "version.info").write_text(version)


if __name__ == "__main__":
    verkit()
