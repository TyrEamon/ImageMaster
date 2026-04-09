var DMZJ_BASE_URL = "https://www.dmzj.com";
var DMZJ_PAGE_SIZE = 20;
var DMZJ_RANKING_SIZE = 24;

function trim(value) {
  if (value === null || value === undefined) {
    return "";
  }
  return String(value).replace(/^\s+|\s+$/g, "");
}

function fallback(value, fallbackValue) {
  value = trim(value);
  return value ? value : fallbackValue;
}

function detailUrl(comicPy) {
  comicPy = trim(comicPy);
  if (!comicPy) {
    return DMZJ_BASE_URL;
  }
  return DMZJ_BASE_URL + "/info/" + comicPy + ".html";
}

function chapterUrl(comicID, chapterID) {
  return DMZJ_BASE_URL + "/view/" + trim(comicID) + "/" + trim(chapterID) + ".html";
}

function defaultHeaders() {
  return {
    "Accept": "application/json,text/plain,*/*",
    "Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
    "Referer": DMZJ_BASE_URL + "/"
  };
}

function toArray(value) {
  if (!value) {
    return [];
  }
  if (Object.prototype.toString.call(value) === "[object Array]") {
    return value;
  }
  return [];
}

function parsePageUrls(raw) {
  if (!raw) {
    return [];
  }

  if (typeof raw === "string") {
    var text = trim(raw);
    if (!text) {
      return [];
    }
    try {
      raw = JSON.parse(text);
    } catch (err) {
      raw = [text];
    }
  }

  var list = toArray(raw);
  var result = [];
  for (var index = 0; index < list.length; index += 1) {
    var item = trim(list[index]);
    if (item) {
      result.push(item);
    }
  }

  return result;
}

function search(query, page, runtime) {
  query = trim(query);
  page = Number(page || 1);
  if (page < 1) {
    page = 1;
  }

  if (!query) {
    return {
      query: "",
      page: 1,
      hasMore: false,
      total: 0,
      items: []
    };
  }

  var payload = runtime.getJSON(
    DMZJ_BASE_URL + "/api/v1/comic1/search?keyword=" + runtime.urlQuery(query) + "&page=" + page,
    defaultHeaders()
  );
  var list = (((payload || {}).data || {}).comic_list) || [];
  var items = [];

  for (var index = 0; index < list.length; index += 1) {
    var comic = list[index] || {};
    var comicPy = trim(comic.comic_py);
    var title = trim(comic.name);
    if (!comicPy || !title) {
      continue;
    }

    items.push({
      id: comicPy,
      title: title,
      cover: trim(comic.cover),
      summary: fallback(comic.last_update_chapter_name, "动漫之家漫画"),
      primaryLabel: fallback(comic.authors, "动漫之家"),
      secondaryLabel: trim(comic.last_update_chapter_name),
      detailUrl: detailUrl(comicPy)
    });
  }

  return {
    query: query,
    page: page,
    hasMore: items.length >= DMZJ_PAGE_SIZE,
    total: items.length,
    items: items
  };
}

function ranking(kind, page, runtime) {
  kind = trim(kind || "latest").toLowerCase();
  page = Number(page || 1);
  if (page < 1) {
    page = 1;
  }

  var payload = runtime.getJSON(
    DMZJ_BASE_URL +
      "/api/v1/comic1/update_list?channel=pc&app_name=dmzj&version=1.0.0&page=" +
      page +
      "&size=" +
      DMZJ_RANKING_SIZE,
    defaultHeaders()
  );

  var list = (((payload || {}).data || {}).list) || [];
  var items = [];

  for (var index = 0; index < list.length; index += 1) {
    var comic = list[index] || {};
    var comicPy = trim(comic.comic_py);
    var title = trim(comic.title);
    if (!comicPy || !title) {
      continue;
    }

    items.push({
      id: comicPy,
      title: title,
      cover: trim(comic.cover),
      summary: fallback(comic.lastUpdateChapterName, "最新更新"),
      primaryLabel: "动漫之家",
      secondaryLabel: trim(comic.lastUpdateChapterName),
      detailUrl: detailUrl(comicPy)
    });
  }

  return {
    kind: kind,
    page: page,
    total: items.length,
    items: items
  };
}

