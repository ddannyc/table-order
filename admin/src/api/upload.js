import client from './client'

// Uploads an image file to the backend (Cloudflare R2). Returns { url }.
export function uploadImage(file) {
  const fd = new FormData()
  fd.append('file', file)
  return client.post('/merchant/upload', fd)
}
