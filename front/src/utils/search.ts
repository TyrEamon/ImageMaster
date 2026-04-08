import * as OpenCC from 'opencc-js'

type TextConverter = (input: string) => string

const toSimplifiedConverters: TextConverter[] = [
  OpenCC.Converter({ from: 'tw', to: 'cn' }),
  OpenCC.Converter({ from: 'hk', to: 'cn' }),
  OpenCC.Converter({ from: 'jp', to: 'cn' }),
]

const extraVariantMap = new Map<string, string>([
  ['毎', '每'],
  ['気', '气'],
  ['処', '处'],
  ['隠', '隐'],
  ['絵', '绘'],
  ['綺', '绮'],
  ['髪', '发'],
  ['艶', '艳'],
  ['雙', '双'],
  ['樂', '乐'],
  ['櫻', '樱'],
  ['戀', '恋'],
  ['聲', '声'],
  ['氣', '气'],
  ['裡', '里'],
  ['麼', '么'],
  ['樣', '样'],
  ['這', '这'],
  ['為', '为'],
  ['與', '与'],
  ['後', '后'],
  ['臺', '台'],
])

const extraVariantPattern =
  extraVariantMap.size > 0
    ? new RegExp(`[${Array.from(extraVariantMap.keys()).join('')}]`, 'g')
    : null

function safeConvert(converter: TextConverter, input: string) {
  try {
    return converter(input)
  } catch {
    return input
  }
}

function applyExtraVariants(input: string) {
  if (!extraVariantPattern) {
    return input
  }

  return input.replace(extraVariantPattern, (char) => extraVariantMap.get(char) ?? char)
}

function normalizeSeparators(input: string) {
  return input
    .replace(/[\[\]【】()（）「」『』〔〕〈〉《》]/g, ' ')
    .replace(/[~!@#$%^&*+=:;"'`?,，。；：、|\\/]+/g, ' ')
    .replace(/[_-]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
}

function stripBracketSegments(input: string) {
  return input
    .replace(/\[[^\]]*\]/g, ' ')
    .replace(/【[^】]*】/g, ' ')
    .replace(/\([^)]*\)/g, ' ')
    .replace(/（[^）]*）/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
}

function stripConventionTokens(input: string) {
  return input
    .replace(/\bc\d{2,4}\b/gi, ' ')
    .replace(/\b(?:vol|episode|chapter|part|pixiv|fanbox|dlsite)\b/gi, ' ')
    .replace(/\d{5,}/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
}

function compactNormalizedText(input: string) {
  return normalizeSearchText(input).replace(/\s+/g, '')
}

export function normalizeSearchText(input: string) {
  if (!input) {
    return ''
  }

  let normalized = input.normalize('NFKC').toLowerCase()

  for (const converter of toSimplifiedConverters) {
    normalized = safeConvert(converter, normalized)
  }

  normalized = applyExtraVariants(normalized)

  return normalizeSeparators(normalized)
}

export function splitSearchKeywords(input: string) {
  const normalized = normalizeSearchText(input)
  return normalized ? normalized.split(' ') : []
}

export function buildMangaSearchIndex(name: string, path: string) {
  const normalizedPath = path.replace(/[\\/]/g, ' ')
  const strippedTitle = stripBracketSegments(name)
  const cleanedTitle = stripConventionTokens(strippedTitle)

  const variants = new Set<string>([
    normalizeSearchText(name),
    normalizeSearchText(normalizedPath),
    normalizeSearchText(`${name} ${normalizedPath}`),
    normalizeSearchText(strippedTitle),
    normalizeSearchText(cleanedTitle),
    compactNormalizedText(name),
    compactNormalizedText(strippedTitle),
    compactNormalizedText(cleanedTitle),
  ])

  return Array.from(variants)
    .filter(Boolean)
    .join(' ')
}
