export function mapNames(items?: Array<{ id?: string; name?: string }>) {
  return (items ?? [])
    .map((item) => item.name)
    .filter(Boolean)
    .join(' / ')
}

export function displayValue(value: unknown) {
  const text = String(value ?? '').trim()
  return text || '未设置'
}

const GENDER_LABEL: Record<string, string> = {
  UNKNOWN: '未知',
  MALE: '男',
  FEMALE: '女',
}

export function genderLabel(value: unknown) {
  const key = String(value ?? '').trim()
  if (!key) return '未知'
  return GENDER_LABEL[key] || key
}
