<template>
  <div class="search-suggest" ref="root">
    <el-input
      v-model="localValue"
      :placeholder="placeholder"
      :class="['search-input', compact ? 'compact' : '']"
      clearable
      @focus="open"
      @input="onInput"
      @keyup.down.prevent="onDown"
      @keyup.up.prevent="onUp"
      @keyup.enter.prevent="onEnter"
      @keyup.esc="close"
    >
      <template #suffix>
        <el-button class="search-btn" type="warning" round @click="onClickSearch">
          <el-icon><Search /></el-icon>
        </el-button>
      </template>
    </el-input>

    <div v-show="show && items.length" class="suggestions" ref="panel">
      <ul>
        <li v-for="(it, idx) in items" :key="idx" :class="['sugg', {active: idx === highlighted}]"
            @mouseenter="highlighted = idx" @click="select(it)">
          <div class="sugg-name" v-html="hl(getName(it))"></div>
          <div class="sugg-desc">{{ getDesc(it) }}</div>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { Search } from '@element-plus/icons-vue'
import sampleStores from '@/data/sampleStores.js'

const props = defineProps({
  modelValue: { type: String, default: '' },
  placeholder: { type: String, default: '搜索店铺 / 美食' },
  compact: { type: Boolean, default: false },
  maxRecommend: { type: Number, default: 5 },
  stores: { type: Array, default: () => sampleStores }
})
const emit = defineEmits(['update:modelValue', 'search', 'select'])

const localValue = ref(props.modelValue)
watch(() => props.modelValue, v => localValue.value = v)
watch(localValue, v => emit('update:modelValue', v))

const show = ref(false)
const highlighted = ref(-1)
const root = ref<HTMLElement | null>(null)
const panel = ref<HTMLElement | null>(null)

const items = computed(() => {
  const q = (localValue.value || '').toString().trim().toLowerCase()
  if (!q) return (props.stores || []).slice(0, props.maxRecommend)
  const nameMatches = (props.stores || []).filter((s: any) => {
    const n = getName(s).toLowerCase()
    return n.includes(q)
  })
  const dishMatches = (props.stores || []).filter((s: any) => (s.dishes || []).some((d: any) => ((d.name || d.title || '') + '').toLowerCase().includes(q)))
  const combined = [...nameMatches, ...dishMatches].filter((v, i, a) => a.indexOf(v) === i)
  return combined.length ? combined : (props.stores || []).slice(0, props.maxRecommend)
})

function getName(it: any){
  return (it && (it.name || it.title || it.storeName || it.shopName || ''))
}
function getDesc(it: any){
  if (!it) return ''
  // prefer description-like fields; fallback to tags or first dish
  const desc = it.desc || it.description || it.summary || it.brief || it.shortDesc || ''
  if (desc) return desc
  if (Array.isArray(it.tags) && it.tags.length) return it.tags.slice(0,3).join(' · ')
  if (Array.isArray(it.categories) && it.categories.length) return it.categories.slice(0,3).join(' · ')
  if (Array.isArray(it.dishes) && it.dishes.length) return (it.dishes[0].name || it.dishes[0].title || '')
  return ''
}

function open(){ show.value = true; highlighted.value = -1 }
function close(){ show.value = false; highlighted.value = -1 }

function onInput(){ /* localValue already updated via v-model */ }
function onClickSearch(){ emit('search', localValue.value); close() }
function select(it: any){ emit('select', it); close() }

function onDown(){ if (!items.value.length) return; highlighted.value = Math.min(highlighted.value + 1, items.value.length - 1) }
function onUp(){ if (!items.value.length) return; highlighted.value = Math.max(highlighted.value - 1, 0) }
function onEnter(){ if (highlighted.value >= 0 && items.value[highlighted.value]) select(items.value[highlighted.value]); else { emit('search', localValue.value); close() } }

function hl(text: string){ const q = (localValue.value || '').toString().trim(); if (!q) return text || ''; try { const re = new RegExp('(' + q.replace(/[.*+?^${}()|[\\]\\]/g, '\\$&') + ')', 'ig'); return (text||'').toString().replace(re, '<strong>$1</strong>') } catch(e){ return text||'' } }

function onDocClick(e: MouseEvent){
  const t = e.target as Node
  if (!root.value) return
  if (!root.value.contains(t)) close()
}

onMounted(()=> document.addEventListener('click', onDocClick))
onBeforeUnmount(()=> document.removeEventListener('click', onDocClick))
</script>

<style scoped>
.search-suggest { width: 100%; position: relative }
.search {
  width: 80%;
  position: relative; /* 让下拉可以绝对定位 */
}

/* 调整 el-input 外观 */
.search-input :deep(.el-input__wrapper) {
  background-color: #fffef4;
  border-radius: 30px;
  border: 2px solid #faad14;
  box-shadow: 0 2px 6px rgba(250, 173, 20, 0.25);
  padding-right: 0px; /* 给按钮留出空间 */
  height: 46px;
  transition: 0.2s;
}

/* 聚焦效果 */
.search-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 3px rgba(255, 213, 79, 0.3);
}
/* 内部文字 */
.search-input :deep(.el-input__inner) {
  font-size: 15px;
  color: #8c6d1f;
}
/* 搜索按钮样式（嵌入输入框内部） */
.search-btn {
  position: absolute;
  right: 5px;
  top: 1px;
  height: 38px;
  width: 38px;
  border-radius: 50%;
  background-color: #faad14;
  color: white;
  border: none;
  box-shadow: 0 2px 4px rgba(250, 173, 20, 0.4);
  cursor: pointer;
  transition: 0.2s;
}
.search-btn:hover {
  background-color: #ffd666;
  color: #ad8b00;
}
.suggestions { position: absolute; left: 0; top: calc(100% + 8px); width: 100%; background: #fff; border-radius: 8px; box-shadow: 0 8px 24px rgba(0,0,0,0.12); z-index: 1500; max-height: 320px; overflow: auto; border: 1px solid rgba(0,0,0,0.06) }
.suggestions ul{ list-style:none; margin:0; padding:8px 0 }
.sugg{ padding: 10px 14px; cursor:pointer; display:flex; flex-direction:column; gap:4px }
.sugg + .sugg{ border-top:1px solid rgba(0,0,0,0.06) }
.sugg:hover, .sugg.active { background:#fff9e6 }
.sugg-name strong{ background: rgba(255,235,59,0.5); padding:0 2px }
.sugg-desc{ font-size:12px; color:#888 }
.search-input.compact :deep(.el-input__inner){ padding:6px 36px 6px 8px; font-size:14px }
</style>
