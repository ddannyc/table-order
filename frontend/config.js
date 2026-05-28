/**
 * 全局配置
 *
 * 运行时通过 wx.getAccountInfoSync() 自动检测小程序版本，
 * 选择对应环境的后端地址。无需手动切换。
 *
 * develop — 开发版/预览（开发者工具），每人按需修改占位地址
 * trial   — 体验版，同正式版地址
 * release — 正式版
 */

const envConfig = {
  develop: {
    baseUrl: 'http://172.25.143.50:8080/api'
  },
  trial: {
    baseUrl: 'https://respectful-comfort-production-6bc5.up.railway.app/api'
  },
  release: {
    baseUrl: 'https://respectful-comfort-production-6bc5.up.railway.app/api'
  }
}

const accountInfo = wx.getAccountInfoSync()
const env = accountInfo.miniProgram.envVersion || 'release'
const API_BASE = (envConfig[env] && envConfig[env].baseUrl) || envConfig.release.baseUrl

module.exports = { API_BASE }