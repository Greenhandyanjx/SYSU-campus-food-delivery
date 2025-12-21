<template>
  <el-dialog
    v-model="visible"
    :title="modalTitle"
    width="90%"
    max-width="900px"
    destroy-on-close
    @close="handleClose"
    class="amap-modal"
  >
    <div class="amap-container">
      <div id="amap" class="amap-wrapper"></div>
      <div v-if="loading" class="loading-overlay">
        <div class="loading-content">
          <el-icon class="loading-icon">
            <Loading />
          </el-icon>
          <span class="loading-text">{{ loadingText }}</span>
        </div>
      </div>
      <!-- 位置类型切换按钮 -->
      <div v-if="showLocationSwitch" class="location-switch">
        <el-button-group>
          <el-button
            :type="activeLocationType === 'merchant' ? 'primary' : ''"
            @click="switchLocationType('merchant')"
            size="small"
          >
            <i class="iconfont icon-merchant"></i>
            商家位置
          </el-button>
          <el-button
            :type="activeLocationType === 'user' ? 'primary' : ''"
            @click="switchLocationType('user')"
            size="small"
          >
            <i class="iconfont icon-user"></i>
            用户位置
          </el-button>
        </el-button-group>
      </div>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">关闭</el-button>
        <el-button v-if="hasValidLocation" type="info" @click="centerMap">
          <i class="iconfont icon-center"></i>
          重新定位
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, onUnmounted, computed } from 'vue';
import { Loading } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';
import amapLoader from '@/utils/amap';

interface LocationData {
  title: string;
  address: string;
  type: 'merchant' | 'user';
}

interface RiderLocationData {
  lng: number;
  lat: number;
  accuracy?: number;
}

interface Props {
  modelValue: boolean;
  merchantData?: LocationData;
  userData?: LocationData;
  defaultLocation?: [number, number]; // [lng, lat]
  initialLocationType?: 'merchant' | 'user';
  showRiderLocation?: boolean; // 是否显示骑手位置
  riderLocation?: RiderLocationData; // 骑手位置数据
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
  defaultLocation: () => [113.299, 23.099], // 中山大学默认位置
  initialLocationType: 'merchant',
  showRiderLocation: false
});

const emit = defineEmits<{
  'update:modelValue': [value: boolean];
}>();

const visible = ref(false);
const loading = ref(false);
const loadingText = ref('正在加载地图...');
const activeLocationType = ref<'merchant' | 'user'>(props.initialLocationType);

// 计算属性
const modalTitle = computed(() => {
  // 如果显示骑手位置
  if (props.showRiderLocation) {
    return '骑手当前位置';
  }

  const data = currentLocationData.value;
  return data ? `${data.type === 'merchant' ? '商家' : '用户'}位置：${data.title}` : '查看位置';
});

const showLocationSwitch = computed(() => {
  return props.merchantData && props.userData;
});

const currentLocationData = computed(() => {
  return activeLocationType.value === 'merchant' ? props.merchantData : props.userData;
});

const hasValidLocation = computed(() => {
  const data = currentLocationData.value;
  return data && data.address && data.address.trim();
});

// 地图相关变量
let AMap: any = null;
let map: any = null;
let geocoder: any = null;
let currentMarker: any = null;

