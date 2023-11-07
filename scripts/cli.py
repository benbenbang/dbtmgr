import tomlkit
from pathlib import Path
from rich import print


def verkit():
    print("[cyan]ℹ️ generating [bold]version.info[/bold] file...[/cyan]")
    project_root = Path(__file__).parent.parent
    pyproject_toml = tomlkit.loads((project_root / "pyproject.toml").read_text())
    version = pyproject_toml.get("tool", {}).get("poetry", {}).get("version", "")
    Path(project_root / "version.info").write_text(version)
    print(f"[green]🎉 successfully generated version.info file with version: [bold]{version}[/bold][/green]")


if __name__ == "__main__":
    verkit()
