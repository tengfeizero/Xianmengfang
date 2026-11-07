<template>
  <div class="login">
    <!-- 品牌侧 -->
    <section class="brand-side">
      <div class="brand-top">
        <div class="mark">
          <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M4 19V9.5L12 4l8 5.5V19" stroke-linecap="round" stroke-linejoin="round" />
            <path d="M9 19v-5h6v5" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </div>
        <span>现梦坊 · 商家后台</span>
      </div>

      <div class="brand-copy">
        <h1>把孩子的想象<br />做成真实的玩偶</h1>
        <p>拼装创作、AI 生图、一键下单 —— 订单、商品与内容都在这里管理。</p>
        <ul class="points">
          <li><i />订单全流程：接单 · 生产 · 打包 · 发货</li>
          <li><i />拼装素材与商品规格统一维护</li>
          <li><i />角色权限与操作日志可追溯</li>
        </ul>
      </div>

      <div class="brand-foot">© 2026 现梦坊 · 单厂部署版</div>

      <!-- 装饰：模拟拼装图层 -->
      <div class="deco deco-1" />
      <div class="deco deco-2" />
      <div class="deco deco-3" />
    </section>

    <!-- 表单侧 -->
    <section class="form-side">
      <div class="login-card">
        <h2>登录后台</h2>
        <p class="sub">请使用管理员账号登录</p>

        <el-form :model="form" @keyup.enter="onLogin">
          <el-input
            v-model="form.username"
            placeholder="用户名"
            size="large"
            :prefix-icon="User"
            class="field"
          />
          <el-input
            v-model="form.password"
            type="password"
            placeholder="密码"
            size="large"
            show-password
            :prefix-icon="Lock"
            class="field"
          />
          <el-button type="primary" size="large" class="submit" :loading="loading" @click="onLogin">
            登 录
          </el-button>
        </el-form>

        <p class="hint">首次部署默认账号 admin / admin123，上线前请立即修改密码。</p>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Lock, User } from '@element-plus/icons-vue'
import { adminLogin } from '@/api/modules/auth'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()
const form = reactive({ username: '', password: '' })
const loading = ref(false)

async function onLogin() {
  if (!form.username || !form.password) {
    ElMessage.warning('请输入用户名和密码')
    return
  }
  loading.value = true
  try {
    const data = await adminLogin(form.username, form.password)
    auth.login(data.token)
    await auth.loadMe()
    router.push(auth.firstPath)
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login {
  display: flex;
  height: 100vh;
  background: #fff;
}

/* 品牌侧 */
.brand-side {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  width: 46%;
  min-width: 420px;
  padding: 36px 44px;
  overflow: hidden;
  color: #fff;
  background: radial-gradient(120% 90% at 10% 0%, #2a2f3d 0%, #1b1f2a 45%, #14171f 100%);
}
.brand-top {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  color: #c9cfdb;
  z-index: 2;
}
.mark {
  display: grid;
  place-items: center;
  width: 36px;
  height: 36px;
  border-radius: 10px;
  color: #fff;
  background: linear-gradient(135deg, var(--brand) 0%, #ff7a5c 100%);
  box-shadow: 0 6px 18px rgba(242, 69, 95, 0.35);
}
.brand-copy {
  z-index: 2;
}
.brand-copy h1 {
  margin: 0 0 16px;
  font-size: 34px;
  line-height: 1.35;
  letter-spacing: 0.5px;
  font-weight: 600;
}
.brand-copy p {
  margin: 0 0 22px;
  max-width: 430px;
  font-size: 14px;
  line-height: 1.8;
  color: #9aa3b5;
}
.points {
  margin: 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: 10px;
}
.points li {
  display: flex;
  align-items: center;
  gap: 9px;
  font-size: 13px;
  color: #b8c0cf;
}
.points i {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: var(--brand);
  flex-shrink: 0;
}
.brand-foot {
  font-size: 12px;
  color: #6d7688;
  z-index: 2;
}
/* 装饰圆：低饱和大圆营造层次，不抢内容 */
.deco {
  position: absolute;
  border-radius: 50%;
  pointer-events: none;
}
.deco-1 {
  right: -140px;
  top: -120px;
  width: 420px;
  height: 420px;
  background: radial-gradient(circle at 30% 30%, rgba(242, 69, 95, 0.22), transparent 65%);
}
.deco-2 {
  right: 40px;
  bottom: -180px;
  width: 380px;
  height: 380px;
  border: 1px solid rgba(255, 255, 255, 0.08);
}
.deco-3 {
  right: 120px;
  bottom: -110px;
  width: 240px;
  height: 240px;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

/* 表单侧 */
.form-side {
  flex: 1;
  display: grid;
  place-items: center;
  padding: 24px;
  background: #fff;
}
.login-card {
  width: 100%;
  max-width: 340px;
}
.login-card h2 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  letter-spacing: -0.3px;
  color: var(--text-1);
}
.sub {
  margin: 6px 0 26px;
  font-size: 13px;
  color: var(--text-3);
}
.field {
  margin-bottom: 14px;
}
.field :deep(.el-input__wrapper) {
  padding: 4px 14px;
  border-radius: var(--radius-sm);
  box-shadow: 0 0 0 1px var(--line-1) inset;
}
.field :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--brand) inset;
}
.submit {
  width: 100%;
  margin-top: 8px;
  height: 42px;
  font-size: 14px;
  letter-spacing: 2px;
}
.hint {
  margin: 18px 0 0;
  font-size: 12px;
  line-height: 1.7;
  color: var(--text-4);
}

@media (max-width: 880px) {
  .brand-side {
    display: none;
  }
}
</style>
