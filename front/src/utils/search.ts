import * as OpenCC from "opencc-js/core";
import * as Locale from "opencc-js/preset";

const converters = [
    OpenCC.ConverterFactory(Locale.from.t, Locale.to.cn),
    OpenCC.ConverterFactory(Locale.from.tw, Locale.to.cn),
    OpenCC.ConverterFactory(Locale.from.hk, Locale.to.cn),
    OpenCC.ConverterFactory(Locale.from.jp, Locale.to.cn),
];

function normalizeBase(text: string) {
    return text.normalize("NFKC").toLowerCase().trim();
}

function collapseWhitespace(text: string) {
    return text.replace(/\s+/g, " ").trim();
}

function stripSymbols(text: string) {
    return text.replace(/[\s\p{P}\p{S}]+/gu, "");
}

function collectVariants(text: string) {
    const normalized = normalizeBase(text);
    if (!normalized) {
        return [];
    }

    const variants = new Set<string>();
    const candidates = [normalized, ...converters.map((converter) => converter(normalized))];

    for (const candidate of candidates) {
        const withSpaces = collapseWhitespace(candidate);
        const compact = stripSymbols(candidate);

        if (withSpaces) {
            variants.add(withSpaces);
        }

        if (compact) {
            variants.add(compact);
        }
    }

    return Array.from(variants);
}

export function buildSearchIndex(text: string) {
    return collectVariants(text).join(" ");
}

export function buildSearchKeywordGroups(query: string) {
    return collapseWhitespace(query)
        .split(" ")
        .filter(Boolean)
        .map((keyword) => collectVariants(keyword));
}
