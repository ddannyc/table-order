/**
 * 全局配置
 * 修改 API_BASE 值切换开发/生产环境
 *
 * 使用方式（需在微信开发者工具中重新编译）:
 *   dev:   API_BASE = 'http://10.157.2.132:8080/api'
 *   prod:  API_BASE = 'https://api.yourdomain.com/api'
 *
 * 切换方法: 修改下面一行 const 值，然后重新编译项目
 */
const API_BASE = 'http://localhost:8080/api'

module.exports = { API_BASE }