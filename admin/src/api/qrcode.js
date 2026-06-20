import client from './client'

export const listQRCodes = (shopId) => client.get(`/merchant/shops/${shopId}/qrcodes`)

// Returns { id, shop_id, table_no, token, qr_image } — qr_image is a base64 data URL.
export const generateQRCode = (shopId, tableNo) =>
  client.post(`/merchant/shops/${shopId}/qrcodes`, { table_no: tableNo })
