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
  ['絵', '绘'],
  ['広', '广'],
  ['沢', '泽'],
  ['辺', '边'],
  ['関', '关'],
  ['黒', '黑'],
  ['桜', '樱'],
  ['薬', '药'],
  ['読', '读'],
  ['楽', '乐'],
  ['戦', '战'],
  ['観', '观'],
  ['雑', '杂'],
  ['転', '转'],
  ['伝', '传'],
  ['覚', '觉'],
  ['単', '单'],
  ['隠', '隐'],
  ['営', '营'],
  ['悪', '恶'],
  ['県', '县'],
  ['実', '实'],
  ['験', '验'],
  ['剤', '剂'],
  ['図', '图'],
  ['塩', '盐'],
  ['艶', '艳'],
  ['衛', '卫'],
  ['釣', '钓'],
  ['瀬', '濑'],
  ['亀', '龟'],
  ['円', '圆'],
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

export function normalizeSearchText(input: string) {
  if (!input) {
    return ''
  }

  let normalized = input.normalize('NFKC').toLowerCase()

  for (const converter of toSimplifiedConverters) {
    normalized = safeConvert(converter, normalized)
  }

  normalized = applyExtraVariants(normalized)

  return normalized
    .replace(/[\[\]【】()（）「」『』〈〉《》〔〕［］｛｝{}]/g, ' ')
    .replace(/[~!@#$%^&*+=:;"'`?,，。；：、|\\/]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
}

export function splitSearchKeywords(input: string) {
  const normalized = normalizeSearchText(input)
  return normalized ? normalized.split(' ') : []
}

export function buildMangaSearchIndex(name: string, path: string) {
  return normalizeSearchText(`${name} ${path}`)
}
