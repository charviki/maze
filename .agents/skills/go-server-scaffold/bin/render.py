#!/usr/bin/env python3
"""Render the go-server-scaffold templates into a concrete module directory."""

from __future__ import annotations

import argparse
import re
import shutil
import sys
from pathlib import Path


PLACEHOLDER_PATTERN = re.compile(r"\{\{\.(?P<name>[A-Za-z0-9_]+)\}\}")

# 模板文件名 -> 输出路径模板（相对于 output-dir）
# ServiceName 由 render_text 在路径中替换
TEMPLATE_MAP: dict[str, str] = {
    "go.mod.tmpl": "go.mod",
    "cmd_main.go.tmpl": "cmd/{{.ServiceName}}/main.go",
    "cmd_setup.go.tmpl": "cmd/{{.ServiceName}}/setup.go",
    "config.go.tmpl": "internal/config/config.go",
    "health_service.go.tmpl": "internal/service/health_service.go",
    "grpc_server.go.tmpl": "internal/transport/grpc_server.go",
    "http_health.go.tmpl": "internal/transport/http_health.go",
}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Render the .skills/go-server-scaffold templates into a concrete Go module "
            "directory with deterministic placeholder replacement."
        )
    )
    parser.add_argument(
        "--output-dir",
        required=True,
        help="Target module directory to create, for example /path/to/repo/my-service.",
    )
    parser.add_argument(
        "--service-name",
        required=True,
        help="Service name used for cmd/<service-name>/ and logger name.",
    )
    parser.add_argument(
        "--module-path",
        required=True,
        help="Go module path written into go.mod and imports.",
    )
    parser.add_argument(
        "--env-prefix",
        required=True,
        help="Environment variable prefix consumed by configutil.ApplyEnvOverrides.",
    )
    parser.add_argument(
        "--cradle-replace",
        help="Optional replace target for github.com/charviki/maze-cradle.",
    )
    parser.add_argument(
        "--force",
        action="store_true",
        help="Overwrite an existing output directory if it already exists.",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    skill_root = Path(__file__).resolve().parents[1]
    repo_root = Path(__file__).resolve().parents[3]
    template_root = skill_root / "template"
    output_dir = Path(args.output_dir).expanduser().resolve()
    cradle_replace = args.cradle_replace or str((repo_root / "fabrication" / "cradle").resolve())

    variables = {
        "ServiceName": args.service_name,
        "ModulePath": args.module_path,
        "EnvPrefix": args.env_prefix,
        "CradleReplace": cradle_replace,
    }

    validate_inputs(output_dir=output_dir, template_root=template_root, variables=variables)
    prepare_output_dir(output_dir=output_dir, force=args.force)
    render_tree(template_root=template_root, output_dir=output_dir, variables=variables)

    print(f"[render] generated module at {output_dir}")
    print(f"[render] cradle replace => {cradle_replace}")
    return 0


def validate_inputs(*, output_dir: Path, template_root: Path, variables: dict[str, str]) -> None:
    if not template_root.is_dir():
        raise SystemExit(f"template root does not exist: {template_root}")
    if variables["ServiceName"].strip() == "":
        raise SystemExit("service name must not be empty")
    if variables["ModulePath"].strip() == "":
        raise SystemExit("module path must not be empty")
    if variables["EnvPrefix"].strip() == "":
        raise SystemExit("env prefix must not be empty")
    if output_dir == output_dir.anchor:
        raise SystemExit("refuse to render into filesystem root")
    # 确认每个映射的模板文件都存在
    for tmpl_name in TEMPLATE_MAP:
        if not (template_root / tmpl_name).is_file():
            raise SystemExit(f"missing template file: {tmpl_name}")


def prepare_output_dir(*, output_dir: Path, force: bool) -> None:
    if output_dir.exists():
        if not force:
            raise SystemExit(
                f"output directory already exists: {output_dir} (pass --force to overwrite)"
            )
        shutil.rmtree(output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)


def render_tree(*, template_root: Path, output_dir: Path, variables: dict[str, str]) -> None:
    for tmpl_name, dest_pattern in TEMPLATE_MAP.items():
        source = template_root / tmpl_name
        # 路径中的占位符替换（如 cmd/{{.ServiceName}}/main.go）
        dest_relative = render_text(dest_pattern, variables)
        target = output_dir / dest_relative

        target.parent.mkdir(parents=True, exist_ok=True)
        content = source.read_text(encoding="utf-8")
        target.write_text(render_text(content, variables), encoding="utf-8")


def render_text(text: str, variables: dict[str, str]) -> str:
    def replace(match: re.Match[str]) -> str:
        key = match.group("name")
        try:
            return variables[key]
        except KeyError as exc:
            raise SystemExit(f"missing template variable: {key}") from exc

    return PLACEHOLDER_PATTERN.sub(replace, text)


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except BrokenPipeError:
        # Allow piping to head/jq without noisy tracebacks.
        sys.exit(0)
