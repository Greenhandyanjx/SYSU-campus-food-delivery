<template>
  <div class="orders-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-content">
        <div class="title-section">
          <h1 class="page-title">
            <span class="title-icon">🚚</span>
            进行中
          </h1>
          <p class="page-subtitle">查看并管理您的配送订单</p>
        </div>
        <div class="header-actions">
          <div class="location-status-info" :class="{ 'tracking': isLocationTracking, 'error': locationError }">
            <i class="iconfont" :class="locationStatusIcon"></i>
            <span>{{ locationStatusText }}</span>
          </div>
          <el-button :loading="loading" type="primary" size="large" @click="load">
            <i class="iconfont icon-refresh"></i>
            刷新状态
          </el-button>
        </div>
      </div>
    </div>

    <!-- 统计卡片区域 -->
    <div class="stats-container" v-if="!loading">
      <div class="stat-card" :class="{ active: tab === '3' }" @click="tab = '3'">
        <div class="stat-icon">🏪</div>
        <div class="stat-content">
          <div class="stat-value">{{ count3 }}</div>
          <div class="stat-label">待取餐</div>
          <div class="stat-desc">商家已准备好</div>
        </div>
        <div class="stat-arrow" v-if="count3 > 0">
          <i class="el-icon-arrow-right"></i>
        </div>
      </div>
      <div class="stat-card" :class="{ active: tab === '4' }" @click="tab = '4'">
        <div class="stat-icon">🛵</div>
        <div class="stat-content">
          <div class="stat-value">{{ count4 }}</div>
          <div class="stat-label">派送中</div>
          <div class="stat-desc">正在配送途中</div>
        </div>
        <div class="stat-arrow" v-if="count4 > 0">
          <i class="el-icon-arrow-right"></i>
        </div>
      </div>
    </div>

    <!-- 标签页 -->
    <el-tabs v-model="tab" class="status-tabs" v-if="!loading">
      <el-tab-pane name="3">
        <template #label>
          <span class="tab-label">
            <span class="tab-icon">🏪</span>
            待取餐
            <span class="tab-badge" v-if="count3 > 0">{{ count3 }}</span>
          </span>
        </template>
      </el-tab-pane>
      <el-tab-pane name="4">
        <template #label>
          <span class="tab-label">
            <span class="tab-icon">🛵</span>
            派送中
            <span class="tab-badge" v-if="count4 > 0">{{ count4 }}</span>
          </span>
        </template>
      </el-tab-pane>
    </el-tabs>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="3" animated />
    </div>

    <!-- 订单列表 -->
    <div class="orders-grid" v-else-if="filtered.length > 0">
      <TransitionGroup name="order-list" tag="div">
        <RiderOrderCard
          v-for="o in filtered"
          :key="o.id"
          :order="o"
          mode="ongoing"
          class="order-item"
          @open-chat="handleOpenChat"
        >
          <template #actions>
            <div class="action-buttons">
              <!-- 状态操作按钮 -->
              <el-button v-if="o.status === 3" type="primary" size="large" @click="pickup(o.id)">
                <i class="iconfont icon-pickup"></i>
                确认取货
              </el-button>
              <div v-else class="deliver-action-container">
                <el-button
                  type="success"
                  size="large"
                  @click="deliver(o.id)"
                  :loading="deliveringId === o.id"
                  :disabled="deliveringId !== null && deliveringId !== o.id"
                >
                  <i class="iconfont icon-deliver"></i>
                  {{ deliveringId === o.id ? '正在确认送达...' : '确认送达' }}
                </el-button>
                <div class="delivery-tip">
                  <i class="iconfont icon-location"></i>
                  <span>送达需在收货点附近</span>
                </div>
              </div>

              <!-- 查看位置按钮 -->
              <el-button
                v-if="o.status === 3"
                type="info"
                size="large"
                @click="showMerchantLocation(o)"
                class="nav-button"
              >
                <i class="iconfont icon-map"></i>
                查看商家位置
              </el-button>
              <el-button
                v-else
                type="info"
                size="large"
                @click="showUserLocation(o)"
                class="nav-button"
              >
                <i class="iconfont icon-map"></i>
                查看用户位置
              </el-button>
            </div>
          </template>
        </RiderOrderCard>
      </TransitionGroup>
    </div>

    <!-- 空状态 -->
    <div v-else class="empty-state">
      <el-empty :description="tab === '3' ? '暂无待取餐订单' : '暂无派送中订单'" :image-size="180">
        <el-button type="primary" @click="load">刷新页面</el-button>
      </el-empty>
    </div>

    <!-- 地图弹窗 -->
    <AmapModal
      v-model="showMapModal"
      :merchant-data="currentOrderData?.status === 3 ? {
        title: currentOrderData?.restaurant || '商家',
        address: currentOrderData?.pickupAddress || '',
        type: 'merchant'
      } : undefined"
      :user-data="currentOrderData?.status === 4 ? {
        title: currentOrderData?.customer || '用户',
        address: currentOrderData?.deliveryAddress || '',
        type: 'user'
      } : undefined"
      :default-location="mapDefaultLocation"
      :initial-location-type="currentOrderData?.status === 3 ? 'merchant' : 'user'"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { ElMessage } from "element-plus";
