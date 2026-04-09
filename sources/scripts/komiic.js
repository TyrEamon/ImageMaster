var KOMIIC_BASE_URL = "https://komiic.com";
var KOMIIC_API_URL = KOMIIC_BASE_URL + "/api/query";
var KOMIIC_PAGE_SIZE = 20;

var KOMIIC_QUERIES = {
  recentUpdate:
    "query recentUpdate($pagination: Pagination!) { recentUpdate(pagination: $pagination) { id title status year imageUrl authors { id name __typename } categories { id name __typename } dateUpdated monthViews views favoriteCount lastBookUpdate lastChapterUpdate __typename } }",
  searchComicAndAuthorQuery:
    "query searchComicAndAuthorQuery($keyword: String!) { searchComicsAndAuthors(keyword: $keyword) { comics { id title status year imageUrl authors { id name __typename } categories { id name __typename } dateUpdated monthViews views favoriteCount lastBookUpdate lastChapterUpdate __typename } authors { id name chName enName wikiLink comicCount views __typename } __typename } }",
  chapterByComicId:
    "query chapterByComicId($comicId: ID!) { chaptersByComicId(comicId: $comicId) { id serial type dateCreated dateUpdated size __typename } }",
  imagesByChapterId:
    "query imagesByChapterId($chapterId: ID!) { imagesByChapterId(chapterId: $chapterId) { id kid height width __typename } }"
};

function trim(value) {
  if (value === null || value === undefined) {
    return "";
  }
  return String(value).replace(/^\s+|\s+$/g, "");
}

function fallback(value, fallbackValue) {
  var text = trim(value);
  return text ? text : fallbackValue;
}

function toArray(value) {
  if (!value || Object.prototype.toString.call(value) !== "[object Array]") {
    return [];
  }
  return value;
}

function defaultHeaders() {
  return {
    "Accept": "application/json,text/plain,*/*",
    "Accept-Language": "zh-TW,zh;q=0.9,en;q=0.8",
    "Referer": KOMIIC_BASE_URL + "/"
  };
}

function detailUrl(comicId) {
  comicId = trim(comicId);
  if (!comicId) {
    return KOMIIC_BASE_URL;
  }
  return KOMIIC_BASE_URL + "/comic/" + comicId;
}

function chapterReferer(comicId, chapterId) {
  return KOMIIC_BASE_URL + "/comic/" + trim(comicId) + "/chapter/" + trim(chapterId) + "/images/all";
}

function encodeItem(data) {
  return JSON.stringify(data || {});
}

function decodeItem(raw) {
  if (!trim(raw)) {
    return {};
  }

  try {
    return JSON.parse(raw);
  } catch (error) {
    return { id: trim(raw) };
  }
}

function normalizeComicItem(comic) {
  comic = comic || {};

  var authorNames = [];
  var authors = toArray(comic.authors);
  for (var authorIndex = 0; authorIndex < authors.length; authorIndex += 1) {
    var authorName = trim((authors[authorIndex] || {}).name);
    if (authorName) {
      authorNames.push(authorName);
    }
  }

  var categoryNames = [];
  var categories = toArray(comic.categories);
  for (var categoryIndex = 0; categoryIndex < categories.length; categoryIndex += 1) {
    var categoryName = trim((categories[categoryIndex] || {}).name);
    if (categoryName) {
      categoryNames.push(categoryName);
    }
  }

  var itemData = {
    id: trim(comic.id),
    title: trim(comic.title),
    cover: trim(comic.imageUrl),
    status: trim(comic.status),
    year: trim(comic.year),
    author: authorNames.join("，"),
    categories: categoryNames,
    views: trim(comic.views),
    update: trim(comic.dateUpdated),
    lastChapterUpdate: trim(comic.lastChapterUpdate),
    lastBookUpdate: trim(comic.lastBookUpdate)
  };

  return {
    id: encodeItem(itemData),
    title: itemData.title,
    cover: itemData.cover,
    summary: buildSummary(itemData),
    primaryLabel: fallback(itemData.author, "Komiic"),
    secondaryLabel: buildSecondaryLabel(itemData),
    detailUrl: detailUrl(itemData.id)
  };
}

function buildSummary(itemData) {
  var parts = [];
  if (itemData.status) {
    parts.push(itemData.status === "ONGOING" ? "連載中" : "已完結");
  }
  if (itemData.categories && itemData.categories.length > 0) {
    parts.push(itemData.categories.join(" / "));
  }
  if (itemData.year) {
    parts.push(itemData.year);
  }
  return parts.join(" · ");
}

function buildSecondaryLabel(itemData) {
  var segments = [];
  if (itemData.lastChapterUpdate) {
    segments.push(itemData.lastChapterUpdate + "話");
  }
  if (itemData.lastBookUpdate) {
    segments.push(itemData.lastBookUpdate + "卷");
  }
  return segments.join(" | ");
}