// 地址智能补全函数
const enhanceAddress = (address: string, type: 'merchant' | 'user'): string => {
  if (!address || address.trim() === '') {
    return '';
  }

  const originalAddress = address.trim();

  // 如果地址太简单，尝试智能补全
  if (address.length < 5) {
    console.warn(`⚠️ 地址过于简单: "${originalAddress}"，尝试智能补全`);

    if (type === 'merchant') {
      // 商家地址补全 - 中山大学珠海校区常见地点
      const campusLocations = {
        '容园': '广东省珠海市香洲区中山大学珠海校区榕园',
        '榕园': '广东省珠海市香洲区中山大学珠海校区榕园',
        '荔园': '广东省珠海市香洲区中山大学珠海校区荔园',
        '食堂': '广东省珠海市香洲区中山大学珠海校区食堂',
        '宿舍': '广东省珠海市香洲区中山大学珠海校区学生宿舍',
        '教学楼': '广东省珠海市香洲区中山大学珠海校区教学楼',
        '图书馆': '广东省珠海市香洲区中山大学珠海校区图书馆',
        '超市': '广东省珠海市香洲区中山大学珠海校区超市'
      };

      // 尝试模糊匹配关键词
      for (const [key, location] of Object.entries(campusLocations)) {
        if (originalAddress.includes(key) || key.includes(originalAddress)) {
          console.log(`✅ 智能匹配: "${originalAddress}" -> "${location}"`);
          return location;
        }
      }

      // 处理数字地址（可能是楼号、宿舍号等）
      if (/^\d+$/.test(originalAddress)) {
        const enhancedAddress = `广东省珠海市香洲区中山大学珠海校区${originalAddress}栋`;
        console.log(`✅ 数字地址补全: "${originalAddress}" -> "${enhancedAddress}"`);
        return enhancedAddress;
      }

      // 处理"容9"这类格式（数字+文字或文字+数字）
      const containsNumber = /\d/.test(originalAddress);
      if (containsNumber) {
        const enhancedAddress = `广东省珠海市香洲区中山大学珠海校区${originalAddress}`;
        console.log(`✅ 楼栋地址补全: "${originalAddress}" -> "${enhancedAddress}"`);
        return enhancedAddress;
      }

      // 默认补全到中山大学珠海校区
      const defaultEnhanced = '广东省珠海市香洲区中山大学珠海校区';
      console.log(`⚠️ 默认补全: "${originalAddress}" -> "${defaultEnhanced}"`);
      return defaultEnhanced;
    } else {
      // 用户地址通常比较完整，只做基本处理
      if (address.length < 10) {
        const enhancedAddress = `广东省珠海市香洲区${originalAddress}`;
        console.log(`✅ 用户地址补全: "${originalAddress}" -> "${enhancedAddress}"`);
        return enhancedAddress;
      }
    }
  }

  return originalAddress;
};

// 初始化地图
const initMap = async () => {
  await nextTick();

  const mapContainer = document.getElementById('amap');
  if (!mapContainer) {
    console.error('Map container not found');
    return;
  }

  loading.value = true;
  loadingText.value = '正在加载地图...';

  try {
    // 加载高德地图 SDK
    AMap = await amapLoader.load({
      plugins: ['AMap.Geocoder', 'AMap.Marker', 'AMap.InfoWindow', 'AMap.ToolBar']
    });

    console.log('Initializing map with location data:', currentLocationData.value);
    console.log('Available AMap plugins:', Object.keys(AMap));

    // 创建地图实例
    map = new AMap.Map('amap', {
      zoom: 16,
      center: props.defaultLocation,
      viewMode: '2D',
      resizeEnable: true
    });

    // 等待地图完全加载
    map.on('complete', () => {
      console.log('Map completed loading');
      // 地图加载完成后立即显示位置
      showCurrentLocation();
    });

    console.log('Map initialized successfully');
  } catch (error) {
    console.error('地图初始化失败:', error);
    loadingText.value = '地图加载失败';
    loading.value = false;
  }
};

// 显示当前位置
const showCurrentLocation = () => {
  // 如果显示骑手位置
  if (props.showRiderLocation) {
    showRiderLocationMarker();
    return;
  }

  const data = currentLocationData.value;
  if (!data || !data.address || !data.address.trim()) {
    console.log('No valid address to display');
    loading.value = false;
    return;
  }

  loadingText.value = '正在解析地址...';
  geocodeAndShowMarker(data).finally(() => {
    loading.value = false;
  });
};

