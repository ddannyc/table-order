/**
 * 全局配置
 *
 * API_BASE 通过微信小程序编译时常量 defineConstants 注入。
 * project.config.json  → 生产环境
 * project.private.config.json → 本地开发环境（覆盖 project.config.json）
 *
 * 无需修改此文件切换环境。
 */
const API_BASE = typeof API_BASE !== 'undefined' ? API_BASE : 'http://localhost:8080/api'

module.exports = { API_BASE }