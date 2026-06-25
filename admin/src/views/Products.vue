<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '../stores/auth'
import {
  getMerchantProducts,
  createProduct,
  updateProduct,
  deleteProduct,
  createProductSpec,
  updateProductSpec,
  deleteProductSpec,
} from '../api/product'
import { uploadImage } from '../api/upload'

const auth = useAuthStore()

const STATUS = {
  1: { label: '上架', type: 'success' },
  0: { label: '下架', type: 'info' },
  2: { label: '售罄', type: 'warning' },
}

const allProducts = ref([])
const loading = ref(false)

// Products belonging to the currently selected shop.
const products = computed(() =>
  allProducts.value.filter((p) => p.shop_id === auth.currentShopId),
)

async function load() {
  loading.value = true
  try {
    allProducts.value = await getMerchantProducts()
  } catch {
    // handled by interceptor
  } finally {
    loading.value = false
  }
}

onMounted(load)

// ---- Add / edit dialog ----
const dialogVisible = ref(false)
const saving = ref(false)
const editingId = ref(null)
const form = ref({ name: '', price: 0, category: '', description: '', image: '', status: 1 })

const dialogTitle = computed(() => (editingId.value ? '编辑菜品' : '新增菜品'))

function openAdd() {
  editingId.value = null
  form.value = { name: '', price: 0, category: '', description: '', image: '', status: 1 }
  dialogVisible.value = true
}

function openEdit(row) {
  editingId.value = row.id
  form.value = {
    name: row.name,
    price: row.price,
    category: row.category,
    description: row.description,
    image: row.image,
    status: row.status,
  }
  dialogVisible.value = true
}

async function save() {
  if (!form.value.name) {
    ElMessage.warning('请输入菜品名称')
    return
  }
  if (!(form.value.price > 0)) {
    ElMessage.warning('价格必须大于 0')
    return
  }
  saving.value = true
  try {
    if (editingId.value) {
      await updateProduct(editingId.value, form.value)
    } else {
      await createProduct({ ...form.value, shop_id: auth.currentShopId })
    }
    ElMessage.success('已保存')
    dialogVisible.value = false
    await load()
  } catch {
    // handled by interceptor
  } finally {
    saving.value = false
  }
}

async function toggleStatus(row) {
  const next = row.status === 1 ? 0 : 1
  try {
    await updateProduct(row.id, { status: next })
    ElMessage.success(next === 1 ? '已上架' : '已下架')
    await load()
  } catch {
    // handled by interceptor
  }
}

async function remove(row) {
  try {
    await ElMessageBox.confirm(`确定删除「${row.name}」？`, '删除菜品', { type: 'warning' })
  } catch {
    return // cancelled
  }
  try {
    await deleteProduct(row.id)
    ElMessage.success('已删除')
    await load()
  } catch {
    // handled by interceptor
  }
}

// ---- Spec (variant) management ----
const specDialogVisible = ref(false)
const specProduct = ref(null)
const newSpec = ref({ name: '', price: 0, status: 1 })
const specSaving = ref(false)

function openSpecs(row) {
  specProduct.value = row
  newSpec.value = { name: '', price: 0, status: 1 }
  specDialogVisible.value = true
}

// Reload products and re-point specProduct at the refreshed row.
async function reloadSpecProduct() {
  await load()
  if (specProduct.value) {
    const fresh = allProducts.value.find((p) => p.id === specProduct.value.id)
    if (fresh) specProduct.value = fresh
  }
}

async function addSpec() {
  if (!newSpec.value.name) {
    ElMessage.warning('请输入规格名称')
    return
  }
  if (!(newSpec.value.price > 0)) {
    ElMessage.warning('价格必须大于 0')
    return
  }
  specSaving.value = true
  try {
    await createProductSpec(specProduct.value.id, { ...newSpec.value })
    ElMessage.success('已添加规格')
    newSpec.value = { name: '', price: 0, status: 1 }
    await reloadSpecProduct()
  } catch {
    // handled by interceptor
  } finally {
    specSaving.value = false
  }
}

async function saveSpec(spec) {
  try {
    await updateProductSpec(spec.id, { name: spec.name, price: spec.price, status: spec.status })
    ElMessage.success('已保存')
    await reloadSpecProduct()
  } catch {
    // handled by interceptor
  }
}

async function removeSpec(spec) {
  try {
    await ElMessageBox.confirm(`删除规格「${spec.name}」？`, '删除规格', { type: 'warning' })
  } catch {
    return // cancelled
  }
  try {
    await deleteProductSpec(spec.id)
    ElMessage.success('已删除')
    await reloadSpecProduct()
  } catch {
    // handled by interceptor
  }
}

