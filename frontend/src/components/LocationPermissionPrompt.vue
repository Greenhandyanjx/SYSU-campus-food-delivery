<template>
  <div v-if="show" class="location-permission-prompt" :class="{ 'visible': isVisible }">
    <div class="prompt-content">
      <div class="prompt-icon">
        <i class="iconfont" :class="iconClass"></i>
      </div>
      <div class="prompt-text">
        <div class="prompt-title">{{ title }}</div>
        <div class="prompt-message">{{ message }}</div>
      </div>
      <div class="prompt-actions" v-if="showActions">
        <el-button size="small" @click="checkAndRequestPermission" :loading="requesting">
          重新定位
        </el-button>
        <el-button size="small" type="primary" plain @click="dismiss">
          知道了
        </el-button>
      </div>
      <button v-else class="close-btn" @click="dismiss">✕</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { ElMessage } from 'element-plus';

interface Props {
  error?: string | null;
  showActions?: boolean;
  autoHide?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  error: null,
  showActions: true,
  autoHide: true
});

const emit = defineEmits<{
  dismiss: [];
}>();

const show = ref(false);
const isVisible = ref(false);
const requesting = ref(false);

let hideTimer: number | null = null;

const title = computed(() => {
  if (!props.error) return '定位权限检查';
  if (props.error.includes('权限被拒绝')) return '定位权限被拒绝';
  if (props.error.includes('定位失败')) return '定位服务异常';
  return '定位提示';
});

const message = computed(() => {
  if (!props.error) return '正在检查定位权限...';
  if (props.error.includes('权限被拒绝')) {
    return '骑手配送需要位置信息。请在浏览器设置中允许定位，然后点击重新定位。';
  }
  if (props.error.includes('定位失败')) {
    return '无法获取当前位置，请检查网络连接或尝试刷新页面。';
  }
  return props.error;
});

const iconClass = computed(() => {
  if (!props.error) return 'icon-location-off';
  if (props.error.includes('权限被拒绝')) return 'icon-location-error';
  if (props.error.includes('定位失败')) return 'icon-location-error';
  return 'icon-location-warning';
});

const checkAndRequestPermission = async () => {
  if (requesting.value) return;

  requesting.value = true;

  try {
    // 检查浏览器是否支持定位
    if (!navigator.geolocation) {
      ElMessage.error('浏览器不支持定位功能');
      return;
    }

    // 请求定位权限
    await new Promise<GeolocationPosition>((resolve, reject) => {
      navigator.geolocation.getCurrentPosition(
        resolve,
        reject,
        {
          enableHighAccuracy: true,
          timeout: 10000,
          maximumAge: 0
        }
      );
    });

    ElMessage.success('定位权限获取成功');
    emit('dismiss');

  } catch (error: any) {
    let errorMessage = '定位权限获取失败';

    switch (error.code) {
      case error.PERMISSION_DENIED:
        errorMessage = '定位权限被拒绝，请在浏览器设置中允许定位';
        break;
      case error.POSITION_UNAVAILABLE:
        errorMessage = '无法获取位置信息';
        break;
      case error.TIMEOUT:
        errorMessage = '定位请求超时';
        break;
    }

    ElMessage.error(errorMessage);
  } finally {
    requesting.value = false;
  }
};

const dismiss = () => {
  isVisible.value = false;
  setTimeout(() => {
    show.value = false;
  }, 300);
  emit('dismiss');
};

const showPrompt = () => {
  show.value = true;
  // 添加动画延迟
  setTimeout(() => {
    isVisible.value = true;
  }, 100);

  // 自动隐藏
  if (props.autoHide && !props.error) {
    if (hideTimer) {
      clearTimeout(hideTimer);
    }
    hideTimer = window.setTimeout(() => {
      dismiss();
    }, 5000);
  }
};

const hidePrompt = () => {
  dismiss();
};

// 暴露方法给父组件
defineExpose({
  show: showPrompt,
  hide: hidePrompt
});

// 监听错误变化
onMounted(() => {
  if (props.error) {
    showPrompt();
  }
});

onUnmounted(() => {
  if (hideTimer) {
    clearTimeout(hideTimer);
  }
});
</script>

<style scoped lang="scss">
.location-permission-prompt {
  position: fixed;
  top: 80px;
  left: 50%;
  transform: translateX(-50%) translateY(-20px);
  z-index: 1000;
  transition: all 0.3s ease;
  opacity: 0;
  pointer-events: none;

  &.visible {
    transform: translateX(-50%) translateY(0);
    opacity: 1;
    pointer-events: all;
  }
}

.prompt-content {
  background: #fff;
  border-radius: 12px;
  padding: 16px 20px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  border-left: 4px solid #ff6b6b;
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 320px;
  max-width: 400px;
}

.prompt-icon {
  font-size: 24px;
  flex-shrink: 0;

  .iconfont {
    font-size: 24px;
    display: block;
  }

  .icon-location-error {
    color: #ff6b6b;
  }

  .icon-location-warning {
    color: #ffa726;
  }

  .icon-location-off {
    color: #42a5f5;
  }
}

.prompt-text {
  flex: 1;
}

.prompt-title {
  font-size: 14px;
  font-weight: 600;
  color: #333;
  margin-bottom: 4px;
}

.prompt-message {
  font-size: 13px;
  color: #666;
  line-height: 1.4;
}

.prompt-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.close-btn {
  background: none;
  border: none;
  color: #999;
  font-size: 18px;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s ease;

  &:hover {
    background: #f5f5f5;
    color: #666;
  }
}

// 移动端适配
@media (max-width: 768px) {
  .location-permission-prompt {
    top: 70px;
    left: 20px;
    right: 20px;
    transform: translateX(0) translateY(-20px);

    &.visible {
      transform: translateX(0) translateY(0);
    }
  }

  .prompt-content {
    min-width: auto;
    border-left: none;
    border-top: 4px solid #ff6b6b;
  }

  .prompt-actions {
    flex-direction: column;
    gap: 6px;

    :deep(.el-button) {
      font-size: 12px;
      padding: 6px 12px;
    }
  }
}
</style>