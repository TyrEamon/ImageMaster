import { StartCrawl } from "../../../../wailsjs/go/api/CrawlerAPI";
 import { toast  } from 'vue-sonner'

/**
 * 下载结果接口
 */
export interface DownloadResult {
  success: boolean;
  taskId?: string;
  error?: string;
}

/**
 * 下载选项接口
 */
export interface DownloadOptions {
  url: string;
  onStart?: (taskId: string) => void;
  onError?: (error: string) => void;
  onSuccess?: (taskId: string) => void;
}

/**
 * 验证URL格式
 * @param url 要验证的URL
 * @returns 是否为有效URL
 */
export function validateUrl(url: string): boolean {
  if (!url || !url.trim()) {
    return false;
  }
  
  try {
    new URL(url.trim());
    return true;
  } catch {
    return false;
  }
}

/**
 * 执行下载任务
 * @param options 下载选项
 * @returns Promise<DownloadResult>
 */
export async function executeDownload(options: DownloadOptions): Promise<DownloadResult> {
  const { url, onStart, onError, onSuccess } = options;
  
  // 验证URL
  if (!validateUrl(url)) {
    const error = '请输入有效的网址';
    onError?.(error);
    return { success: false, error };
  }
  
  try {
    // 调用下载开始回调
    onStart?.(url.trim());
    
    // 执行爬取
    const taskId = await StartCrawl(url.trim());
    
    if (taskId) {
      // 下载任务创建成功
      onSuccess?.(taskId);
      return { success: true, taskId };
    } else {
      const error = '下载失败，请检查网址是否正确';
      onError?.(error);
      return { success: false, error };
    }
  } catch (err: any) {
    const error = `下载出错: ${err.message || '未知错误'}`;
    onError?.(error);
    return { success: false, error };
  }
}

/**
 * 创建下载处理器
 * @param callbacks 回调函数集合
 * @returns 下载处理函数
 */
export function createDownloadHandler(callbacks: {
  onStart?: () => void;
  onSuccess?: (taskId: string, url: string) => void;
  onError?: (error: string) => void;
  onFinally?: () => void;
}) {
  return async (url: string): Promise<DownloadResult> => {
    const { onStart, onSuccess, onError, onFinally } = callbacks;
    
    try {
      onStart?.();
      toast.info('下载任务已添加到队列')
      const result = await executeDownload({
        url,
        onError,
        onSuccess: (taskId) => onSuccess?.(taskId, url)
      });
      
      return result;
    } finally {
      onFinally?.();
    }
  };
}

/**
 * 格式化错误消息
 * @param error 错误对象或字符串
 * @returns 格式化后的错误消息
 */
export function formatDownloadError(error: any): string {
  if (typeof error === 'string') {
    return error;
  }
  
  if (error?.message) {
    return `下载出错: ${error.message}`;
  }
  
  return '下载出错: 未知错误';
}