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
    emit({"type": "error", "message": str(message)})
    return code


def safe_attr(obj, *names, default=""):
    for name in names:
        if not hasattr(obj, name):
            continue
        try:
            value = getattr(obj, name)
            if callable(value):
                value = value()
            if value is not None and value != "":
                return value
        except Exception:
            continue
    return default


def normalize_list(value):
    if value is None:
        return []
    if isinstance(value, (list, tuple, set)):
        return [str(item).strip() for item in value if str(item).strip()]
    if isinstance(value, str):
        return [part.strip() for part in value.split(",") if part.strip()]
    return [str(value).strip()] if str(value).strip() else []


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


def get_properties(obj):
    getter = getattr(obj, "get_properties_dict", None)
    if callable(getter):
        try:
            data = getter()
            if isinstance(data, dict):
                return data
        except Exception:
            pass
    return {}


def normalize_search_page(page):
    items = []

    iterator = getattr(page, "iter_id_title_tag", None)
    if callable(iterator):
        try:
            for album_id, title, tags in iterator():
                items.append(
                    {
                        "id": str(album_id),
                        "title": str(title).strip(),
                        "cover": "",
                        "summary": " / ".join(normalize_list(tags)),
                        "primaryLabel": "JM",
                        "secondaryLabel": " / ".join(normalize_list(tags)),
                        "detailUrl": f"https://18comic.vip/album/{album_id}",
                    }
                )
            return items
        except Exception:
            pass

    for raw in list(page):
        if isinstance(raw, (list, tuple)) and len(raw) >= 2:
            album_id = str(raw[0]).strip()
            title = str(raw[1]).strip()
            tags = normalize_list(raw[2] if len(raw) > 2 else [])
        else:
            album_id = str(safe_attr(raw, "id", "album_id", default="")).strip()
            title = str(safe_attr(raw, "title", "name", default=str(raw))).strip()
            tags = normalize_list(safe_attr(raw, "tag_list", "tags", default=[]))

        if not album_id or not title:
            continue

        items.append(
            {
                "id": album_id,
                "title": title,
                "cover": "",
                "summary": " / ".join(tags),
                "primaryLabel": "JM",
                "secondaryLabel": " / ".join(tags),
                "detailUrl": f"https://18comic.vip/album/{album_id}",
            }
        )

    return items


def normalize_album_detail(album):
    props = get_properties(album)
    title = str(props.get("title") or safe_attr(album, "title", "name", default="Untitled")).strip()
    author = str(props.get("author") or safe_attr(album, "author", default="Unknown")).strip()
    status = str(props.get("status") or safe_attr(album, "status", default="Unknown")).strip()
    summary = str(
        props.get("comment")
        or props.get("description")
        or safe_attr(album, "comment", "description", default="No summary available.")
    ).strip()
    album_id = str(props.get("album_id") or safe_attr(album, "album_id", "id", default="")).strip()

    tags = []
    for key in ("tag_list", "tags", "works", "actors"):
        tags.extend(normalize_list(props.get(key)))
    # dedupe while preserving order
    deduped_tags = list(dict.fromkeys([tag for tag in tags if tag]))

    chapters = []
    try:
        for index, photo in enumerate(album):
            photo_props = get_properties(photo)
            photo_id = str(photo_props.get("photo_id") or safe_attr(photo, "photo_id", "id", default="")).strip()
            photo_name = str(photo_props.get("title") or safe_attr(photo, "title", "name", default="")).strip()
            if not photo_name:
                photo_name = f"Chapter {index + 1}"
            if not photo_id:
                continue
            chapters.append(
                {
                    "id": f"p{photo_id}" if not str(photo_id).startswith("p") else str(photo_id),
                    "name": photo_name,
                    "url": f"https://18comic.vip/photo/{photo_id}",
                    "index": index,
                    "updatedLabel": "",
                }
            )
    except Exception:
        pass

    return {
        "item": {
            "id": album_id or safe_attr(album, "album_id", "id", default=""),
            "title": title or "Untitled",
            "cover": "",
            "summary": summary or "No summary available.",
            "author": author or "Unknown",
            "status": status or "Unknown",
            "tags": deduped_tags,
            "detailUrl": f"https://18comic.vip/album/{album_id}" if album_id else "",
            "chapters": chapters,
        }
    }


