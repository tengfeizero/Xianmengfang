<template>
  <el-container style="height: 100vh">
    <el-aside width="200px" style="background: #001529">
      <div class="logo">现梦坊</div>
      <el-menu
        :default-active="$route.path"
        background-color="#001529"
        text-color="#c8c9cc"
        active-text-color="#ffffff"
        router
      >
        <el-menu-item index="/orders">订单管理</el-menu-item>
        <el-menu-item index="/products">商品管理</el-menu-item>
        <el-menu-item index="/assets">素材管理</el-menu-item>
        <el-menu-item index="/configs">系统配置</el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header style="background: #fff; display: flex; align-items: center; justify-content: space-between">
        <b>{{ $route.meta.title }}</b>
        <el-button link type="danger" @click="onLogout">退出登录</el-button>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { clearToken } from '../api'

const router = useRouter()

function onLogout() {
  clearToken()
  router.push('/login')
}
</script>

<style scoped>
.logo {
  color: #fff;
  font-size: 18px;
  font-weight: bold;
  text-align: center;
  padding: 20px 0;
}
</style>