// 地址解析并显示标记
const geocodeAndShowMarker = (data: any) => {
  return new Promise<void>((resolve, reject) => {
    if (!AMap || !map) {
      reject(new Error('AMap or map not loaded'));
      return;
    }

    console.log('=== 地址解析调试信息 ===');
    console.log('AMap.Geocoder available:', typeof AMap.Geocoder);
    console.log('原始地址数据:', {
      title: data.title,
      address: data.address,
      type: data.type,
      addressLength: data.address?.length || 0
    });

    // 智能补全地址
    const enhancedAddress = enhanceAddress(data.address, data.type);
    console.log('增强后地址:', enhancedAddress);

    // 检查地址是否为空
    if (!enhancedAddress) {
      console.warn('地址为空，显示默认位置');
      showDefaultMarker(data);
      resolve();
      return;
    }

    // 更新数据对象中的地址
    const enhancedData = { ...data, address: enhancedAddress };

    // 动态加载地理编码器插件
    const loadGeocoder = () => {
      return new Promise<void>((resolve, reject) => {
        if (typeof AMap.Geocoder === 'function') {
          console.log('✅ AMap.Geocoder 已可用');
          resolve();
          return;
        }

        console.log('🔄 动态加载 AMap.Geocoder 插件...');
        (window as any).AMap.plugin(['AMap.Geocoder'], () => {
          console.log('✅ AMap.Geocoder 插件加载成功');
          resolve();
        });
      });
    };

    // 等待插件加载完成
    loadGeocoder().then(() => {
      // 初始化地理编码器
      if (!geocoder) {
        geocoder = new AMap.Geocoder({
          city: '珠海',
          radius: 1000,
          extensions: 'base'
        });
        console.log('✅ 地理编码器初始化完成');
      }

      // 清除之前的标记
      if (currentMarker) {
        map.remove(currentMarker);
        currentMarker = null;
      }

      console.log('开始地址解析:', enhancedAddress);

      geocoder.getLocation(enhancedAddress, (status: string, result: any) => {
        console.log('地理编码结果:', {
          status: status,
          result: result,
          hasGeocodes: result?.geocodes?.length || 0
        });

        if (status === 'complete' && result.geocodes && result.geocodes.length > 0) {
          const location = result.geocodes[0].location;
          // 使用更精确的经纬度（8位小数用于显示，原始值用于地图定位）
          const preciseLng = Number(location.lng).toFixed(8);
          const preciseLat = Number(location.lat).toFixed(8);
          const preciseLocation = [Number(location.lng), Number(location.lat)];

          console.log('✅ 地址解析成功:', {
            原地址: data.address,
            增强地址: enhancedAddress,
            经度: preciseLng,
            纬度: preciseLat,
            解析级别: result.geocodes[0].level,
            匹配度: result.geocodes[0].confidence
          });

          // 更新地图中心点 - 使用更精确的经纬度
          map.setCenter(preciseLocation);
          map.setZoom(18); // 提高缩放级别以获得更精确的视图

          // 创建标记图标
          const markerIcon = data.type === 'merchant'
            ? '🏪' // 商家图标
            : '🏠'; // 用户图标

          // 添加标记 - 使用精确的经纬度
          currentMarker = new AMap.Marker({
            position: preciseLocation,
            title: data.title,
            animation: "AMAP_ANIMATION_DROP",
            content: `<div style="
              background: ${data.type === 'merchant' ? '#ff6b6b' : '#4ecdc4'};
              color: white;
              width: 40px;
              height: 40px;
              border-radius: 50%;
              display: flex;
              align-items: center;
              justify-content: center;
              font-size: 20px;
              border: 3px solid white;
              box-shadow: 0 2px 8px rgba(0,0,0,0.3);
            ">${markerIcon}</div>`
          });
          map.add(currentMarker);

          // 添加信息窗体
          const infoWindow = new AMap.InfoWindow({
            content: `<div style="padding: 12px; max-width: 250px; word-wrap: break-word;">
              <div style="font-weight: bold; margin-bottom: 8px; font-size: 16px;">
                ${data.type === 'merchant' ? '🏪 商家' : '🏠 用户'}：${data.title}
              </div>
              <div style="color: #666; font-size: 14px; line-height: 1.4;">
                ${enhancedAddress}
              </div>
              ${enhancedAddress !== data.address ? `<div style="color: #ff6b6b; font-size: 12px; margin-top: 4px;">📍 原地址: ${data.address}</div>` : ''}
              <div style="color: #999; font-size: 12px; margin-top: 8px;">
                📍 精确坐标: ${preciseLng}, ${preciseLat}
              </div>
            </div>`,
            offset: new AMap.Pixel(0, -30)
          });

          currentMarker.on('click', () => {
            infoWindow.open(map, preciseLocation);
          });

          // 自动打开信息窗体 - 使用精确的经纬度
          setTimeout(() => {
            if (map && infoWindow) {
              infoWindow.open(map, preciseLocation);
            }
          }, 500);

          resolve();
        } else {
          console.error('❌ 地址解析失败:', {
            地址: enhancedAddress,
            状态: status,
            错误信息: result?.info || '未知错误',
            完整结果: result
          });

          // 地址解析失败，显示默认位置
          showDefaultMarker(enhancedData);
          resolve();
        }
      });
    }).catch((error) => {
      console.error('❌ 地理编码器插件加载失败:', error);
      reject(error);
    });
  });
};