function postGraphQL(operationName, variables, runtime) {
  return runtime.postJSON(
    KOMIIC_API_URL,
    {
      query: KOMIIC_QUERIES[operationName],
      operationName: operationName,
      variables: variables || {}
    },
    defaultHeaders()
  );
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

  if (page > 1) {
    return {
      query: query,
      page: page,
      hasMore: false,
      total: 0,
      items: []
    };
  }

  var payload = postGraphQL("searchComicAndAuthorQuery", { keyword: query }, runtime);
  var comics = ((((payload || {}).data || {}).searchComicsAndAuthors) || {}).comics || [];
  var items = [];

  for (var index = 0; index < comics.length; index += 1) {
    var item = normalizeComicItem(comics[index]);
    if (trim(item.title)) {
      items.push(item);
    }
  }

  return {
    query: query,
    page: page,
    hasMore: false,
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

  var payload = postGraphQL(
    "recentUpdate",
    {
      pagination: {
        limit: KOMIIC_PAGE_SIZE,
        offset: (page - 1) * KOMIIC_PAGE_SIZE,
        orderBy: "DATE_UPDATED",
        asc: true
      }
    },
    runtime
  );

  var comics = (((payload || {}).data || {}).recentUpdate) || [];
  var items = [];
  for (var index = 0; index < comics.length; index += 1) {
    var item = normalizeComicItem(comics[index]);
    if (trim(item.title)) {
      items.push(item);
    }
  }

  return {
    kind: kind,
    page: page,
    total: items.length,
    items: items
  };
}

function detail(itemID, runtime) {
  var comic = decodeItem(itemID);
  var comicId = trim(comic.id);
  if (!comicId) {
    throw new Error("missing komiic comic id");
  }

  var payload = postGraphQL("chapterByComicId", { comicId: comicId }, runtime);
  var rawChapters = (((payload || {}).data || {}).chaptersByComicId) || [];
  var chapters = [];

  for (var index = 0; index < rawChapters.length; index += 1) {
    var chapter = rawChapters[index] || {};
    var chapterId = trim(chapter.id);
    var serial = trim(chapter.serial);
    if (!chapterId) {
      continue;
    }

    var chapterType = trim(chapter.type);
    var chapterLabel = chapterType === "book" ? "卷" : "話";
    var chapterName = "第" + fallback(serial, String(index + 1)) + chapterLabel;
    if (trim(chapter.size)) {
      chapterName += "（" + trim(chapter.size) + "P）";
    }

    chapters.push({
      id: encodeItem({
        comicId: comicId,
        comicTitle: comic.title,
        chapterId: chapterId,
        chapterTitle: chapterName
      }),
      name: chapterName,
      url: chapterId,
      index: index,
      updatedLabel: chapterType === "book" ? "單行本" : "章節"
    });
  }

  return {
    item: {
      id: itemID,
      title: fallback(comic.title, "Komiic漫畫"),
      cover: trim(comic.cover),
      summary: buildDetailSummary(comic),
      author: fallback(comic.author, "Unknown author"),
      status: comic.status === "ONGOING" ? "連載中" : fallback(comic.status, "Unknown"),
      tags: toArray(comic.categories),
      detailUrl: detailUrl(comicId),
      chapters: chapters
    }
  };
}

function buildDetailSummary(comic) {
  var lines = [];
  if (comic.status) {
    lines.push("狀態：" + (comic.status === "ONGOING" ? "連載中" : comic.status));
  }
  if (comic.year) {
    lines.push("年份：" + comic.year);
  }
  if (comic.author) {
    lines.push("作者：" + comic.author);
  }
  if (comic.categories && comic.categories.length > 0) {
    lines.push("類型：" + comic.categories.join("，"));
  }
  if (comic.views) {
    lines.push("點閱：" + comic.views);
  }
  if (comic.update) {
    lines.push("最近更新：" + comic.update);
  }
  return lines.join("\n");
}

function images(chapterID, runtime) {
  var chapter = decodeItem(chapterID);
  var comicId = trim(chapter.comicId);
  var rawChapterId = trim(chapter.chapterId);
  if (!comicId || !rawChapterId) {
    throw new Error("invalid komiic chapter id");
  }

  var payload = postGraphQL("imagesByChapterId", { chapterId: rawChapterId }, runtime);
  var rawImages = (((payload || {}).data || {}).imagesByChapterId) || [];
  var urls = [];
  var entries = [];
  var referer = chapterReferer(comicId, rawChapterId);

  for (var index = 0; index < rawImages.length; index += 1) {
    var image = rawImages[index] || {};
    var kid = trim(image.kid);
    if (!kid) {
      continue;
    }

    var imageUrl = KOMIIC_BASE_URL + "/api/image/" + kid;
    urls.push(imageUrl);
    entries.push({
      url: imageUrl,
      referer: referer,
      headers: {
        Referer: referer
      }
    });
  }

  if (urls.length === 0) {
    throw new Error("komiic chapter returned no readable page urls");
  }

  return {
    comicTitle: fallback(chapter.comicTitle, "Komiic漫畫"),
    chapterTitle: fallback(chapter.chapterTitle, "Chapter"),
    chapterUrl: referer,
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
