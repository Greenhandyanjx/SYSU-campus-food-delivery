<template>
  <div class="selectInput" ref="selectRef">
    <el-input
      v-model="displayText"
      placeholder="请选择口味"
      readonly
      clearable
      @focus="toggleDropdown(true)"
    />

    <!-- 下拉框部分 -->
    <transition name="fade">
      <div
        v-if="showDropdown && dishFlavorsData.length"
        class="flavorSelect"
      >
        <div
          v-for="(flavor, i) in dishFlavorsData"
          :key="i"
          class="flavor-group"
        >
          <div class="flavor-name">{{ flavor.name }}</div>
          <div class="flavor-values">
            <span
              v-for="(v, j) in flavor.value"
              :key="j"
              :class="[
                'flavor-option',
                selectedMap[flavor.name]?.includes(v) ? 'selected' : ''
              ]"
              @mousedown.prevent="toggleOption(flavor.name, v)"
            >
              {{ v }}
            </span>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'

const props = defineProps<{
  dishFlavorsData?: any[]
  value?: string[] // 传入父组件的选中值（多选）
  index?: number
}>()

const emit = defineEmits(['update:value', 'select'])

const showDropdown = ref(false)
const displayText = ref('') // 输入框展示文本
const selectRef = ref<HTMLElement>()

// 响应式的口味列表
const dishFlavorsData = computed(() => props.dishFlavorsData || [])

// 记录当前选中项，形如：{ 甜味: ['无糖', '半糖'], 辣度: ['中辣'] }
const selectedMap = ref<Record<string, string[]>>({})

// 初始化展示值
watch(
  () => props.value,
  (v) => {
    if (Array.isArray(v)) {
      displayText.value = v.join('、')
    }
  },
  { immediate: true }
)

// 切换下拉框显隐
function toggleDropdown(st?: boolean) {
  showDropdown.value = st ?? !showDropdown.value
}

// 点击外部自动关闭
function handleClickOutside(e: MouseEvent) {
  if (selectRef.value && !selectRef.value.contains(e.target as Node)) {
    showDropdown.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})
onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})

// 点击选项（多选）
function toggleOption(type: string, val: string) {
  const arr = selectedMap.value[type] || []
  const idx = arr.indexOf(val)
  if (idx === -1) {
    arr.push(val)
  } else {
    arr.splice(idx, 1)
  }
  selectedMap.value[type] = arr

  // 展示文本
  const allValues = Object.entries(selectedMap.value)
    .flatMap(([k, v]) => v.map((x) => `${k}:${x}`))
  displayText.value = allValues.join('、')

  // 通知父组件
  emit('update:value', allValues)
  emit('select', type, props.index, val)
}
</script>

<style lang="scss" scoped>
.selectInput {
  position: relative;
  width: 100%;
  min-width: 120px;

  .flavorSelect {
    position: absolute;
    width: 100%;
    background: #fff;
    border-radius: 6px;
    border: 1px solid #e5e6eb;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
    top: 42px;
    z-index: 999;
    padding: 6px 10px;
    max-height: 240px; /* 限制高度，防止太高 */
    overflow-y: auto;  /* 超出滚动 */
  }

  /* 分组标题 */
  .flavor-group {
    padding: 4px 0;
    border-bottom: 1px dashed #f2f2f2;
    &:last-child {
      border-bottom: none;
    }

    .flavor-name {
      font-weight: 600;
      font-size: 13px;
      color: #606266;
      margin-bottom: 4px;
      text-align: left;
    }

    .flavor-values {
      display: flex;
      flex-wrap: wrap;
      gap: 4px; /* 控制行间距 */
    }

    .flavor-option {
      display: inline-block;
      padding: 2px 8px;
      font-size: 12px;
      border-radius: 4px;
      border: 1px solid #dcdfe6;
      cursor: pointer;
      transition: all 0.2s ease;
      color: #606266;
      background: #fff;

      &:hover {
        background-color: #f5f7fa;
        border-color: #c0c4cc;
      }

      &.selected {
        background: #fff7e6;
        border-color: #ffb300;
        color: #ff9800;
        font-weight: 500;
      }
    }
  }
}

/* 淡入淡出动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