import { riderApi, type RiderOrderItem } from "@/api/rider";
import RiderOrderCard from "@/components/rider/RiderOrderCard.vue";
import AmapModal from "@/components/AmapModal.vue";

const list = ref<RiderOrderItem[]>([]);
const loading = ref(false);
const tab = ref<"3" | "4">("3");

// 地图弹窗相关
const showMapModal = ref(false);
const mapDefaultLocation = ref<[number, number]>([113.299, 23.099]);
const currentOrderData = ref<any>(null);

// 送达按钮loading状态
const deliveringId = ref<number | null>(null);

// 定位状态相关
const isLocationTracking = ref(false);
const locationError = ref<string | null>(null);

const locationStatusIcon = computed(() => {
  if (locationError.value) return "icon-location-error";
  if (isLocationTracking.value) return "icon-location-on";
  return "icon-location-off";
});

const locationStatusText = computed(() => {
  if (locationError.value) return "定位异常";
  if (isLocationTracking.value) return "定位正常";
  return "未定位";
});

const count3 = computed(() => list.value.filter(x => x.status === 3).length);
const count4 = computed(() => list.value.filter(x => x.status === 4).length);

const filtered = computed(() => {
  const st = Number(tab.value);
  return list.value.filter(x => x.status === st);
});

const load = async () => {
  loading.value = true;
  try {
    const res = await riderApi.getOngoing();
    list.value = res.data.data || [];

    console.log('📦 === 订单数据加载完成 ===');
    console.log('📋 订单总数:', list.value.length);
    console.log('📊 待取餐订单数 (status=3):', count3.value);
    console.log('📊 派送中订单数 (status=4):', count4.value);

    // 详细输出每个订单的地址信息
    list.value.forEach((order, index) => {
      console.log(`📍 订单${index + 1} [ID:${order.id}] 地址信息:`);
      console.log(`  🏪 商家: ${order.restaurant}`);
      console.log(`  📮 pickupAddress: "${order.pickupAddress}" (长度: ${order.pickupAddress?.length || 0})`);
      console.log(`  🏠 客户: ${order.customer}`);
      console.log(`  📍 deliveryAddress: "${order.deliveryAddress}" (长度: ${order.deliveryAddress?.length || 0})`);
      console.log(`  📋 状态: ${order.status} (${order.status === 3 ? '待取餐' : '派送中'})`);
    });
    console.log('=====================================');

    // 自动切到有数据的 tab
    if (tab.value === "3" && count3.value === 0 && count4.value > 0) tab.value = "4";
    if (tab.value === "4" && count4.value === 0 && count3.value > 0) tab.value = "3";
  } finally {
    loading.value = false;
  }
};

const pickup = async (id: number) => {
  try {
    await riderApi.pickupOrder(id);
    ElMessage.success("已取货，进入派送中");
    await load();
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.msg || "取货失败");
  }
};