// ---- Image upload ----
function beforeUpload(file) {
  const okType = ['image/jpeg', 'image/png', 'image/webp'].includes(file.type)
  if (!okType) {
    ElMessage.error('只支持 jpg/png/webp')
    return false
  }
  if (file.size > 5 * 1024 * 1024) {
    ElMessage.error('图片不能超过 5MB')
    return false
  }
  return true
}

async function customUpload(option) {
  try {
    const res = await uploadImage(option.file)
    form.value.image = res.url
    ElMessage.success('图片已上传')
  } catch (e) {
    option.onError(e)
  }
}
</script>

<template>
  <el-card>
    <template #header>
      <div class="header">
        <span>菜品管理</span>
        <el-button type="primary" :disabled="!auth.currentShopId" @click="openAdd">
          新增菜品
        </el-button>
      </div>
    </template>

    <el-alert
      v-if="!auth.currentShopId"
      type="info"
      :closable="false"
      title="请先在顶部选择店铺；若还没有店铺，请到「店铺与返利」创建。"
    />

    <el-table v-else v-loading="loading" :data="products" stripe>
      <el-table-column label="图片" width="80">
        <template #default="{ row }">
          <el-image
            v-if="row.image"
            :src="row.image"
            fit="cover"
            style="width: 48px; height: 48px; border-radius: 4px"
            :preview-src-list="[row.image]"
            preview-teleported
          />
          <span v-else class="no-img">—</span>
        </template>
      </el-table-column>
      <el-table-column prop="name" label="名称" min-width="120" />
      <el-table-column prop="category" label="分类" width="120" />
      <el-table-column label="价格" width="100">
        <template #default="{ row }">¥{{ row.price }}</template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="STATUS[row.status]?.type">{{ STATUS[row.status]?.label }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="规格" width="90">
        <template #default="{ row }">
          <el-button link type="primary" @click="openSpecs(row)">
            规格({{ row.specs ? row.specs.length : 0 }})
          </el-button>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="220">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="warning" @click="toggleStatus(row)">
            {{ row.status === 1 ? '下架' : '上架' }}
          </el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
      <template #empty>暂无菜品</template>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="480px">
      <el-form label-width="72px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="菜品名称" />
        </el-form-item>
        <el-form-item label="价格" required>
          <el-input-number v-model="form.price" :min="0" :precision="2" :step="1" />
        </el-form-item>
        <el-form-item label="分类">
          <el-input v-model="form.category" placeholder="如：热菜 / 饮品" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="图片">
          <div class="upload-row">
            <el-image
              v-if="form.image"
              :src="form.image"
              fit="cover"
              style="width: 64px; height: 64px; border-radius: 4px; margin-right: 12px"
            />
            <el-upload
              :show-file-list="false"
              :before-upload="beforeUpload"
              :http-request="customUpload"
              accept="image/jpeg,image/png,image/webp"
            >
              <el-button>{{ form.image ? '更换图片' : '上传图片' }}</el-button>
            </el-upload>
          </div>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width: 120px">
            <el-option :value="1" label="上架" />
            <el-option :value="0" label="下架" />
            <el-option :value="2" label="售罄" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="specDialogVisible"
      :title="`规格管理 - ${specProduct?.name || ''}`"
      width="560px"
    >
      <el-table :data="specProduct?.specs || []" size="small">
        <el-table-column label="名称">
          <template #default="{ row }"><el-input v-model="row.name" size="small" /></template>
        </el-table-column>
        <el-table-column label="价格" width="150">
          <template #default="{ row }">
            <el-input-number v-model="row.price" :min="0" :precision="2" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <el-select v-model="row.status" size="small">
              <el-option :value="1" label="上架" />
              <el-option :value="0" label="下架" />
              <el-option :value="2" label="售罄" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="130">
          <template #default="{ row }">
            <el-button link type="primary" @click="saveSpec(row)">保存</el-button>
            <el-button link type="danger" @click="removeSpec(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty>暂无规格（不设置则按菜品价格售卖）</template>
      </el-table>

      <div class="spec-add">
        <el-input v-model="newSpec.name" placeholder="规格名，如 600ml" style="width: 160px" />
        <el-input-number v-model="newSpec.price" :min="0" :precision="2" :step="1" />
        <el-select v-model="newSpec.status" style="width: 100px">
          <el-option :value="1" label="上架" />
          <el-option :value="0" label="下架" />
          <el-option :value="2" label="售罄" />
        </el-select>
        <el-button type="primary" :loading="specSaving" @click="addSpec">添加</el-button>
      </div>

      <template #footer>
        <el-button @click="specDialogVisible = false">关闭</el-button>
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
.no-img {
  color: #ccc;
}
.upload-row {
  display: flex;
  align-items: center;
}
.spec-add {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
}
</style>
