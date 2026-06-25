export function compactNumber(value: number | undefined | null): string {
  const n = Number(value || 0)
  if (Math.abs(n) >= 100000000) return `${(n / 100000000).toFixed(2)}亿`
  if (Math.abs(n) >= 10000) return `${(n / 10000).toFixed(1)}万`
  return String(Math.round(n))
}

export function percent(value: number | undefined | null): string {
  return `${((Number(value || 0)) * 100).toFixed(1)}%`
}

export function score(value: number | undefined | null): string {
  return Number(value || 0).toFixed(2)
}

export function dateTime(value: string | undefined | null): string {
  if (!value) return '-'
  return new Date(value).toLocaleString('zh-CN', { hour12: false })
}

export function avatarSeed(name: string) {
  const code = Array.from(name || '玩家').reduce((sum, ch) => sum + ch.charCodeAt(0), 0)
  return `/assets/icons/png/64/user.png?seed=${code % 10}`
}
