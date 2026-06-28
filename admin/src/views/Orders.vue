<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { getMerchantOrders } from '../api/order'
import { useAuthStore } from '../stores/auth'
import {
  orderStatusLabel,
  shansongStatusLabel,
  isPaid,
  isPrepared,
  needsAction,
} from '../utils/orderBoard'

const auth = useAuthStore()

const loading = ref(false)
const orders = ref([])
const total = ref(0)
const revenue = ref(0)
const rewarded = ref(0)

const tab = ref('pending')
const date = ref('')
const type = ref('')

// MVP: tab buckets (待处理 etc.) are client-side predicates, so we fetch the
// shop's recent orders in one capped page and partition locally.
const PAGE_CAP = 100

async function load() {
  if (!auth.currentShopId) {
    orders.value = []
    total.value = 0
    return
  }
  loading.value = true
  try {
    const params = { shop_id: auth.currentShopId, page_size: PAGE_CAP }
    if (date.value) params.date = date.value
    if (type.value) params.type = type.value
    const res = await getMerchantOrders(params)
    orders.value = res.orders || []
    total.value = res.total || 0
    revenue.value = res.revenue || 0
    rewarded.value = res.rewarded || 0
  } catch {
    // surfaced by axios interceptor
  } finally {
    loading.value = false
  }
}

function inBucket(o, which) {
  switch (which) {
    case 'pending':
      return needsAction(o)
    case 'active':
      return o.status === 2 && !needsAction(o)
    case 'done':
      return o.status === 3 || o.status === 4
    default:
      return true
  }
}

const filtered = computed(() => orders.value.filter((o) => inBucket(o, tab.value)))
const pendingCount = computed(() => orders.value.filter((o) => inBucket(o, 'pending')).length)
const capped = computed(() => total.value > orders.value.length)

function typeLabel(o) {
  return o.order_type === 'delivery' ? '外卖' : '堂食'
}
function fmtTime(s) {
  return s ? s.replace('T', ' ').slice(0, 16) : ''
}

onMounted(load)
watch([() => auth.currentShopId, date, type], load)
</script>

<template>
  <div class="orders">
    <div class="toolbar">
      <el-radio-group v-model="tab">
        <el-radio-button value="pending">待处理 ({{ pendingCount }})</el-radio-button>
        <el-radio-button value="active">进行中</el-radio-button>
        <el-radio-button value="done">已完成</el-radio-button>
        <el-radio-button value="all">全部</el-radio-button>
      </el-radio-group>

      <div class="filters">
        <el-select v-model="type" placeholder="类型" clearable class="type-select">
          <el-option label="堂食" value="dine_in" />
          <el-option label="外卖" value="delivery" />
        </el-select>
        <el-date-picker v-model="date" type="date" value-format="YYYY-MM-DD" placeholder="按日期" clearable />
      </div>
    </div>

    <div class="summary">
      共 {{ total }} 单 · 营业额 ¥{{ revenue.toFixed(2) }} · 返利 ¥{{ rewarded.toFixed(2) }}
      <span v-if="capped" class="cap-hint">（仅显示最近 {{ orders.length }} 条，请用日期/类型筛选）</span>
    </div>

    <el-table v-loading="loading" :data="filtered" empty-text="暂无订单">
      <el-table-column prop="order_no" label="订单号" min-width="160" />
      <el-table-column label="类型" width="80">
        <template #default="{ row }">{{ typeLabel(row) }}</template>
      </el-table-column>
      <el-table-column label="金额" width="100">
        <template #default="{ row }">¥{{ Number(row.amount).toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="支付" width="90">
        <template #default="{ row }">
          <el-tag :type="isPaid(row) ? 'success' : 'info'" size="small">
            {{ orderStatusLabel(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="履约 / 配送" min-width="140">
        <template #default="{ row }">
          <template v-if="row.order_type === 'delivery'">
            <el-tag
              :type="row.delivery && row.delivery.shansong_status === -1 ? 'danger' : 'warning'"
              size="small"
            >
              {{ shansongStatusLabel(row.delivery ? row.delivery.shansong_status : 0) }}
            </el-tag>
          </template>
          <template v-else-if="isPaid(row)">
            <el-tag :type="isPrepared(row) ? 'success' : 'warning'" size="small">
              {{ isPrepared(row) ? '已出餐' : '待出餐' }}
            </el-tag>
          </template>
          <span v-else>—</span>
        </template>
      </el-table-column>
      <el-table-column label="下单时间" width="150">
        <template #default="{ row }">{{ fmtTime(row.created_at) }}</template>
      </el-table-column>
    </el-table>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  gap: 12px;
  flex-wrap: wrap;
}
.filters {
  display: flex;
  gap: 8px;
}
.type-select {
  width: 110px;
}
.summary {
  color: #666;
  font-size: 13px;
  margin-bottom: 12px;
}
.cap-hint {
  color: #e6a23c;
}
</style>
