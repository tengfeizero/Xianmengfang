<template>
  <div class="layout">
    <!-- 侧边导航 -->
    <aside class="side">
      <div class="brand">
        <div class="brand-mark">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M4 19V9.5L12 4l8 5.5V19" stroke-linecap="round" stroke-linejoin="round" />
            <path d="M9 19v-5h6v5" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </div>
        <div class="brand-text">
          <b>现梦坊</b>
          <span>商家后台</span>
        </div>
      </div>

      <el-scrollbar class="side-scroll">
        <el-menu :default-active="route.path" class="side-menu" router>
          <template v-for="g in auth.menuGroups" :key="g.key">
            <p v-if="g.name" class="menu-group">{{ g.name }}</p>
            <el-menu-item v-for="m in g.items" :key="m.path" :index="m.path">
              <el-icon class="menu-icon"><component :is="ICONS[m.icon]" /></el-icon>
              <span>{{ m.title }}</span>
            </el-menu-item>
          </template>
        </el-menu>
      </el-scrollbar>

      <div class="side-foot">
        <span class="health"><i />系统正常</span>
        <span class="ver">v1.4</span>
      </div>
    </aside>

    <!-- 主区域 -->
    <div class="main">
      <header class="topbar">
        <div class="crumbs">
          <span class="root">现梦坊</span>
          <el-icon class="sep"><ArrowRight /></el-icon>
          <span class="current">{{ route.meta.title || '工作台' }}</span>
        </div>

        <el-dropdown trigger="click" @command="onCommand">
          <div class="account">
            <el-avatar :size="30" class="avatar">{{ initial }}</el-avatar>
            <div class="account-text">
              <b>{{ displayName }}</b>
              <span>{{ auth.me?.role.name || '—' }}</span>
            </div>
            <el-icon class="caret"><ArrowDown /></el-icon>
          </div>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item disabled>账号：{{ auth.me?.username || '—' }}</el-dropdown-item>
              <el-dropdown-item disabled>角色：{{ auth.me?.role.name || '—' }}</el-dropdown-item>
              <el-dropdown-item divided command="logout">
                <el-icon><SwitchButton /></el-icon>
                退出登录
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </header>

      <main class="content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, type Component } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ArrowDown,
  ArrowRight,
  Avatar,
  Document,
  Goods,
  Key,
  MagicStick,
  Odometer,
  Picture,
  Setting,
  SwitchButton,
  Tickets,
  User,
} from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

/** 菜单图标表：config/menu.ts 里只存名字，避免配置文件依赖组件 */
const ICONS: Record<string, Component> = {
  Odometer,
  Tickets,
  Goods,
  Picture,
  MagicStick,
  User,
  Avatar,
  Key,
  Document,
  Setting,
}

const displayName = computed(() => auth.me?.nickname || auth.me?.username || '未登录')
const initial = computed(() => (displayName.value || '?').slice(0, 1).toUpperCase())

function onCommand(cmd: string) {
  if (cmd === 'logout') {
    auth.logout()
    router.push('/login')
  }
}
</script>

<style scoped>
.layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

/* ---------- 侧栏 ---------- */
.side {
  display: flex;
  flex-direction: column;
  width: var(--side-w);
  flex-shrink: 0;
  background: var(--side-bg);
  color: var(--side-text);
}
.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  height: var(--header-h);
  padding: 0 18px;
  background: var(--side-bg-deep);
  flex-shrink: 0;
}
.brand-mark {
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  border-radius: 8px;
  color: #fff;
  background: linear-gradient(135deg, var(--brand) 0%, #ff7a5c 100%);
  box-shadow: 0 2px 8px rgba(242, 69, 95, 0.35);
}
.brand-text {
  display: flex;
  align-items: baseline;
  gap: 6px;
  line-height: 1;
}
.brand-text b {
  color: #fff;
  font-size: 15px;
  letter-spacing: 0.5px;
}
.brand-text span {
  font-size: 11px;
  color: var(--side-text-dim);
}

.side-scroll {
  flex: 1;
  padding: 10px 0 16px;
}

.side-menu {
  border-right: none;
  background: transparent;
  --el-menu-bg-color: transparent;
  --el-menu-text-color: var(--side-text);
  --el-menu-hover-bg-color: var(--side-hover);
  --el-menu-active-color: #fff;
  --el-menu-item-height: 42px;
  --el-menu-base-level-padding: 16px;
}
.side-menu :deep(.el-menu-item) {
  position: relative;
  height: 42px;
  margin: 2px 10px;
  padding-left: 14px !important;
  padding-right: 12px;
  border-radius: 8px;
  font-size: 13.5px;
  transition: background-color 0.15s ease, color 0.15s ease;
}
.side-menu :deep(.el-menu-item:hover) {
  color: #fff;
}
.side-menu :deep(.el-menu-item.is-active) {
  color: #fff;
  background: var(--side-active);
  font-weight: 500;
}
/* 选中态左侧竖条：电商后台的经典指示 */
.side-menu :deep(.el-menu-item.is-active)::before {
  content: '';
  position: absolute;
  left: -10px;
  top: 10px;
  bottom: 10px;
  width: 3px;
  border-radius: 0 3px 3px 0;
  background: var(--brand);
}
.menu-icon {
  margin-right: 10px;
  font-size: 16px;
  color: var(--side-text-dim);
  transition: color 0.15s ease;
}
.side-menu :deep(.el-menu-item:hover) .menu-icon,
.side-menu :deep(.el-menu-item.is-active) .menu-icon {
  color: var(--brand);
}

.menu-group {
  margin: 16px 0 6px;
  padding: 0 24px;
  font-size: 11px;
  letter-spacing: 1.5px;
  color: var(--side-text-dim);
}
.menu-group:first-child {
  margin-top: 4px;
}

.side-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 42px;
  padding: 0 18px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  font-size: 11px;
  color: var(--side-text-dim);
  flex-shrink: 0;
}
.health {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.health i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #2ec27e;
  box-shadow: 0 0 0 3px rgba(46, 194, 126, 0.15);
}
.ver {
  font-family: ui-monospace, Menlo, Consolas, monospace;
}

/* ---------- 主区域 ---------- */
.main {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
  background: var(--page-bg);
}

.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: var(--header-h);
  padding: 0 22px;
  background: #fff;
  border-bottom: 1px solid var(--line-1);
  flex-shrink: 0;
}
.crumbs {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-3);
}
.crumbs .sep {
  font-size: 11px;
  color: var(--text-4);
}
.crumbs .current {
  color: var(--text-1);
  font-weight: 500;
}

.account {
  display: flex;
  align-items: center;
  gap: 9px;
  padding: 4px 8px 4px 4px;
  border-radius: 20px;
  cursor: pointer;
  transition: background-color 0.15s ease;
}
.account:hover {
  background: var(--fill-1);
}
.avatar {
  background: linear-gradient(135deg, var(--brand) 0%, #ff8a5c 100%);
  color: #fff;
  font-size: 13px;
  font-weight: 600;
}
.account-text {
  display: flex;
  flex-direction: column;
  line-height: 1.25;
}
.account-text b {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-1);
}
.account-text span {
  font-size: 11px;
  color: var(--text-3);
}
.caret {
  font-size: 12px;
  color: var(--text-4);
}

.content {
  flex: 1;
  overflow-y: auto;
  padding: 18px 22px 28px;
}
</style>