function detail(itemID, runtime) {
  itemID = trim(itemID);
  if (!itemID) {
    throw new Error("missing dmzj item id");
  }

  var payload = runtime.getJSON(
    DMZJ_BASE_URL +
      "/api/v1/comic1/comic/detail?channel=pc&app_name=dmzj&version=1.0.0&comic_py=" +
      runtime.urlQuery(itemID),
    defaultHeaders()
  );

  var comic = (((payload || {}).data || {}).comicInfo) || {};
  if (!trim(comic.title)) {
    throw new Error("dmzj detail returned empty comic info");
  }

  var groups = comic.chapterList || [];
  var chapters = [];
  for (var groupIndex = 0; groupIndex < groups.length; groupIndex += 1) {
    var group = groups[groupIndex] || {};
    var rawChapters = group.data || [];
    var reversed = [];

    for (var chapterIndex = 0; chapterIndex < rawChapters.length; chapterIndex += 1) {
      var chapter = rawChapters[chapterIndex] || {};
      var chapterID = trim(chapter.chapter_id);
      if (!chapterID) {
        continue;
      }

      reversed.push({
        id: trim(comic.id) + "|" + chapterID,
        name: trim(chapter.chapter_title),
        url: trim(comic.id) + "|" + chapterID,
        index: chapters.length + reversed.length,
        updatedLabel: trim(group.title)
      });
    }

    reversed.reverse();
    chapters = chapters.concat(reversed);
  }

  var tags = [];
  if (trim(comic.zone)) {
    tags.push(trim(comic.zone));
  }
  if (trim(comic.status)) {
    tags.push(trim(comic.status));
  }

  return {
    item: {
      id: itemID,
      title: trim(comic.title),
      cover: trim(comic.cover),
      summary: fallback(comic.description, "No summary available."),
      author: fallback(comic.authors, "Unknown author"),
      status: fallback(comic.status, "Unknown"),
      tags: tags,
      detailUrl: detailUrl(itemID),
      chapters: chapters
    }
  };
}

function images(chapterID, runtime) {
  chapterID = trim(chapterID);
  var parts = chapterID.split("|");
  if (parts.length !== 2) {
    throw new Error("invalid dmzj chapter id: " + chapterID);
  }

  var comicID = trim(parts[0]);
  var rawChapterID = trim(parts[1]);
  if (!comicID || !rawChapterID) {
    throw new Error("invalid dmzj chapter id: " + chapterID);
  }

  var payload = runtime.getJSON(
    DMZJ_BASE_URL +
      "/api/v1/comic1/chapter/detail?channel=pc&app_name=dmzj&version=1.0.0&comic_id=" +
      runtime.urlQuery(comicID) +
      "&chapter_id=" +
      runtime.urlQuery(rawChapterID),
    defaultHeaders()
  );

  var info = (((payload || {}).data || {}).chapterInfo) || {};
  var urls = parsePageUrls(info.page_url_hd);
  if (!urls.length) {
    urls = parsePageUrls(info.page_url);
  }
  if (!urls.length) {
    throw new Error("dmzj chapter returned no readable page urls");
  }

  var entries = [];
  for (var index = 0; index < urls.length; index += 1) {
    entries.push({ url: urls[index] });
  }

  return {
    comicTitle: fallback(info.comic_title, "动漫之家"),
    chapterTitle: fallback(info.chapter_title, "Chapter " + rawChapterID),
    chapterUrl: chapterUrl(comicID, rawChapterID),
    images: urls,
    entries: entries,
    hasNext: false,
    nextUrl: ""
  };
}

var source = {
  search: search,
  ranking: ranking,
  detail: detail,
  images: images
};