const deliver = async (id: number) => {
  // 设置对应按钮的loading状态
  deliveringId.value = id;

  // 🚨 确认送达调试信息
  const currentOrder = list.value.find(o => o.id === id);
  console.log('🚨 [确认送达调试] 点击送达按钮:', {
    订单ID: id,
    订单状态: currentOrder?.status,
    商家名称: currentOrder?.restaurant,
    商家地址: currentOrder?.pickupAddress,
    收货地址: currentOrder?.deliveryAddress,
    操作: '准备调用后端 deliverOrder API'
  });

  try {
    await riderApi.deliverOrder(id);
    ElMessage.success("🎉 送达成功！感谢您的配送服务");
    await load();
  } catch (e: any) {
    const errorMsg = e?.response?.data?.msg || "送达失败";

    console.error("❌ 送达失败详情:", {
      订单ID: id,
      错误消息: errorMsg,
      响应状态: e?.response?.status,
      完整错误: e
    });

    // 处理不同类型的失败情况
    if (errorMsg.includes("未获取到骑手当前位置")) {
      // 显示主要错误信息
      ElMessage.error({
        message: "❌ 无法获取您的位置",
        duration: 4000,
        showClose: true
      });

      // 延迟显示详细的引导文案
      setTimeout(() => {
        ElMessage({
          type: "warning",
          message: "💡 请先打开浏览器定位权限，然后点击右上角定位状态刷新位置",
          duration: 6000,
          showClose: true,
          dangerouslyUseHTMLString: true
        });
      }, 1000);

      // 再次延迟显示操作指引
      setTimeout(() => {
        ElMessage({
          type: "info",
          message: "📍 步骤：1.允许定位 → 2.刷新位置 → 3.重新尝试送达",
          duration: 5000,
          showClose: true
        });
      }, 3000);

    } else if (errorMsg.includes("不在收货点附近") || errorMsg.includes("距离约")) {
      // 解析距离信息
      const distanceMatch = errorMsg.match(/距离约\s*(\d+)\s*米/);
      const distance = distanceMatch ? distanceMatch[1] : "未知";

      ElMessage.error({
        message: `❌ 距离收货点过远（约 ${distance} 米）`,
        duration: 4000,
        showClose: true
      });

      setTimeout(() => {
        ElMessage({
          type: "warning",
          message: `💡 请导航至1公里范围内的收货点，然后重试送达`,
          duration: 5000,
          showClose: true
        });
      }, 1500);

      // 建议使用地图功能
      setTimeout(() => {
        ElMessage({
          type: "info",
          message: "🗺️ 您可以点击'查看用户位置'按钮查看具体收货地点",
          duration: 5000,
          showClose: true
        });
      }, 3000);

    } else if (errorMsg.includes("无法解析收货地址坐标")) {
      ElMessage.error({
        message: "❌ 收货地址解析失败",
        duration: 4000,
        showClose: true
      });

      const address = errorMsg.includes("：") ? errorMsg.split("：")[1] : "用户地址";

      setTimeout(() => {
        ElMessage({
          type: "warning",
          message: `⚠️ 地址: ${address}`,
          duration: 6000,
          showClose: true
        });
      }, 1500);

      setTimeout(() => {
        ElMessage({
          type: "info",
          message: "📞 如遇问题，请联系客服处理地址信息",
          duration: 5000,
          showClose: true
        });
      }, 3500);

    } else if (errorMsg.includes("骑手位置异常") || errorMsg.includes("位置异常")) {
      ElMessage.error({
        message: "❌ 您的位置异常，请重新获取定位后再试",
        duration: 4000,
        showClose: true
      });

      setTimeout(() => {
        ElMessage({
          type: "warning",
          message: "💡 请确保在珠海地区并允许浏览器定位权限",
          duration: 5000,
          showClose: true
        });
      }, 1500);

    } else if (errorMsg.includes("订单状态不允许")) {
      ElMessage.error({
        message: "❌ 当前订单状态无法送达",
        duration: 3000,
        showClose: true
      });

    } else if (errorMsg.includes("不属于你")) {
      ElMessage.error({
        message: "❌ 您不是此订单的配送员",
        duration: 3000,
        showClose: true
      });

    } else {
      // 其他未知错误
      ElMessage.error({
        message: `❌ 送达失败：${errorMsg}`,
        duration: 4000,
        showClose: true
      });

      setTimeout(() => {
        ElMessage({
          type: "info",
          message: "💡 如问题持续存在，请刷新页面或联系技术支持",
          duration: 5000,
          showClose: true
        });
      }, 2000);
    }

  } finally {
    // 无论成功还是失败，都清除loading状态，保持按钮可重复点击
    deliveringId.value = null;
  }
};

// 处理打开聊天事件
const handleOpenChat = (data: { type: 'user' | 'merchant'; id: number; name: string }) => {
  // 发送全局事件，聊天组件会监听这个事件
  window.dispatchEvent(new CustomEvent('rider:openChat', { detail: data }));
};

// 显示商家位置
const showMerchantLocation = (order: RiderOrderItem) => {
  console.log('📍 === 商家位置调试信息 ===');
  console.log('📋 订单ID:', order.id);
  console.log('🏪 商家名称:', order.restaurant);
  console.log('📮 商家地址 (pickupAddress):', order.pickupAddress);
  console.log('📊 地址长度:', order.pickupAddress?.length || 0);
  console.log('🏠 客户名称:', order.customer);
  console.log('📍 配送地址 (deliveryAddress):', order.deliveryAddress);
  console.log('📊 配送地址长度:', order.deliveryAddress?.length || 0);
  console.log('📋 订单状态:', order.status);
  console.log('🏷️ 地址来源说明: 商家使用 pickupAddress 字段');
  console.log('🏷️ 地址来源说明: 用户使用 deliveryAddress 字段');
  console.log('=====================================');

  currentOrderData.value = order;
  showMapModal.value = true;
};

