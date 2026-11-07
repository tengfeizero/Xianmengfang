<template>
  <div class="stat-strip">
    <div v-for="(s, i) in items" :key="i" class="strip-item">
      <span class="strip-label">{{ s.label }}</span>
      <div class="strip-value" :class="s.tone">
        <span v-if="s.prefix" class="pre">{{ s.prefix }}</span>{{ s.value
        }}<i v-if="s.unit">{{ s.unit }}</i>
      </div>
      <span v-if="s.extra" class="strip-extra">{{ s.extra }}</span>
    </div>
  </div>
</template>

<script lang="ts">
// 明细页顶部汇总条：一屏看清「记录数 / 累计流入 / 累计流出」这类口径
export interface StatItem {
  label: string
  value: string | number
  prefix?: string
  unit?: string
  extra?: string
  tone?: 'up' | 'down'
}
</script>

<script setup lang="ts">
defineProps<{ items: StatItem[] }>()
</script>

<style scoped>
.stat-strip {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  margin-bottom: 16px;
  background: var(--panel-bg);
  border: 1px solid var(--line-1);
  border-radius: var(--radius);
  box-shadow: var(--shadow-panel);
  overflow: hidden;
}
.strip-item {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 16px 20px;
}
.strip-item + .strip-item::before {
  content: '';
  position: absolute;
  left: 0;
  top: 14px;
  bottom: 14px;
  width: 1px;
  background: var(--line-2);
}
.strip-label {
  font-size: 12px;
  color: var(--text-3);
}
.strip-value {
  display: flex;
  align-items: baseline;
  font-size: 24px;
  font-weight: 600;
  line-height: 1.2;
  color: var(--text-1);
  font-variant-numeric: tabular-nums;
}
.strip-value .pre {
  margin-right: 1px;
  font-size: 15px;
  font-weight: 500;
}
.strip-value i {
  margin-left: 3px;
  font-style: normal;
  font-size: 13px;
  font-weight: 400;
  color: var(--text-3);
}
.strip-value.up {
  color: var(--el-color-success);
}
.strip-value.down {
  color: var(--el-color-danger);
}
.strip-extra {
  font-size: 12px;
  color: var(--text-3);
}
</style>
