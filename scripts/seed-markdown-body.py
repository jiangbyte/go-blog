#!/usr/bin/env python3
"""将 docs/seed/01-markdown-kitchen-sink.md 写入 blog_post（Markdown 特性测试文）。"""
from __future__ import annotations

import os
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
MD = ROOT / "docs" / "seed" / "01-markdown-kitchen-sink.md"
UID = "01a128a9a64073408ce8976b098575fd"


def mysql_quote(s: str) -> str:
    return "'" + s.replace("\\", "\\\\").replace("'", "''") + "'"


def main() -> int:
    if not MD.is_file():
        print(f"skip: missing {MD}", file=sys.stderr)
        return 0
    body = MD.read_text(encoding="utf-8")
    sql = (
        f"UPDATE blog_post SET body = {mysql_quote(body)}, "
        f"word_count = {len(body)}, reading_time_minutes = 18 "
        f"WHERE uid = '{UID}';\n"
    )
    docker = os.environ.get("MYSQL_DOCKER", "mysql")
    user = os.environ.get("MYSQL_USER", "root")
    password = os.environ.get("MYSQL_PASSWORD", "infra123!")
    database = os.environ.get("MYSQL_DATABASE", "go_blog")
    cmd = [
        "docker",
        "exec",
        "-i",
        docker,
        "mysql",
        f"-u{user}",
        f"--password={password}",
        "--default-character-set=utf8mb4",
        database,
    ]
    r = subprocess.run(cmd, input=sql.encode("utf-8"), capture_output=True)
    if r.returncode != 0:
        sys.stderr.write(r.stderr.decode("utf-8", errors="replace"))
        return r.returncode
    print(f"seed markdown body → {UID} ({len(body)} chars)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
