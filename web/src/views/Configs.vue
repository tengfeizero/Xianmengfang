<template>
  <div class="page">
    <PageHeader title="系统配置" desc="影响用户端与 AI 生成行为的全局参数，保存后立即生效，改动会记入操作日志。" />

    <div class="config-wrap">
      <div class="panel">
        <div class="panel-head">
          <div class="panel-title">AI 生成配额</div>
        </div>
        <div class="panel-body">
          <el-form label-width="150px" class="cfg-form">
            <el-form-item label="每人每日免费次数">
              <el-input-number v-model="aiLimit" :min="0" :max="1000" :step="1" controls-position="right" />
              <span class="form-tip">次；0 表示关闭 AI 生成</span>
            </el-form-item>
            <el-form-item>
              <el-button v-if="canEdit" type="primary" :loading="saving" @click="save">保存</el-button>
              <span v-else class="form-tip">当前角色只能查看配置</span>
            </el-form-item>
          </el-form>
        </div>
      </div>

      <div class="panel">
        <div class="panel-head">
          <div class="panel-title">
            AI 生图 · 腾讯混元 Hy Image
            <el-tag v-if="keyMasked" type="success" effect="light" size="small">已配置</el-tag>
            <el-tag v-else type="info" effect="light" size="small">未配置，使用 mock</el-tag>
          </div>
        </div>
        <div class="panel-body">
          <el-form label-width="150px" class="cfg-form">
            <el-form-item label="Secret Key (sk)">
              <el-input
                v-model="aiApiKey"
                type="password"
                show-password
                :placeholder="keyMasked ? `已配置：${keyMasked}（留空不修改）` : '未配置，当前使用 mock 占位实现'"
                clearable
              />
            </el-form-item>
            <el-form-item label="Secret ID (ak)">
              <el-input
                v-model="aiSecretId"
                type="password"
                show-password
                :placeholder="sidMasked ? `已配置：${sidMasked}（留空不修改）` : '可选，TokenHub 控制台的 Key ID'"
                clearable
              />
            </el-form-item>
            <el-form-item label="接口地址">
              <el-input v-model="aiBaseUrl" placeholder="https://tokenhub.tencentmaas.com" />
            </el-form-item>
            <el-form-item label="模型 ID">
              <el-input v-model="aiModel" placeholder="hy-image-v3.5-preview" />
            </el-form-item>
            <el-form-item label="请求格式">
              <el-select v-model="aiStyle" class="w260">
                <el-option label="messages（腾讯云 TokenHub）" value="messages" />
                <el-option label="images（OpenAI 风格端点）" value="images" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="canEdit">
              <el-button type="primary" :loading="savingAI" @click="saveAI">保存 AI 配置</el-button>
              <el-button :disabled="!keyMasked" @click="clearKey">清空密钥</el-button>
            </el-form-item>
          </el-form>

          <div class="notice">
            <el-icon><InfoFilled /></el-icon>
            <span>密钥只存服务器数据库并脱敏展示，不进代码仓库；清空后 AI 生成自动回落 mock 占位实现。</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { InfoFilled } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import { CONFIG_KEYS, listConfigs, updateConfig } from '@/api/modules/configs'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const canEdit = computed(() => auth.hasPerm('config:edit'))

const aiLimit = ref(3)
const saving = ref(false)

const aiApiKey = ref('')
const aiSecretId = ref('')
const aiBaseUrl = ref('')
const aiModel = ref('')
const aiStyle = ref('messages')
const keyMasked = ref('')
const sidMasked = ref('')
const savingAI = ref(false)

onMounted(async () => {
  try {
    const cfgs = await listConfigs()
    const get = (k: string) => cfgs.find((c) => c.key === k)?.value
    const v = get(CONFIG_KEYS.aiDailyFreeLimit)
    if (v !== undefined) aiLimit.value = parseInt(v, 10) || 0
    aiBaseUrl.value = get(CONFIG_KEYS.aiBaseUrl) ?? ''
    aiModel.value = get(CONFIG_KEYS.aiModel) ?? ''
    aiStyle.value = get(CONFIG_KEYS.aiStyle) ?? 'messages'
    keyMasked.value = get(CONFIG_KEYS.aiApiKey) ?? ''
    sidMasked.value = get(CONFIG_KEYS.aiSecretId) ?? ''
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

async function saveAI() {
  savingAI.value = true
  try {
    await updateConfig(CONFIG_KEYS.aiBaseUrl, aiBaseUrl.value.trim())
    await updateConfig(CONFIG_KEYS.aiModel, aiModel.value.trim())
    await updateConfig(CONFIG_KEYS.aiStyle, aiStyle.value)
    // 密钥输入框留空 = 不修改（后台显示的是脱敏值）
    if (aiApiKey.value.trim() !== '') {
      await updateConfig(CONFIG_KEYS.aiApiKey, aiApiKey.value.trim())
      aiApiKey.value = ''
    }
    if (aiSecretId.value.trim() !== '') {
      await updateConfig(CONFIG_KEYS.aiSecretId, aiSecretId.value.trim())
      aiSecretId.value = ''
    }
    const cfgs = await listConfigs()
    keyMasked.value = cfgs.find((c) => c.key === CONFIG_KEYS.aiApiKey)?.value ?? ''
    sidMasked.value = cfgs.find((c) => c.key === CONFIG_KEYS.aiSecretId)?.value ?? ''
    ElMessage.success('AI 配置已保存，立即生效')
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  } finally {
    savingAI.value = false
  }
}

async function clearKey() {
  savingAI.value = true
  try {
    await updateConfig(CONFIG_KEYS.aiApiKey, '')
    await updateConfig(CONFIG_KEYS.aiSecretId, '')
    keyMasked.value = ''
    sidMasked.value = ''
    ElMessage.success('已清空密钥，AI 生成回落 mock')
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '操作失败')
  } finally {
    savingAI.value = false
  }
}
</script>

<style scoped>
.config-wrap {
  max-width: 860px;
}
.cfg-form :deep(.el-form-item) {
  margin-bottom: 18px;
}
.w260 {
  width: 260px;
}
.notice {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-top: 4px;
  padding: 10px 12px;
  border-radius: var(--radius-sm);
  background: #f7f9fc;
  border: 1px solid #eaf0f8;
  color: var(--text-3);
  font-size: 12.5px;
  line-height: 1.6;
}
.notice .el-icon {
  margin-top: 2px;
  color: #2b6ef2;
  flex-shrink: 0;
}
</style>
