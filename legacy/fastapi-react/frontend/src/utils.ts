export function formatNumber(value: number | null | undefined): string {
  const n = Number(value || 0);
  if (Math.abs(n) >= 100000000) return `${(n / 100000000).toFixed(2)}亿`;
  if (Math.abs(n) >= 10000) return `${(n / 10000).toFixed(1)}万`;
  if (Math.abs(n) >= 1000) return n.toLocaleString("zh-CN", { maximumFractionDigits: 0 });
  return n.toLocaleString("zh-CN", { maximumFractionDigits: 2 });
}

export function percent(value: number | null | undefined): string {
  if (value === null || value === undefined) return "对手为 0";
  return `${(value * 100).toFixed(1)}%`;
}

export function dateTime(value?: string | null): string {
  if (!value) return "-";
  return new Date(value).toLocaleString("zh-CN", { hour12: false });
}

export function shortDate(value?: string | null): string {
  if (!value) return "-";
  return new Date(value).toLocaleDateString("zh-CN");
}
