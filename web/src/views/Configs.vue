<template>
  <el-card shadow="never" style="max-width: 560px">
    <template #header><b>系统配置</b></template>
    <el-form label-width="160px">
      <el-form-item label="每人每日 AI 免费生成">
        <el-input-number v-model="aiLimit" :min="0" :max="1000" :step="1" />
        <span style="margin-left: 8px; color: #909399">次（0 = 关闭 AI 生成）</span>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </el-form-item>
    </el-form>
    <el-alert type="info" :closable="false"
      title="说明：AI 生图当前为 mock 占位实现，配额已生效；接入真实生图 API 后此配置直接作用于真实调用。" />
  </el-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { api } from '../api'

interface Config { key: string; value: string }

const aiLimit = ref(3)
const saving = ref(false)

onMounted(async () => {
  try {
    const cfgs = await api<Config[]>('GET', '/admin/configs')
    const v = cfgs.find((c) => c.key === 'ai_daily_free_limit')?.value
    if (v !== undefined) aiLimit.value = parseInt(v, 10) || 0
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载失败')
  }
})

async function save() {
  saving.value = true
  try {
    await api('PUT', '/admin/configs', { key: 'ai_daily_free_limit', value: String(aiLimit.value) })
    ElMessage.success('已保存')
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  } finally {
    saving.value = false
  }
}
</script>