// 显示默认位置标记
const showDefaultMarker = (data: any) => {
  if (!map) return;

  console.log('⚠️ 显示默认位置:', {
    title: data.title,
    address: data.address,
    type: data.type,
    defaultLocation: props.defaultLocation
  });

  // 清除之前的标记
  if (currentMarker) {
    map.remove(currentMarker);
  }

  const markerIcon = data.type === 'merchant' ? '🏪' : '🏠';

  currentMarker = new AMap.Marker({
    position: props.defaultLocation,
    title: data.title,
    content: `<div style="
      background: ${data.type === 'merchant' ? '#ff6b6b' : '#4ecdc4'};
      color: white;
      width: 40px;
      height: 40px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 20px;
      border: 3px solid white;
      box-shadow: 0 2px 8px rgba(0,0,0,0.3);
    ">${markerIcon}</div>`
  });
  map.add(currentMarker);

  const infoWindow = new AMap.InfoWindow({
    content: `<div style="padding: 12px; max-width: 250px;">
      <div style="font-weight: bold; margin-bottom: 8px; font-size: 16px;">
        ${data.type === 'merchant' ? '🏪 商家' : '🏠 用户'}：${data.title}
      </div>
      <div style="color: #666; font-size: 14px; line-height: 1.4;">
        ${data.address || '默认位置：中山大学珠海校区'}
      </div>
      <div style="color: #ff6b6b; font-size: 12px; margin-top: 8px;">
        ⚠️ 地址解析失败，显示默认位置
      </div>
    </div>`,
    offset: new AMap.Pixel(0, -30)
  });

  currentMarker.on('click', () => {
    infoWindow.open(map, props.defaultLocation);
  });

  // 自动打开信息窗体
  setTimeout(() => {
    if (map && infoWindow) {
      infoWindow.open(map, props.defaultLocation);
    }
  }, 500);
};

// 切换位置类型
const switchLocationType = (type: 'merchant' | 'user') => {
  if (activeLocationType.value === type) return;

  activeLocationType.value = type;
  loading.value = true;
  loadingText.value = '正在切换位置...';

  nextTick(() => {
    showCurrentLocation();
  });
};

