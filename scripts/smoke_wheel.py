"""Install the built wheel in isolation and verify its CLI data files."""

import pathlib
import subprocess
import sys
import tempfile
import venv


def main():
    wheels = sorted(pathlib.Path("dist").glob("agentsmd-*.whl"))
    if len(wheels) != 1:
        raise SystemExit(f"expected one wheel in dist/, found {len(wheels)}")

    wheel = wheels[0].resolve()
    with tempfile.TemporaryDirectory(prefix="agentsmd-wheel-") as directory:
        environment = pathlib.Path(directory)
        venv.EnvBuilder(with_pip=True).create(environment)
        scripts = environment / ("Scripts" if sys.platform == "win32" else "bin")
        python = scripts / ("python.exe" if sys.platform == "win32" else "python")
        command = scripts / ("agentsmd.exe" if sys.platform == "win32" else "agentsmd")

        subprocess.run([python, "-m", "pip", "install", str(wheel)], check=True)
        subprocess.run([command, "--version"], check=True)
        result = subprocess.run(
            [command, "template", "list"], check=True, capture_output=True, text=True
        )
        if "minimal" not in result.stdout.splitlines():
            raise SystemExit("installed wheel cannot find the bundled minimal template")

    print(f"wheel smoke test passed: {wheel.name}")


if __name__ == "__main__":
    main()
