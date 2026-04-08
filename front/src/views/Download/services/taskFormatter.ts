
// 格式化时间
export function formatTime(timeStr: string): string {
  if (!timeStr) return '';
  const date = new Date(timeStr);
  return `${date.toLocaleDateString()} ${date.toLocaleTimeString()}`;
}

// 计算两个时间差，格式化为秒
export function calculateTimeDifference(startTime: string, endTime: string): string {
  const startTimeDate = new Date(startTime);
  const endTimeDate = new Date(endTime);
  const timeDifference = endTimeDate.getTime() - startTimeDate.getTime();
  const seconds = Math.floor(timeDifference / 1000);
  return `${seconds}秒`;
}

// 格式化进度信息
export function formatProgress(current: number, total: number): string {
  if (total <= 0) return '准备下载中...';
  const percentage = Math.round((current / total) * 100);
  return `${current}/${total} 张图片 (${percentage}%)`;
}

// 计算进度百分比
export function calculateProgressPercentage(current: number, total: number): number {
  if (total <= 0) return 0;
  return Math.round((current / total) * 100);
}
