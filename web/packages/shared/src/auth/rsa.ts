/** RSA-OAEP(SHA-256) 密码传输加密；优先 Web Crypto，否则 node-forge */
import * as forgeModule from 'node-forge'

type ForgeModule = typeof forgeModule

function getForge(): ForgeModule {
  const mod = forgeModule as ForgeModule & { default?: ForgeModule }
  return mod.default ?? mod
}

function normalizePublicKeyBase64(publicKeyB64OrPem: string) {
  return publicKeyB64OrPem
    .replace('-----BEGIN PUBLIC KEY-----', '')
    .replace('-----END PUBLIC KEY-----', '')
    .replace(/\s/g, '')
}

function importSpkiPublicKeyForge(publicKeyB64OrPem: string) {
  const forge = getForge()
  const der = forge.util.decode64(normalizePublicKeyBase64(publicKeyB64OrPem))
  const asn1 = forge.asn1.fromDer(der)
  return forge.pki.publicKeyFromAsn1(asn1)
}

function encryptTextRsaOaepForge(
  publicKey: ReturnType<typeof importSpkiPublicKeyForge>,
  value: string,
) {
  const forge = getForge()
  const encrypted = publicKey.encrypt(forge.util.encodeUtf8(value), 'RSA-OAEP', {
    md: forge.md.sha256.create(),
    mgf1: { md: forge.md.sha256.create() },
  })
  return forge.util.encode64(encrypted)
}

async function importSpkiPublicKeyWebCrypto(publicKeyB64OrPem: string) {
  const binary = globalThis.atob(normalizePublicKeyBase64(publicKeyB64OrPem))
  const bytes = new Uint8Array(binary.length)
  for (let index = 0; index < binary.length; index += 1) {
    bytes[index] = binary.charCodeAt(index)
  }
  return globalThis.crypto.subtle.importKey(
    'spki',
    bytes.buffer,
    { name: 'RSA-OAEP', hash: 'SHA-256' },
    false,
    ['encrypt'],
  )
}

function arrayBufferToBase64(buffer: ArrayBuffer) {
  const bytes = new Uint8Array(buffer)
  let binary = ''
  for (let index = 0; index < bytes.byteLength; index += 1) {
    binary += String.fromCharCode(bytes[index])
  }
  return globalThis.btoa(binary)
}

async function encryptTextRsaOaepWebCrypto(publicKey: CryptoKey, value: string) {
  const encrypted = await globalThis.crypto.subtle.encrypt(
    { name: 'RSA-OAEP' },
    publicKey,
    new TextEncoder().encode(value),
  )
  return arrayBufferToBase64(encrypted)
}

function canUseWebCrypto() {
  return globalThis.isSecureContext && Boolean(globalThis.crypto?.subtle)
}

export async function encryptPassword(publicKeyPem: string, password: string) {
  if (!publicKeyPem?.trim()) {
    throw new Error('密码加密公钥无效')
  }
  if (canUseWebCrypto()) {
    try {
      const publicKey = await importSpkiPublicKeyWebCrypto(publicKeyPem)
      return await encryptTextRsaOaepWebCrypto(publicKey, password)
    } catch {
      // HTTP 或偶发失败时回退 forge
    }
  }
  return encryptTextRsaOaepForge(importSpkiPublicKeyForge(publicKeyPem), password)
}
