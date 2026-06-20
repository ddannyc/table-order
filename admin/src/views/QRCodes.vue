<script setup>
import { ref, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'
import { listQRCodes, generateQRCode } from '../api/qrcode'

const auth = useAuthStore()

const codes = ref([])
const loading = ref(false)
const tableNo = ref('')
const generating = ref(false)

// Image returned by the most recent generation (base64 data URL).
const lastImage = ref('')
const lastTableNo = ref('')
const dialogVisible = ref(false)

async function load() {
  if (!auth.currentShopId) {
    codes.value = []
    return
  }
  loading.value = true
  try {
    codes.value = await listQRCodes(auth.currentShopId)
  } catch {
    // handled by interceptor
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch(() => auth.currentShopId, load)

async function generate() {
  if (!tableNo.value.trim()) {
    ElMessage.warning('请输入桌号')
    return
  }
  generating.value = true
  try {
    const res = await generateQRCode(auth.currentShopId, tableNo.value.trim())
    lastImage.value = res.qr_image
    lastTableNo.value = res.table_no
    dialogVisible.value = true
    tableNo.value = ''
    await load()
  } catch {
    // handled by interceptor
  } finally {
    generating.value = false
  }
}

function download() {
  if (!lastImage.value) return
  const a = document.createElement('a')
  a.href = lastImage.value
  a.download = `桌码_${lastTableNo.value}.png`
  a.click()
}
</script>

<template>
  <el-card>
    <template #header>
      <div class="header">
        <span>桌台二维码</span>
        <div v-if="auth.currentShopId" class="gen">
          <el-input
            v-model="tableNo"
            placeholder="桌号，如 A01"
            style="width: 160px"
            @keyup.enter="generate"
          />
          <el-button type="primary" :loading="generating" @click="generate">生成桌码</el-button>
        </div>
      </div>
    </template>

    <el-alert
      v-if="!auth.currentShopId"
      type="info"
      :closable="false"
      title="请先在顶部选择店铺。"
    />

    <el-table v-else v-loading="loading" :data="codes" stripe>
      <el-table-column prop="table_no" label="桌号" min-width="120" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'">
            {{ row.status === 1 ? '启用' : '停用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="token" label="Token" min-width="180" show-overflow-tooltip />
      <el-table-column label="创建时间" width="200">
        <template #default="{ row }">{{ new Date(row.created_at).toLocaleString() }}</template>
      </el-table-column>
      <template #empty>暂无桌码，输入桌号后点「生成桌码」</template>
    </el-table>

    <el-dialog v-model="dialogVisible" title="桌码已生成" width="320px" align-center>
      <div class="qr-box">
        <el-image v-if="lastImage" :src="lastImage" style="width: 256px; height: 256px" />
        <p class="qr-table">桌号：{{ lastTableNo }}</p>
      </div>
      <template #footer>
        <el-button @click="dialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="download">下载图片</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<style scoped>
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.gen {
  display: flex;
  gap: 8px;
}
.qr-box {
  text-align: center;
}
.qr-table {
  margin: 12px 0 0;
  color: #606266;
}
</style>
