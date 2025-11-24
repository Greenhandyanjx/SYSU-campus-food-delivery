<template>
  <div class="my-bg">
    <!-- 主内容卡片，和 My 页面完全一致的样式 -->
    <div class="my-page support-page">
      <div class="support-chat-wrap">
        <UserChatList />
      </div>
    </div>
  </div>
</template>

<script setup>
import UserChatList from '@/components/Chat/UserChatList.vue'
</script>

<style scoped>
/* ====================== 完全复用 My 页面的背景 ====================== */
.my-bg {
  width: 100%;
  min-height: 100vh;
  background: url('/src/assets/login/img_denglu_bj.jpg') center/cover no-repeat;
  background-attachment: fixed;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 60px 0;
}

/* ====================== 主卡片样式（和 My 页面一模一样） ====================== */
.my-page {
  width: 60%;
  background: rgba(255, 248, 225, 0.96);        /* 淡米黄色毛玻璃底 */
  border-radius: 16px;
  box-shadow: 0 8px 24px rgba(255, 193, 7, 0.35);
  padding: 28px;                                /* 内部留白 */
  backdrop-filter: blur(6px);
  transition: 0.3s;
  position: relative;
  z-index: 2;
  /* 让卡片整体高度撑满可用空间 */
  min-height: calc(100vh - 120px);
  display: flex;
  flex-direction: column;
}

.my-page:hover {
  transform: translateY(-3px);
  box-shadow: 0 10px 28px rgba(255, 193,7,0.45);
}

/* ====================== Support 页面专属布局 ====================== */
.support-page {
  padding: 0;                       /* 去掉外层 padding，让聊天列表贴边 */
}

.support-chat-wrap {
  flex: 1;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  border-radius: 12px;
  overflow: hidden;
  background: #fff;
  box-shadow: 0 4px 16px rgba(255, 193, 7, 0.12);
}

/* 关键：让 UserChatList 完美填充 */
.support-chat-wrap :deep(.user-chat-list) {
  height: 100%;
  border-radius: 12px;
}

/* 确保消息列表可以滚动（通常组件内部已经有 ul/li 结构） */
.support-chat-wrap :deep(ul),
.support-chat-wrap :deep(.chat-message-list) {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
}

/* 可选：如果你组件最外层是 div 而不是 ul，也加上 */
.support-chat-wrap :deep(.message-container) {
  flex: 1;
  overflow-y: auto;
}

/* ====================== 移动端适配（和 My 页面保持一致） ====================== */
@media (max-width: 900px) {
  .my-page {
    width: 92%;
    padding: 16px;
    min-height: calc(100vh - 80px);
  }
  
  .support-page {
    padding: 0;
  }
  
  .support-chat-wrap {
    border-radius: 10px;
  }
}
</style>