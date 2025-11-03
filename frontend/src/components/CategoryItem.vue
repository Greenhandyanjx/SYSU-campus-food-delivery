<template>
  <div
    class="category-item"
    :class="{ active }
    "
    @click.stop="onClick"
    role="button"
    tabindex="0"
    @keyup.enter="onClick"
  >
    <div class="icon-wrap">
      <img :src="icon" alt="" />
    </div>
    <div class="label">{{ label }}</div>
  </div>
</template>

<script setup lang="ts">
import { defineProps, defineEmits } from 'vue'

const props = defineProps<{ label: string; icon: string; active?: boolean }>()
const emit = defineEmits<{
  (e: 'select'): void
}>()

function onClick() {
  emit('select')
}
</script>

<style scoped>
.category-item {
  width: 72px;
  text-align: center;
  cursor: pointer;
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 8px 6px;
  border-radius: 12px;
  transition: transform 0.18s ease, box-shadow 0.18s ease, background-color 0.18s ease;
  background: transparent;
}
.category-item .icon-wrap {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fffef4;
  border: 1px solid rgba(250,173,20,0.12);
  box-shadow: 0 1px 4px rgba(250,173,20,0.06);
}
.category-item img { width: 26px; height: 26px }
.category-item .label {
  font-size: 12px;
  color: #8c6d1f;
}
.category-item:hover {
  transform: translateY(-6px) scale(1.02);
  box-shadow: 0 8px 20px rgba(250,173,20,0.18);
  background: linear-gradient(180deg, rgba(255,245,204,0.5), rgba(255,250,230,0.6));
}
.category-item.active {
  background: linear-gradient(180deg, #fff7e6, #fff3d6);
  box-shadow: 0 10px 26px rgba(250,173,20,0.24);
  transform: translateY(-6px);
}
</style>