// 显示骑手位置标记
const showRiderLocationMarker = () => {
  if (!map) {
    console.warn('地图未初始化');
    loading.value = false;
    return;
  }

  // 验证骑手位置数据
  if (!props.riderLocation ||
      typeof props.riderLocation.lng !== 'number' ||
      typeof props.riderLocation.lat !== 'number' ||
      isNaN(props.riderLocation.lng) ||
      isNaN(props.riderLocation.lat) ||
      props.riderLocation.lng === 0 ||
      props.riderLocation.lat === 0) {
    console.warn('骑手位置数据无效:', props.riderLocation);
    loadingText.value = '骑手位置数据无效';

    // 显示默认位置或提示信息
    setTimeout(() => {
      loading.value = false;
      ElMessage.warning('⚠️ 骑手位置数据无效，请等待定位完成');
    }, 1000);
    return;
  }

  loadingText.value = '正在显示骑手位置...';

  console.log('📍 [骑手位置] 显示骑手当前位置:', {
    lng: props.riderLocation.lng,
    lat: props.riderLocation.lat,
    accuracy: props.riderLocation.accuracy
  });

  // 清除之前的标记
  if (currentMarker) {
    map.remove(currentMarker);
    currentMarker = null;
  }

  const riderPosition = [props.riderLocation.lng, props.riderLocation.lat];

  // 设置地图中心到骑手位置
  map.setCenter(riderPosition);
  map.setZoom(18);

  // 创建骑手标记
  currentMarker = new AMap.Marker({
    position: riderPosition,
    title: '骑手当前位置',
    animation: "AMAP_ANIMATION_DROP",
    content: `<div style="
      background: #4CAF50;
      color: white;
      width: 40px;
      height: 40px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 20px;
      border: 3px solid white;
      box-shadow: 0 2px 8px rgba(0,0,0,0.3);
    ">🏍️</div>`
  });
  map.add(currentMarker);

  // 添加信息窗体
  const infoWindow = new AMap.InfoWindow({
    content: `<div style="padding: 12px; max-width: 250px; word-wrap: break-word;">
      <div style="font-weight: bold; margin-bottom: 8px; font-size: 16px;">
        🏍️ 骑手当前位置
      </div>
      <div style="color: #666; font-size: 14px; line-height: 1.4;">
        经度: ${props.riderLocation.lng.toFixed(8)}<br>
        纬度: ${props.riderLocation.lat.toFixed(8)}
      </div>
      ${props.riderLocation.accuracy ? `<div style="color: #999; font-size: 12px; margin-top: 8px;">
        📍 定位精度: ${Math.round(props.riderLocation.accuracy)}米
      </div>` : ''}
      <div style="color: #4CAF50; font-size: 12px; margin-top: 8px;">
        ✅ 位置追踪正常
      </div>
    </div>`,
    offset: new AMap.Pixel(0, -30)
  });

  currentMarker.on('click', () => {
    infoWindow.open(map, riderPosition);
  });

  // 自动打开信息窗体
  setTimeout(() => {
    if (map && infoWindow) {
      infoWindow.open(map, riderPosition);
    }
    loading.value = false;
  }, 500);
};

// 重新定位到当前标记
const centerMap = () => {
  if (currentMarker && map) {
    const position = currentMarker.getPosition();
    map.setCenter([position.lng, position.lat]);
    map.setZoom(17);
  }
};

const handleClose = () => {
  visible.value = false;
  emit('update:modelValue', false);

  // 清理地图
  if (map) {
    map.destroy();
    map = null;
  }
  geocoder = null;
  currentMarker = null;
};

// 监听弹窗显示状态
watch(() => props.modelValue, (newVal) => {
  visible.value = newVal;
  if (newVal) {
    // 重置位置类型为初始值
    activeLocationType.value = props.initialLocationType;
    initMap();
  }
});

// 监听弹窗显示状态（双向绑定）
watch(visible, (newVal) => {
  emit('update:modelValue', newVal);
});

// 组件卸载时清理
onUnmounted(() => {
  if (map) {
    map.destroy();
    map = null;
  }
});
</script>

<style scoped lang="scss">
.amap-modal {
  :deep(.el-dialog) {
    border-radius: 12px;
    overflow: hidden;
  }

  :deep(.el-dialog__body) {
    padding: 0;
    position: relative;
  }
}

.amap-container {
  position: relative;
  height: 500px;
  width: 100%;
}

.amap-wrapper {
  height: 100%;
  width: 100%;
  background: #f5f5f5;
}

.location-switch {
  position: absolute;
  top: 16px;
  right: 16px;
  z-index: 1000;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 8px;
  padding: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(10px);

  :deep(.el-button-group) {
    .el-button {
      font-size: 12px;
      padding: 6px 12px;
      border-radius: 6px;
      font-weight: 600;

      &.el-button--primary {
        background: var(--el-color-primary);
        border-color: var(--el-color-primary);
        color: white;
      }
    }
  }
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.9);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.loading-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.loading-icon {
  font-size: 28px;
  color: var(--el-color-primary);
  animation: rotating 2s linear infinite;
}

.loading-text {
  font-size: 14px;
  color: var(--el-color-primary);
  font-weight: 600;
}

@keyframes rotating {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.dialog-footer {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  padding: 16px 20px;
  border-top: 1px solid #ebeef5;
}

:deep(.el-button) {
  border-radius: 20px;
  padding: 8px 20px;
  font-weight: 600;
}

// 移动端适配
@media (max-width: 768px) {
  .amap-container {
    height: 400px;
  }

  .dialog-footer {
    flex-direction: column;

    :deep(.el-button) {
      width: 100%;
    }
  }

  :deep(.el-dialog) {
    width: 95% !important;
    margin: 0 auto;
  }
}
</style>