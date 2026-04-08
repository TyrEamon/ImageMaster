export function debounce(fn: (...args: any[]) => void, delay: number) {
    let timer: number | null = null;
    return (...args: any[]) => {
        if (timer) {
            clearTimeout(timer);
        }
        timer = setTimeout(() => fn(...args), delay);
    };
}

export function UrlEncode(url: string) {
    return encodeURIComponent(url);
}

export function UrlDecode(url: string) {
    return decodeURIComponent(url);
}

export { buildSearchIndex, buildSearchKeywordGroups } from './search';
