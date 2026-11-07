<template>
  <el-container class="layout">
    <el-aside width="200px" class="aside">
      <div class="logo">现梦坊</div>
      <el-menu :default-active="route.path" class="menu" router>
        <el-menu-item v-for="m in auth.menus" :key="m.path" :index="m.path">{{ m.title }}</el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <b>{{ route.meta.title }}</b>
        <div class="user">
          <span class="who">
            {{ auth.me?.nickname || auth.me?.username || '—' }}
            <el-tag v-if="auth.me" size="small" type="info" class="role">{{ auth.me.role.name }}</el-tag>
          </span>
          <el-button link type="danger" @click="onLogout">退出登录</el-button>
        </div>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

function onLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.layout {
  height: 100vh;
}
.aside {
  background: #001529;
}
.logo {
  color: #fff;
  font-size: 18px;
  font-weight: bold;
  text-align: center;
  padding: 20px 0;
}
.menu {
  border-right: none;
  background: #001529;
  --el-menu-text-color: #c8c9cc;
  --el-menu-hover-bg-color: #0a2540;
  --el-menu-active-color: #fff;
}
.header {
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.user {
  display: flex;
  align-items: center;
  gap: 12px;
}
.who {
  color: #606266;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.role {
  transform: scale(0.92);
}
</style>
