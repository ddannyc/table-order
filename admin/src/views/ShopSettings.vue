<script setup>
import { ref, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'
import { getShops, createShop, updateShop } from '../api/shop'
import { toPercent, toDecimal, parseCategories } from '../utils/reward'

const auth = useAuthStore()

const loading = ref(false)
const saving = ref(false)
const shops = ref([])
const hasShop = ref(true)

// Editable form. Reward rates are held as percentages for the UI.
const form = ref(blankForm())

function blankForm() {
  return {
    name: '',
    description: '',
    address: '',
    phone: '',
    hours: '',
    latitude: 0,
    longitude: 0,
    rewardSelf: 3,
    rewardL1: 10,
    rewardL2: 4,
    ceiling: 50,
    excludeCategories: [],
  }
}

function fillForm(shop) {
  form.value = {
    name: shop.name || '',
    description: shop.description || '',
    address: shop.address || '',
    phone: shop.phone || '',
    hours: shop.hours || '',
    latitude: shop.latitude || 0,
    longitude: shop.longitude || 0,
    rewardSelf: toPercent(shop.reward_rate_self),
    rewardL1: toPercent(shop.reward_rate_level1),
    rewardL2: toPercent(shop.reward_rate_level2),
    ceiling: toPercent(shop.reward_ceiling),
    excludeCategories: parseCategories(shop.reward_exclude_categories),
  }
}

async function load() {
  loading.value = true
  try {
    shops.value = await getShops()
    if (!shops.value.length) {
      hasShop.value = false
      form.value = blankForm()
      return
    }
    hasShop.value = true
    if (!auth.currentShopId) auth.setShop(shops.value[0].id)
    const current = shops.value.find((s) => s.id === auth.currentShopId) || shops.value[0]
    auth.setShop(current.id)
    fillForm(current)
  } catch {
    // handled by interceptor
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch(() => auth.currentShopId, load)

async function save() {
  if (!form.value.name) {
    ElMessage.warning('请输入店铺名称')
    return
  }
  saving.value = true
  try {
    if (!hasShop.value) {
      const shop = await createShop({
        name: form.value.name,
        description: form.value.description,
        address: form.value.address,
        phone: form.value.phone,
        hours: form.value.hours,
      })
      auth.setShop(shop.id)
      ElMessage.success('店铺已创建，请刷新页面以在顶部切换器中看到它')
      hasShop.value = true
    } else {
      await updateShop(auth.currentShopId, {
        name: form.value.name,
        description: form.value.description,
        address: form.value.address,
        phone: form.value.phone,
        hours: form.value.hours,
        latitude: form.value.latitude,
        longitude: form.value.longitude,
        reward_rate_self: toDecimal(form.value.rewardSelf),
        reward_rate_level1: toDecimal(form.value.rewardL1),
        reward_rate_level2: toDecimal(form.value.rewardL2),
        reward_ceiling: toDecimal(form.value.ceiling),
        reward_exclude_categories: JSON.stringify(form.value.excludeCategories),
      })
      ElMessage.success('已保存')
    }
    await load()
  } catch {
    // handled by interceptor
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <el-card v-loading="loading">
    <template #header>{{ hasShop ? '店铺与返利' : '创建店铺' }}</template>

    <el-alert
      v-if="!hasShop"
      type="info"
      :closable="false"
      title="你还没有店铺，先填写基本信息创建。创建后可再配置返利比例。"
      style="margin-bottom: 16px"
    />

    <el-form label-width="120px" style="max-width: 560px">
      <el-divider content-position="left">店铺信息</el-divider>
      <el-form-item label="店铺名称" required>
        <el-input v-model="form.name" />
      </el-form-item>
      <el-form-item label="简介">
        <el-input v-model="form.description" type="textarea" :rows="2" />
      </el-form-item>
      <el-form-item label="地址">
        <el-input v-model="form.address" />
      </el-form-item>
      <el-form-item label="电话">
        <el-input v-model="form.phone" />
      </el-form-item>
      <el-form-item label="营业时间">
        <el-input v-model="form.hours" placeholder="如：10:00-22:00" />
      </el-form-item>
      <el-form-item label="门店坐标">
        <el-input-number v-model="form.latitude" :precision="6" :step="0.0001" :controls="false" placeholder="纬度" />
        <span class="unit">纬度</span>
        <el-input-number v-model="form.longitude" :precision="6" :step="0.0001" :controls="false" placeholder="经度" style="margin-left: 12px" />
        <span class="unit">经度</span>
      </el-form-item>

      <template v-if="hasShop">
        <el-divider content-position="left">返利配置（百分比）</el-divider>
        <el-form-item label="自购返利">
          <el-input-number v-model="form.rewardSelf" :min="0" :max="100" :precision="2" :step="0.5" />
          <span class="unit">%</span>
        </el-form-item>
        <el-form-item label="直推返利">
          <el-input-number v-model="form.rewardL1" :min="0" :max="100" :precision="2" :step="0.5" />
          <span class="unit">%</span>
        </el-form-item>
        <el-form-item label="间推返利">
          <el-input-number v-model="form.rewardL2" :min="0" :max="100" :precision="2" :step="0.5" />
          <span class="unit">%</span>
        </el-form-item>
        <el-form-item label="金币抵扣上限">
          <el-input-number v-model="form.ceiling" :min="0" :max="100" :precision="2" :step="1" />
          <span class="unit">%</span>
        </el-form-item>
        <el-form-item label="不参与返利分类">
          <el-select
            v-model="form.excludeCategories"
            multiple
            filterable
            allow-create
            default-first-option
            placeholder="输入分类名后回车，可多个"
            style="width: 100%"
          />
        </el-form-item>
      </template>

      <el-form-item>
        <el-button type="primary" :loading="saving" @click="save">
          {{ hasShop ? '保存' : '创建店铺' }}
        </el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>

<style scoped>
.unit {
  margin-left: 8px;
  color: #909399;
}
</style>
