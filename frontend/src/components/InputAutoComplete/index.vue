<!--  -->
<template>
  <div class="input-auto-complete">
    <el-input
      v-model="input"
      :placeholder="placeholder"
      style="width: 250px"
      clearable
      @clear="init"
      @keyup.enter.native="init"
    >
      <i
        slot="prefix"
        class="el-input__icon el-icon-search"
        style="cursor: pointer"
        @click="init"
      />
    </el-input>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{
  data?: Array<any>
  placeholder?: string
  ObKey?: string
}>()

const emit = defineEmits<{
  (e: 'init', value: any): void
}>()

const input = ref('')
const placeholder = props.placeholder ?? ''

function init() {
  emit('init', input.value)
}
</script>
<style scoped>
.input-auto-complete {
  display: inline-block;
}
</style>
