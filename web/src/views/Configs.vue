<template>
  <el-card shadow="never" class="config-card">
    <template #header><b>系统配置</b></template>
    <el-form label-width="180px">
      <el-form-item label="每人每日 AI 免费生成">
        <el-input-number v-model="aiLimit" :min="0" :max="1000" :step="1" />
        <span class="hint">次（0 = 关闭 AI 生成）</span>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </el-form-item>
    </el-form>
    <el-alert
      type="info"
      :closable="false"
      title="说明：AI 生图当前为 mock 占位实现，配额已生效；接入真实生图 API 后此配置直接作用于真实调用。"
    />
  </el-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { CONFIG_KEYS, listConfigs, updateConfig } from '@/api/modules/configs'

const aiLimit = ref(3)
const saving = ref(false)

onMounted(async () => {
  try {
    const cfgs = await listConfigs()
    const v = cfgs.find((c) => c.key === CONFIG_KEYS.aiDailyFreeLimit)?.value
    if (v !== undefined) aiLimit.value = parseInt(v, 10) || 0
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载失败')
  }
})

async function save() {
  saving.value = true
  try {
    await updateConfig(CONFIG_KEYS.aiDailyFreeLimit, String(aiLimit.value))
    ElMessage.success('已保存')
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.config-card {
  max-width: 560px;
}
.hint {
  margin-left: 8px;
  color: #909399;
}
</style>
