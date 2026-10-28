/** Author: go-blog — 管理端简历 PDF 下载（调后端导出，真实文字层） */

import { isAxiosError } from 'axios'
import { resumeApi } from '@/api'
import { getFilenameFromContentDisposition, saveBlob } from '@/utils'

/**
 * 请求后端生成并下载简历 PDF。
 * 体验与普通文件下载一致；PDF 内文字可选中/检索。
 */
export async function downloadResumeDocument(
  resumeId: string,
  fallbackName = '简历',
): Promise<void> {
  const id = String(resumeId || '').trim()
  if (!id) throw new Error('缺少简历 ID')

  try {
    const response = await resumeApi.exportPdf(id)
    const disposition = response.headers?.['content-disposition']
    const dispositionText = Array.isArray(disposition)
      ? disposition[0]
      : typeof disposition === 'string'
        ? disposition
        : undefined
    const filename =
      getFilenameFromContentDisposition(dispositionText) ||
      `${
        String(fallbackName || '简历')
          .replace(/[/\\?%*:|"<>']/g, '_')
          .trim() || '简历'
      }.pdf`

    const data = response.data
    if (data instanceof Blob && data.type && data.type.includes('application/json')) {
      const text = await data.text()
      let msg = '下载失败'
      try {
        const json = JSON.parse(text)
        msg = String(json?.message || json?.msg || msg)
      } catch {
        /* ignore */
      }
      throw new Error(msg)
    }

    saveBlob(data, filename.endsWith('.pdf') ? filename : `${filename}.pdf`)
  } catch (err) {
    if (isAxiosError(err)) {
      const raw = err.response?.data
      if (raw instanceof Blob) {
        try {
          const json = JSON.parse(await raw.text())
          throw new Error(String(json?.message || json?.msg || '下载失败'))
        } catch (inner) {
          if (inner instanceof Error && inner.message !== '下载失败') throw inner
        }
      }
      throw new Error(err.message || '下载失败')
    }
    throw err
  }
}