// 显示用户位置
const showUserLocation = (order: RiderOrderItem) => {
  console.log('📍 === 用户位置调试信息 ===');
  console.log('📋 订单ID:', order.id);
  console.log('🏪 商家名称:', order.restaurant);
  console.log('📮 商家地址 (pickupAddress):', order.pickupAddress);
  console.log('📊 商家地址长度:', order.pickupAddress?.length || 0);
  console.log('🏠 客户名称:', order.customer);
  console.log('📍 配送地址 (deliveryAddress):', order.deliveryAddress);
  console.log('📊 配送地址长度:', order.deliveryAddress?.length || 0);
  console.log('📋 订单状态:', order.status);
  console.log('🏷️ 地址来源说明: 商家使用 pickupAddress 字段');
  console.log('🏷️ 地址来源说明: 用户使用 deliveryAddress 字段');
  console.log('=====================================');

  currentOrderData.value = order;
  showMapModal.value = true;
};

// 监听定位状态变化
const setupLocationListener = () => {
  // 监听定位状态
  window.addEventListener('rider:locationStatus', (event: any) => {
    isLocationTracking.value = event.detail.isTracking;
    locationError.value = event.detail.error || null;

    // 更新默认位置
    if (event.detail.location) {
      mapDefaultLocation.value = [event.detail.location.longitude || event.detail.location.lng, event.detail.location.latitude || event.detail.location.lat];
    }
  });

  // 监听位置更新
  window.addEventListener('rider:locationUpdate', (event: any) => {
    if (event.detail.location) {
      mapDefaultLocation.value = [event.detail.location.longitude || event.detail.location.lng, event.detail.location.latitude || event.detail.location.lat];
    }
  });
};

onMounted(() => {
  load();
  setupLocationListener();
});
</script>

<style scoped lang="scss">
.orders-page {
  padding: 24px;
  background: var(--rider-bg);
  min-height: calc(100vh - 60px);
}

// 页面头部
.page-header {
  background: linear-gradient(135deg, var(--rider-primary) 0%, var(--rider-primary-dark) 100%);
  border-radius: var(--rider-radius);
  padding: 30px;
  margin-bottom: 24px;
  box-shadow: var(--rider-shadow);
  color: #fff;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 24px;
}

.title-section {
  .page-title {
    font-size: 28px;
    font-weight: 800;
    margin: 0 0 8px 0;
    display: flex;
    align-items: center;
    gap: 12px;

    .title-icon {
      font-size: 32px;
    }
  }

  .page-subtitle {
    font-size: 16px;
    opacity: 0.9;
    margin: 0;
  }
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 16px;

  .location-status-info {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 16px;
    background: rgba(255, 255, 255, 0.15);
    border: 1px solid rgba(255, 255, 255, 0.3);
    border-radius: 20px;
    font-size: 14px;
    font-weight: 600;
    color: rgba(255, 255, 255, 0.9);

    .iconfont {
      font-size: 16px;
    }

    &.tracking {
      background: rgba(103, 194, 58, 0.2);
      border-color: rgba(103, 194, 58, 0.4);
      color: #fff;
    }

    &.error {
      background: rgba(245, 108, 108, 0.2);
      border-color: rgba(245, 108, 108, 0.4);
      color: #fff;
      animation: pulse 2s infinite;
    }
  }

  @keyframes pulse {
    0% { opacity: 1; }
    50% { opacity: 0.6; }
    100% { opacity: 1; }
  }

  :deep(.el-button) {
    background: rgba(255, 255, 255, 0.2);
    border-color: rgba(255, 255, 255, 0.3);
    color: #fff;
    padding: 12px 24px;
    font-size: 15px;
    font-weight: 600;

    &:hover {
      background: rgba(255, 255, 255, 0.3);
      border-color: rgba(255, 255, 255, 0.5);
    }
  }
}

// 统计卡片
.stats-container {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 20px;
  margin-bottom: 24px;
}

