import { ApiError, errorDetail, notifyUnauthorized } from './client'
import { readToken } from '@/lib/token'

// XHR rather than ofetch: fetch has no upload progress event, and the progress
// bar is the whole point of the dialog.
export function uploadFile(
  bucket: string,
  key: string,
  file: File,
  onProgress?: (percent: number) => void,
): Promise<void> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest()
    xhr.open('POST', `/api/buckets/${encodeURIComponent(bucket)}/upload`)
    xhr.setRequestHeader('Authorization', `Bearer ${readToken()}`)

    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable && onProgress) {
        onProgress(Math.round((e.loaded / e.total) * 100))
      }
    }

    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve()
        return
      }
      if (xhr.status === 401) notifyUnauthorized()
      reject(new ApiError(xhr.status, errorDetail(xhr.responseText, xhr.statusText)))
    }

    xhr.onerror = () => reject(new ApiError(0, 'Network error'))

    const form = new FormData()
    form.append('file', file)
    form.append('key', key)
    xhr.send(form)
  })
}
