/** Author: go-blog — IPv4 子网划分计算（纯前端） */

export type SubnetResult = {
  input: string
  ip: string
  prefix: number
  netmask: string
  wildcard: string
  network: string
  broadcast: string
  firstHost: string
  lastHost: string
  totalHosts: number
  usableHosts: number
  hostBits: number
}

function parseOctets(ip: string): number[] | null {
  const parts = ip.trim().split('.')
  if (parts.length !== 4) return null
  const octets: number[] = []
  for (const p of parts) {
    if (!/^\d{1,3}$/.test(p)) return null
    const n = Number(p)
    if (n < 0 || n > 255) return null
    octets.push(n)
  }
  return octets
}

export function ipToInt(ip: string): number | null {
  const octets = parseOctets(ip)
  if (!octets) return null
  return ((octets[0]! << 24) >>> 0) + (octets[1]! << 16) + (octets[2]! << 8) + octets[3]!
}

export function intToIp(n: number): string {
  const v = n >>> 0
  return [(v >>> 24) & 255, (v >>> 16) & 255, (v >>> 8) & 255, v & 255].join('.')
}

/** 将点分掩码转为前缀长度；非法返回 null。 */
export function maskToPrefix(mask: string): number | null {
  const n = ipToInt(mask)
  if (n === null) return null
  // 合法掩码：高位连续 1，低位连续 0
  const inv = (~n) >>> 0
  if ((inv & (inv + 1)) !== 0) return null
  let prefix = 0
  let x = n >>> 0
  while (x & 0x80000000) {
    prefix++
    x = (x << 1) >>> 0
  }
  if (x !== 0) return null
  return prefix
}

/**
 * 解析 CIDR（如 192.168.1.10/24）或 IP + 掩码/前缀。
 * maskOrPrefix 可为 "24"、"/24"、"255.255.255.0"。
 */
export function parseSubnetInput(cidrOrIp: string, maskOrPrefix?: string): { ip: string; prefix: number } | { error: string } {
  const raw = String(cidrOrIp || '').trim()
  if (!raw) return { error: '请输入 IP 或 CIDR' }

  let ip = raw
  let prefix: number | null = null

  if (raw.includes('/')) {
    const [left, right] = raw.split('/')
    ip = String(left || '').trim()
    const p = Number(String(right || '').trim())
    if (!Number.isInteger(p) || p < 0 || p > 32) return { error: 'CIDR 前缀须为 0–32 的整数' }
    prefix = p
  } else if (maskOrPrefix != null && String(maskOrPrefix).trim() !== '') {
    const m = String(maskOrPrefix).trim().replace(/^\//, '')
    if (/^\d{1,2}$/.test(m)) {
      const p = Number(m)
      if (p < 0 || p > 32) return { error: '前缀须为 0–32' }
      prefix = p
    } else {
      prefix = maskToPrefix(m)
      if (prefix === null) return { error: '子网掩码无效' }
    }
  } else {
    return { error: '请提供 CIDR（如 192.168.1.0/24）或掩码/前缀' }
  }

  if (ipToInt(ip) === null) return { error: 'IP 地址无效' }
  return { ip, prefix }
}

/** 按 IP + 前缀计算子网信息。 */
export function calculateSubnet(ip: string, prefix: number): SubnetResult | { error: string } {
  if (!Number.isInteger(prefix) || prefix < 0 || prefix > 32) {
    return { error: '前缀须为 0–32' }
  }
  const ipInt = ipToInt(ip)
  if (ipInt === null) return { error: 'IP 地址无效' }

  const hostBits = 32 - prefix
  const maskInt = prefix === 0 ? 0 : ((0xffffffff << hostBits) >>> 0)
  const wildcardInt = (~maskInt) >>> 0
  const networkInt = (ipInt & maskInt) >>> 0
  const broadcastInt = (networkInt | wildcardInt) >>> 0
  const totalHosts = 2 ** hostBits

  let firstHost = networkInt
  let lastHost = broadcastInt
  let usableHosts = totalHosts
  if (prefix < 31) {
    firstHost = (networkInt + 1) >>> 0
    lastHost = (broadcastInt - 1) >>> 0
    usableHosts = Math.max(totalHosts - 2, 0)
  } else if (prefix === 31) {
    // RFC 3021 点对点：两地址均可用
    firstHost = networkInt
    lastHost = broadcastInt
    usableHosts = 2
  } else {
    // /32 单主机
    firstHost = networkInt
    lastHost = networkInt
    usableHosts = 1
  }

  return {
    input: `${ip}/${prefix}`,
    ip,
    prefix,
    netmask: intToIp(maskInt),
    wildcard: intToIp(wildcardInt),
    network: intToIp(networkInt),
    broadcast: intToIp(broadcastInt),
    firstHost: intToIp(firstHost),
    lastHost: intToIp(lastHost),
    totalHosts,
    usableHosts,
    hostBits,
  }
}
