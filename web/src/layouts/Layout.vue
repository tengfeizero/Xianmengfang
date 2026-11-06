<template>
  <el-container class="layout">
    <el-aside width="200px" class="aside">
      <div class="logo">现梦坊</div>
      <el-menu :default-active="route.path" class="menu" router>
        <el-menu-item v-for="m in menus" :key="m.path" :index="m.path">{{ m.title }}</el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <b>{{ route.meta.title }}</b>
        <el-button link type="danger" @click="onLogout">退出登录</el-button>
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

const menus = [
  { path: '/orders', title: '订单管理' },
  { path: '/products', title: '商品管理' },
  { path: '/assets', title: '素材管理' },
  { path: '/configs', title: '系统配置' },
]

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
</style>
