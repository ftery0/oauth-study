export function relativeKo(iso: string): string {
  const t = new Date(iso).getTime()
  if (Number.isNaN(t)) return ''
  const diff = Date.now() - t
  const sec = Math.round(diff / 1000)
  if (sec < 60) return '방금 전'
  const min = Math.round(sec / 60)
  if (min < 60) return `${min}분 전`
  const hr = Math.round(min / 60)
  if (hr < 24) return `${hr}시간 전`
  const day = Math.round(hr / 24)
  if (day < 7) return `${day}일 전`
  return new Date(iso).toLocaleDateString('ko-KR', { month: 'short', day: 'numeric' })
}

export function absoluteKo(iso: string): string {
  return new Date(iso).toLocaleString('ko-KR', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export const STATUS_LABEL_KO: Record<string, string> = {
  open: '응답 대기',
  pending: '처리 중',
  resolved: '해결됨',
  closed: '종료',
}

export const PRIORITY_LABEL_KO: Record<string, string> = {
  low: '낮음',
  normal: '보통',
  high: '높음',
  urgent: '긴급',
}
