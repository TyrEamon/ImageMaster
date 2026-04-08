export function debounce(func: Function, wait: number) {
    let timeout: number | undefined = undefined;
    return function (...args: any[]) {
        clearTimeout(timeout);
        timeout = setTimeout(() => func.apply(null, args), wait);
    };
}

export const throttle = (func: Function, wait: number) => {
    let lastTime = 0;
    return (...args: any[]) => {
        const now = Date.now();
        if (now - lastTime >= wait) {
            func.apply(this, args);
            lastTime = now;
        }
    };
}