.stat-card {
  background: #fff;
  border: 2px solid var(--rider-border);
  border-radius: var(--rider-radius);
  padding: 24px;
  display: flex;
  align-items: center;
  gap: 20px;
  transition: all 0.3s ease;
  box-shadow: var(--rider-shadow);
  cursor: pointer;
  position: relative;
  overflow: hidden;

  &:hover {
    box-shadow: var(--rider-shadow-hover);
    transform: translateY(-2px);
  }

  &.active {
    border-color: var(--rider-primary);
    background: linear-gradient(135deg, rgba(255, 179, 2, 0.05) 0%, rgba(255, 179, 2, 0.02) 100%);

    .stat-icon {
      background: var(--rider-primary);
    }
  }

  .stat-icon {
    font-size: 36px;
    width: 70px;
    height: 70px;
    background: var(--rider-primary-light);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    transition: all 0.3s ease;
  }

  .stat-content {
    flex: 1;

    .stat-value {
      font-size: 32px;
      font-weight: 800;
      color: var(--rider-text);
      margin-bottom: 4px;
    }

    .stat-label {
      font-size: 16px;
      color: var(--rider-text);
      font-weight: 600;
      margin-bottom: 4px;
    }

    .stat-desc {
      font-size: 13px;
      color: var(--rider-sub);
    }
  }

  .stat-arrow {
    font-size: 20px;
    color: var(--rider-primary);
    opacity: 0.8;
  }
}

// 标签页
.status-tabs {
  background: #fff;
  border-radius: var(--rider-radius);
  padding: 20px;
  margin-bottom: 24px;
  box-shadow: var(--rider-shadow);

  :deep(.el-tabs__header) {
    margin: 0;
  }

  :deep(.el-tabs__nav-wrap) {
    &::after {
      display: none;
    }
  }

  :deep(.el-tabs__item) {
    padding: 0 32px;
    font-size: 16px;
    font-weight: 600;
    color: var(--rider-sub);

    &.is-active {
      color: var(--rider-primary);
    }
  }

  :deep(.el-tabs__active-bar) {
    background-color: var(--rider-primary);
    height: 4px;
    border-radius: 2px;
  }
}

.tab-label {
  display: flex;
  align-items: center;
  gap: 8px;

  .tab-icon {
    font-size: 18px;
  }

  .tab-badge {
    background: var(--rider-primary);
    color: #fff;
    font-size: 12px;
    padding: 2px 8px;
    border-radius: 10px;
    font-weight: 600;
    min-width: 20px;
    text-align: center;
  }
}

// 加载状态
.loading-container {
  background: #fff;
  border-radius: var(--rider-radius);
  padding: 30px;
  box-shadow: var(--rider-shadow);
}

// 订单网格
.orders-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(600px, 1fr));
  gap: 20px;
  align-items: start;
}

// 空状态
.empty-state {
  background: #fff;
  border-radius: var(--rider-radius);
  padding: 60px;
  text-align: center;
  box-shadow: var(--rider-shadow);
}

// 过渡动画
.order-list-enter-active,
.order-list-leave-active {
  transition: all 0.5s ease;
}

.order-list-enter-from {
  opacity: 0;
  transform: translateY(20px);
}

.order-list-leave-to {
  opacity: 0;
  transform: scale(0.95);
}

.order-list-move {
  transition: transform 0.5s ease;
}

// 响应式
@media (max-width: 768px) {
  .orders-page {
    padding: 16px;
  }

  .header-content {
    flex-direction: column;
    align-items: flex-start;
  }

  .page-header {
    padding: 20px;
  }

  .page-title {
    font-size: 24px !important;
  }

  .page-subtitle {
    font-size: 14px !important;
  }

  .stats-container {
    grid-template-columns: 1fr;
  }

  .status-tabs {
    padding: 16px;
  }

  .orders-grid {
    grid-template-columns: 1fr;
  }
}

// Icon font styles
.iconfont {
  font-family: "iconfont" !important;
  font-size: 14px;
  font-style: normal;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

.icon-refresh:before { content: "🔄"; }
.icon-pickup:before { content: "📦"; }
.icon-deliver:before { content: "✅"; }
.icon-location:before { content: "📍"; }
.icon-nav:before { content: "🧭"; }
.icon-map:before { content: "🗺️"; }
.icon-center:before { content: "🎯"; }
.icon-merchant:before { content: "🏪"; }
.icon-user:before { content: "👤"; }

// 操作按钮容器
.action-buttons {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
}

// 送达按钮容器
.deliver-action-container {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 8px;
  width: 100%;
}

// 送达提示
.delivery-tip {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 12px;
  color: var(--rider-sub);
  background: var(--rider-primary-light);
  padding: 4px 12px;
  border-radius: 12px;

  .iconfont {
    font-size: 12px;
  }
}

// 导航按钮
.nav-button {
  background: linear-gradient(135deg, #409EFF 0%, #66B1FF 100%);
  border-color: #409EFF;

  &:hover {
    background: linear-gradient(135deg, #337ECC 0%, #5DA3FF 100%);
    border-color: #337ECC;
  }
}
</style>
