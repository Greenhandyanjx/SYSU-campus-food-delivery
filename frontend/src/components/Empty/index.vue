<template>
	<div class="empty-box">
		<img src="@/assets/table_empty.png" alt="empty" v-if="!isSearch" />
		<img src="@/assets/search_table_empty.png" alt="empty" v-else />
		<p>{{ isSearch ? '没有搜索到相关结果' : '暂无数据' }}</p>
	</div>
</template>

<script setup lang="ts">
const props = defineProps<{ isSearch?: boolean }>()
const isSearch = props.isSearch || false
</script>

<style scoped>
.empty-box {
	padding: 40px;
	text-align: center;
	color: #999;
}
.empty-box img {
	width: 120px;
	height: auto;
	margin-bottom: 12px;
}
</style>
