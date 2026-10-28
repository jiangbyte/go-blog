/** Author: go-blog */

import { authApi } from '@/api'
import { encryptFieldsWithPublicKey } from '@/utils/rsaEncrypt'

type PublicKeyPayload = {
  public_key?: string
  publicKey?: string
}

function readPublicKeyPayload(response: unknown): PublicKeyPayload {
  const root = response as { data?: PublicKeyPayload } | PublicKeyPayload
  const payload = root && typeof root === 'object' && 'data' in root ? root.data : root
  return (payload ?? {}) as PublicKeyPayload
}

/** 拉取公钥并用 RSA-OAEP 加密密码字段（mypage 无 key_id）。 */
export async function encryptPasswords<T extends Record<string, string | null | undefined>>(
  fields: T,
) {
  const response = await authApi.publicKey()
  const payload = readPublicKeyPayload(response)
  const publicKey = payload.public_key ?? payload.publicKey
  if (!publicKey) {
    throw new Error('获取密码加密公钥失败')
  }
  return {
    values: await encryptFieldsWithPublicKey(publicKey, fields),
  }
}
