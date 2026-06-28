<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { getShops } from '../api/shop'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const shops = ref([])
const activeMenu = computed(() => route.name)

onMounted(async () => {
  try {
    shops.value = await getShops()
    if (shops.value.length && !auth.currentShopId) {
      auth.setShop(shops.value[0].id)
    }
  } catch {
    // handled by interceptor
  }
})

function onShopChange(id) {
  auth.setShop(id)
}

function logout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <el-container class="layout">
    <el-aside width="200px" class="aside">
      <div class="brand">商家后台</div>
      <el-menu :default-active="activeMenu" router>
        <el-menu-item index="dashboard" :route="{ name: 'dashboard' }">
          <el-icon><DataLine /></el-icon><span>数据看板</span>
        </el-menu-item>
        <el-menu-item index="products" :route="{ name: 'products' }">
          <el-icon><Dish /></el-icon><span>菜品管理</span>
        </el-menu-item>
        <el-menu-item index="orders" :route="{ name: 'orders' }">
          <el-icon><List /></el-icon><span>订单</span>
        </el-menu-item>
        <el-menu-item index="shop" :route="{ name: 'shop' }">
          <el-icon><Shop /></el-icon><span>店铺与返利</span>
        </el-menu-item>
        <el-menu-item index="qrcodes" :route="{ name: 'qrcodes' }">
          <el-icon><Grid /></el-icon><span>桌台二维码</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="header">
        <el-select
          v-if="shops.length"
          :model-value="auth.currentShopId"
          placeholder="选择店铺"
          class="shop-select"
          @update:model-value="onShopChange"
        >
          <el-option v-for="s in shops" :key="s.id" :label="s.name" :value="s.id" />
        </el-select>
        <span v-else class="no-shop">暂无店铺，请先到「店铺与返利」创建</span>

        <div class="header-right">
          <span class="merchant-name">{{ auth.merchant?.name || auth.merchant?.phone }}</span>
          <el-button text @click="logout">退出</el-button>
        </div>
      </el-header>

      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.layout {
  height: 100vh;
}
.aside {
  background: #fff;
  border-right: 1px solid #e6e6e6;
}
.brand {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 16px;
  border-bottom: 1px solid #e6e6e6;
}
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  border-bottom: 1px solid #e6e6e6;
}
.shop-select {
  width: 200px;
}
.no-shop {
  color: #999;
  font-size: 13px;
}
.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}
.merchant-name {
  color: #666;
}
</style>
