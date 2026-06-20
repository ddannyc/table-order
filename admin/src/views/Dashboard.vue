<script setup>
import { ref, onMounted, watch } from 'vue'
import { useAuthStore } from '../stores/auth'
import { getDashboard, getStats } from '../api/stats'

const auth = useAuthStore()

const dashboard = ref({ shops: [], total_users: 0, total_orders: 0, total_revenue: 0 })
const stats = ref({ new_users: 0, orders: 0, revenue: 0, rewarded: 0 })
const date = ref(todayStr())
const loadingDash = ref(false)
const loadingStats = ref(false)

function todayStr() {
  const d = new Date()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${d.getFullYear()}-${m}-${day}`
}

async function loadDashboard() {
  loadingDash.value = true
  try {
    dashboard.value = await getDashboard()
  } catch {
    // handled by interceptor
  } finally {
    loadingDash.value = false
  }
}

async function loadStats() {
  if (!date.value) return
  loadingStats.value = true
  try {
    const params = { date: date.value }
    if (auth.currentShopId) params.shop_id = auth.currentShopId
    stats.value = await getStats(params)
  } catch {
    // handled by interceptor
  } finally {
    loadingStats.value = false
  }
}

onMounted(() => {
  loadDashboard()
  loadStats()
})
watch(date, loadStats)
watch(() => auth.currentShopId, loadStats)
</script>

<template>
  <div>
    <el-card v-loading="loadingDash" style="margin-bottom: 16px">
      <template #header>累计（全部店铺）</template>
      <el-row :gutter="16">
        <el-col :span="6">
          <el-statistic title="累计营业额" :value="dashboard.total_revenue" :precision="2" prefix="¥" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="累计订单数" :value="dashboard.total_orders" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="累计用户数" :value="dashboard.total_users" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="店铺数" :value="dashboard.shops?.length || 0" />
        </el-col>
      </el-row>
    </el-card>

    <el-card v-loading="loadingStats">
      <template #header>
        <div class="stats-header">
          <span>单日统计{{ auth.currentShopId ? '（当前店铺）' : '（全部店铺）' }}</span>
          <el-date-picker
            v-model="date"
            type="date"
            value-format="YYYY-MM-DD"
            :clearable="false"
            placeholder="选择日期"
          />
        </div>
      </template>
      <el-row :gutter="16">
        <el-col :span="6">
          <el-statistic title="营业额" :value="stats.revenue" :precision="2" prefix="¥" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="订单数" :value="stats.orders" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="新增用户" :value="stats.new_users" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="返利发放" :value="stats.rewarded" :precision="2" prefix="¥" />
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<style scoped>
.stats-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
</style>
