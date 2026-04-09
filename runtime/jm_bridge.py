#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import argparse
import json
import os
import re
import sys
from pathlib import Path
from urllib.parse import urlparse


def configure_stdio():
    for stream_name in ("stdout", "stderr"):
        stream = getattr(sys, stream_name, None)
        if stream is None:
            continue

        reconfigure = getattr(stream, "reconfigure", None)
        if callable(reconfigure):
            try:
                reconfigure(encoding="utf-8", errors="replace")
            except Exception:
                pass


configure_stdio()


def emit(event):
    print(json.dumps(event, ensure_ascii=False), flush=True)


def fail(message, code=1):
    emit({"type": "error", "message": message})
    return code


def parse_target(target):
    value = (target or "").strip()
    if not value:
        raise ValueError("empty target")

    if value.lower().startswith("p") and value[1:].isdigit():
        return "photo", value[1:]

    if value.isdigit():
        return "album", value

    parsed = urlparse(value)
    parts = [part for part in parsed.path.split("/") if part]

    for index, part in enumerate(parts):
        label = part.lower()
        if label in {"photo", "album"} and index + 1 < len(parts):
            match = re.search(r"(\d+)", parts[index + 1])
            if match:
                return label, match.group(1)

    fallback = re.search(r"(\d+)", "/".join(parts))
    if fallback:
        kind = "photo" if "photo" in value.lower() else "album"
        return kind, fallback.group(1)

    raise ValueError(f"unsupported JM target: {target}")


def create_option(base_dir, proxy_url):
    try:
        from jmcomic import JmOption
    except Exception as exc:
        raise RuntimeError(
            "failed to import jmcomic. Build the packaged runtime or install jmcomic in the current Python environment."
        ) from exc

    proxies = {}
    if proxy_url:
        proxies = {"http": proxy_url, "https": proxy_url}

    option = JmOption(
        dir_rule={
            "rule": "Bd_Pname",
            "base_dir": base_dir,
        },
        download={
            "cache": True,
            "image": {
                "decode": True,
                "suffix": None,
            },
            "threading": {
                "image": 20,
                "photo": 12,
                "max_workers": 2,
            },
        },
        client={
            "cache": None,
            "domain": [],
            "postman": {
                "type": "curl_cffi",
                "meta_data": {
                    "impersonate": "chrome",
                    "headers": None,
                    "proxies": proxies,
                },
            },
            "impl": "api",
            "retry_times": 5,
        },
        plugins={},
    )

    return option


def build_client(option):
    if hasattr(option, "new_jm_client"):
        return option.new_jm_client()
    if hasattr(option, "build_jm_client"):
        return option.build_jm_client()
    raise RuntimeError("jmcomic client builder not found")


def detect_save_path(base_dir):
    root = Path(base_dir)
    if not root.exists():
        return str(root)

    directories = [path for path in root.iterdir() if path.is_dir()]
    if not directories:
        return str(root)

    directories.sort(key=lambda item: item.stat().st_mtime, reverse=True)
    return str(directories[0])


def run_download(target, output_dir, proxy_url):
    emit({"type": "status", "phase": "preparing", "message": "初始化 JM 下载引擎"})

    try:
        import jmcomic
    except Exception as exc:
        return fail(f"unable to import jmcomic: {exc}")

    target_type, target_id = parse_target(target)
    option = create_option(output_dir, proxy_url)
    client = build_client(option)

    album_name = None
    if target_type == "album":
        try:
            album = client.get_album_detail(target_id)
            album_name = getattr(album, "title", None) or f"JM {target_id}"
            emit({"type": "name", "name": album_name})
        except Exception:
            album_name = f"JM {target_id}"
            emit({"type": "name", "name": album_name})
    else:
        emit({"type": "name", "name": f"JM Photo {target_id}"})

    def progress_callback(current, total, info):
        emit(
            {
                "type": "progress",
                "current": int(current or 0),
                "total": int(total or 0),
                "message": str(info or ""),
            }
        )

    option.progress_callback = progress_callback
    emit({"type": "status", "phase": "downloading", "message": "开始下载"})

    try:
        if target_type == "photo":
            jmcomic.download_photo(target_id, option)
        else:
            jmcomic.download_album(target_id, option)
    except Exception as exc:
        return fail(f"jmcomic download failed: {exc}")

    save_path = detect_save_path(output_dir)
    emit({"type": "result", "savePath": save_path, "name": album_name or f'JM {target_id}'})
    return 0


def main():
    parser = argparse.ArgumentParser(description="ImageMaster JM bridge")
    parser.add_argument("--action", default="download")
    parser.add_argument("--target", required=True)
    parser.add_argument("--output", required=True)
    parser.add_argument("--proxy", default="")
    args = parser.parse_args()

    os.makedirs(args.output, exist_ok=True)

    if args.action != "download":
        return fail(f"unsupported action: {args.action}")

    try:
        return run_download(args.target, args.output, args.proxy)
    except Exception as exc:
        return fail(str(exc))


if __name__ == "__main__":
    raise SystemExit(main())