def normalize_photo_images(photo):
    photo_props = get_properties(photo)
    chapter_title = str(photo_props.get("title") or safe_attr(photo, "title", "name", default="Online chapter")).strip()
    photo_id = str(photo_props.get("photo_id") or safe_attr(photo, "photo_id", "id", default="")).strip()
    chapter_url = f"https://18comic.vip/photo/{photo_id}" if photo_id else ""

    comic_title = ""
    album = safe_attr(photo, "from_album", default=None)
    if album:
        comic_title = str(safe_attr(album, "title", "name", default="")).strip()
    if not comic_title:
        comic_title = str(photo_props.get("series") or photo_props.get("album_name") or "").strip()

    images = []
    entries = []
    try:
        for image in photo:
            image_url = str(safe_attr(image, "img_url", "url", default=image)).strip()
            if not image_url:
                continue
            images.append(image_url)
            entries.append({"url": image_url, "referer": chapter_url, "headers": {"Referer": chapter_url}})
    except Exception:
        pass

    return {
        "comicTitle": comic_title or "JM",
        "chapterTitle": chapter_title or "Online chapter",
        "chapterUrl": chapter_url,
        "images": images,
        "entries": entries,
        "hasNext": False,
        "nextUrl": "",
    }


def run_download(target, output_dir, proxy_url):
    emit({"type": "status", "phase": "preparing", "message": "Initialize JM download engine"})

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
    emit({"type": "status", "phase": "downloading", "message": "Start downloading"})

    try:
        if target_type == "photo":
            jmcomic.download_photo(target_id, option)
        else:
            jmcomic.download_album(target_id, option)
    except Exception as exc:
        return fail(f"jmcomic download failed: {exc}")

    save_path = detect_save_path(output_dir)
    emit({"type": "result", "savePath": save_path, "name": album_name or f"JM {target_id}"})
    return 0


def run_search(query, page, proxy_url):
    option = create_option(os.getcwd(), proxy_url)
    client = build_client(option)
    search_page = client.search_site(query)
    payload = {
        "query": query,
        "page": page,
        "hasMore": False,
        "total": 0,
        "items": [],
    }
    items = normalize_search_page(search_page)
    payload["items"] = items
    payload["total"] = len(items)
    emit({"type": "result", "payload": payload})
    return 0


def run_ranking(kind, page, proxy_url):
    option = create_option(os.getcwd(), proxy_url)
    client = build_client(option)

    ranking_page = None
    errors = []
    try:
        method_name = f"{kind}_ranking"
        if hasattr(client, method_name):
            ranking_page = getattr(client, method_name)(page=page)
    except Exception as exc:
        errors.append(str(exc))

    if ranking_page is None:
        for name in (kind, f"{kind}_ranking"):
            try:
                ranking_page = client.ranking(name, page=page)
                if ranking_page is not None:
                    break
            except Exception as exc:
                errors.append(str(exc))

    if ranking_page is None:
        return fail("unable to load ranking: " + " | ".join(errors))

    items = normalize_search_page(ranking_page)
    emit(
        {
            "type": "result",
            "payload": {
                "kind": kind,
                "page": page,
                "total": len(items),
                "items": items,
            },
        }
    )
    return 0


def run_detail(target, proxy_url):
    option = create_option(os.getcwd(), proxy_url)
    client = build_client(option)
    target_type, target_id = parse_target(target)
    if target_type != "album":
        return fail("detail currently expects a JM album id or album url")
    album = client.get_album_detail(target_id)
    emit({"type": "result", "payload": normalize_album_detail(album)})
    return 0


def run_images(target, proxy_url):
    option = create_option(os.getcwd(), proxy_url)
    client = build_client(option)
    target_type, target_id = parse_target(target)
    if target_type != "photo":
        return fail("images currently expects a JM photo id or photo url")
    try:
        photo = client.get_photo_detail(target_id, False)
    except TypeError:
        photo = client.get_photo_detail(target_id)
    emit({"type": "result", "payload": normalize_photo_images(photo)})
    return 0


def main():
    parser = argparse.ArgumentParser(description="ImageMaster JM bridge")
    parser.add_argument("--action", required=True)
    parser.add_argument("--target", default="")
    parser.add_argument("--output", default="")
    parser.add_argument("--proxy", default="")
    parser.add_argument("--query", default="")
    parser.add_argument("--page", type=int, default=1)
    parser.add_argument("--kind", default="week")
    args = parser.parse_args()

    if args.action == "download":
        if not args.output:
            return fail("missing output directory for download action")
        os.makedirs(args.output, exist_ok=True)
        return run_download(args.target, args.output, args.proxy)
    if args.action == "search":
        return run_search(args.query.strip(), args.page, args.proxy)
    if args.action == "ranking":
        return run_ranking(args.kind.strip().lower(), args.page, args.proxy)
    if args.action == "detail":
        return run_detail(args.target, args.proxy)
    if args.action == "images":
        return run_images(args.target, args.proxy)

    return fail(f"unsupported action: {args.action}")


if __name__ == "__main__":
    raise SystemExit(main())
