<template>
  <div class="map-settings">
    <!-- 顶部导航栏 -->
    <div class="header-bar">
      <div class="back-btn" @click="$router.go(-1)">
        <i class="css-icon back"></i>
      </div>
      <h1 class="page-title">地图设置</h1>
      <div class="save-btn" @click="saveSettings">
        <span>保存</span>
      </div>
    </div>

    <!-- 地图偏好 -->
    <div class="preference-section">
      <h3 class="section-title">地图偏好</h3>

      <div class="map-provider">
        <div class="provider-label">地图提供商</div>
        <div class="provider-options">
          <div
            v-for="provider in mapProviders"
            :key="provider.id"
            class="provider-option"
            :class="{ active: mapSettings.provider === provider.id }"
            @click="selectMapProvider(provider.id)"
          >
            <div class="provider-icon">
              <i :class="provider.iconClass"></i>
            </div>
            <div class="provider-info">
              <div class="provider-name">{{ provider.name }}</div>
              <div class="provider-desc">{{ provider.description }}</div>
            </div>
            <div class="provider-radio">
              <div class="radio-circle" :class="{ checked: mapSettings.provider === provider.id }">
                <div class="radio-dot"></div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="map-display">
        <div class="display-label">地图显示</div>
        <div class="display-options">
          <div class="display-item">
            <div class="display-info">
              <div class="display-title">实时路况</div>
              <div class="display-desc">显示实时交通状况</div>
            </div>
            <el-switch
              v-model="mapSettings.realTimeTraffic"
              @change="updateSetting('realTimeTraffic')"
            />
          </div>

          <div class="display-item">
            <div class="display-info">
              <div class="display-title">卫星地图</div>
              <div class="display-desc">显示卫星影像地图</div>
            </div>
            <el-switch
              v-model="mapSettings.satelliteView"
              @change="updateSetting('satelliteView')"
            />
          </div>

          <div class="display-item">
            <div class="display-info">
              <div class="display-title">3D建筑</div>
              <div class="display-desc">显示3D建筑模型</div>
            </div>
            <el-switch
              v-model="mapSettings.show3DBuildings"
              @change="updateSetting('show3DBuildings')"
            />
          </div>
        </div>
      </div>
    </div>

    <!-- 导航设置 -->
    <div class="preference-section">
      <h3 class="section-title">导航设置</h3>

      <div class="navigation-mode">
        <div class="mode-label">导航模式</div>
        <div class="mode-options">
          <div
            v-for="mode in navigationModes"
            :key="mode.id"
            class="mode-option"
            :class="{ active: mapSettings.navigationMode === mode.id }"
            @click="selectNavigationMode(mode.id)"
          >
            <div class="mode-icon">
              <i :class="mode.iconClass"></i>
            </div>
            <div class="mode-content">
              <div class="mode-name">{{ mode.name }}</div>
              <div class="mode-desc">{{ mode.description }}</div>
            </div>
          </div>
        </div>
      </div>

      <div class="route-preference">
        <div class="preference-label">路线偏好</div>
        <div class="preference-options">
          <div class="preference-item">
            <div class="preference-info">
              <div class="preference-title">避免拥堵</div>
              <div class="preference-desc">优先选择畅通路线</div>
            </div>
            <el-switch
              v-model="mapSettings.avoidCongestion"
              @change="updateSetting('avoidCongestion')"
            />
          </div>

          <div class="preference-item">
            <div class="preference-info">
              <div class="preference-title">避免高速</div>
              <div class="preference-desc">避免高速公路收费</div>
            </div>
            <el-switch
              v-model="mapSettings.avoidHighway"
              @change="updateSetting('avoidHighway')"
            />
          </div>

          <div class="preference-item">
            <div class="preference-info">
              <div class="preference-title">避免收费站</div>
              <div class="preference-desc">选择免费路线</div>
            </div>
            <el-switch
              v-model="mapSettings.avoidToll"
              @change="updateSetting('avoidToll')"
            />
          </div>
        </div>
      </div>
    </div>

    <!-- 语音设置 -->
    <div class="preference-section">
      <h3 class="section-title">语音设置</h3>

      <div class="voice-settings">
        <div class="voice-item">
          <div class="voice-info">
            <div class="voice-title">语音导航</div>
            <div class="voice-desc">开启语音播报导航</div>
          </div>
          <el-switch
            v-model="mapSettings.voiceNavigation"
            @change="onVoiceNavigationToggle"
          />
        </div>

        <template v-if="mapSettings.voiceNavigation">
          <div class="voice-item">
            <div class="voice-label">语音包</div>
            <el-select v-model="mapSettings.voicePackage" placeholder="选择语音包">
              <el-option label="标准女声" value="female_standard" />
              <el-option label="标准男声" value="male_standard" />
              <el-option label="甜美女声" value="female_sweet" />
              <el-option label="磁性男声" value="male_magnetic" />
              <el-option label="方言语音" value="dialect" />
            </el-select>
          </div>

          <div class="voice-item">
            <div class="voice-label">语音音量</div>
            <div class="volume-control">
              <el-slider
                v-model="mapSettings.voiceVolume"
                :min="0"
                :max="100"
                show-input
                :show-input-controls="false"
                @change="updateSetting('voiceVolume')"
              />
            </div>
          </div>

          <div class="voice-item">
            <div class="voice-info">
              <div class="voice-title">详细播报</div>
              <div class="voice-desc">播报更多路况信息</div>
            </div>
            <el-switch
              v-model="mapSettings.detailVoice"
              @change="updateSetting('detailVoice')"
            />
          </div>

          <div class="voice-item">
            <div class="voice-info">
              <div class="voice-title">自动音量调节</div>
              <div class="voice-desc">根据环境噪音调节音量</div>
            </div>
            <el-switch
              v-model="mapSettings.autoVolume"
              @change="updateSetting('autoVolume')"
            />
          </div>
        </template>
      </div>
    </div>

    <!-- 定位设置 -->
    <div class="preference-section">
      <h3 class="section-title">定位设置</h3>

      <div class="location-settings">
        <div class="location-item">
          <div class="location-info">
            <div class="location-title">高精度定位</div>
            <div class="location-desc">使用GPS+基站+WiFi定位</div>
          </div>
          <el-switch
            v-model="mapSettings.highPrecisionLocation"
            @change="updateSetting('highPrecisionLocation')"
          />
        </div>

        <div class="location-item">
          <div class="location-info">
            <div class="location-title">位置上传频率</div>
            <div class="location-desc">设置位置信息上传间隔</div>
          </div>
          <el-select v-model="mapSettings.locationUploadInterval" placeholder="选择上传频率">
            <el-option label="实时上传" value="realtime" />
            <el-option label="5秒" value="5s" />
            <el-option label="10秒" value="10s" />
            <el-option label="30秒" value="30s" />
            <el-option label="1分钟" value="1m" />
          </el-select>
        </div>

        <div class="location-item">
          <div class="location-info">
            <div class="location-title">位置纠偏</div>
            <div class="location-desc">自动修正GPS偏移</div>
          </div>
          <el-switch
            v-model="mapSettings.locationCorrection"
            @change="updateSetting('locationCorrection')"
          />
        </div>

        <div class="location-item">
          <div class="location-info">
            <div class="location-title">显示坐标</div>
            <div class="location-desc">在地图上显示当前位置坐标</div>
          </div>
          <el-switch
            v-model="mapSettings.showCoordinates"
            @change="updateSetting('showCoordinates')"
          />
        </div>
      </div>
    </div>

    <!-- 离线地图 -->
    <div class="preference-section">
      <h3 class="section-title">离线地图</h3>

      <div class="offline-settings">
        <div class="offline-item">
          <div class="offline-info">
            <div class="offline-title">自动下载</div>
            <div class="offline-desc">在有WiFi时自动下载离线地图</div>
          </div>
          <el-switch
            v-model="mapSettings.autoDownloadOffline"
            @change="updateSetting('autoDownloadOffline')"
          />
        </div>

        <div class="offline-maps">
          <div class="offline-label">已下载地图</div>
          <div class="offline-list">
            <div v-for="map in offlineMaps" :key="map.id" class="offline-map-item">
              <div class="map-info">
                <div class="map-name">{{ map.name }}</div>
                <div class="map-size">{{ map.size }}MB</div>
                <div class="map-update">更新于 {{ map.lastUpdate }}</div>
              </div>
              <div class="map-actions">
                <el-button size="small" @click="updateOfflineMap(map.id)">更新</el-button>
                <el-button size="small" type="danger" @click="deleteOfflineMap(map.id)">删除</el-button>
              </div>
            </div>
          </div>

          <el-button class="download-btn" @click="downloadMoreMaps">
            <i class="css-icon download"></i>
            下载更多地图
          </el-button>
        </div>
      </div>
    </div>

    <!-- 地图缓存 -->
    <div class="preference-section">
      <h3 class="section-title">地图缓存</h3>

      <div class="cache-settings">
        <div class="cache-info">
          <div class="cache-stat">
            <div class="cache-size">{{ cacheSize }}MB</div>
            <div class="cache-desc">当前缓存大小</div>
          </div>
          <div class="cache-actions">
            <el-button @click="clearCache">清除缓存</el-button>
          </div>
        </div>

        <div class="cache-item">
          <div class="cache-info">
            <div class="cache-title">自动清理缓存</div>
            <div class="cache-desc">定期清理过期的地图缓存</div>
          </div>
          <el-switch
            v-model="mapSettings.autoClearCache"
            @change="updateSetting('autoClearCache')"
          />
        </div>

        <div class="cache-item">
          <div class="cache-label">缓存保留时间</div>
          <el-select v-model="mapSettings.cacheRetentionPeriod" placeholder="选择保留时间">
            <el-option label="1天" value="1d" />
            <el-option label="3天" value="3d" />
            <el-option label="7天" value="7d" />
            <el-option label="14天" value="14d" />
            <el-option label="30天" value="30d" />
          </el-select>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import riderApi from '@/api/rider'

const router = useRouter()

// 地图提供商
const mapProviders = [
  {
    id: 'amap',
    name: '高德地图',
    description: '数据准确，导航精准',
    iconClass: 'css-icon amap'
  },
  {
    id: 'baidu',
    name: '百度地图',
    description: '覆盖全面，功能丰富',
    iconClass: 'css-icon baidu'
  },
  {
    id: 'tencent',
    name: '腾讯地图',
    description: '界面简洁，响应迅速',
    iconClass: 'css-icon tencent'
  }
]

// 导航模式
const navigationModes = [
  {
    id: 'fastest',
    name: '最快路线',
    description: '优先选择时间最短的路线',
    iconClass: 'css-icon fastest'
  },
  {
    id: 'shortest',
    name: '最短路线',
    description: '优先选择距离最短的路线',
    iconClass: 'css-icon shortest'
  },
  {
    id: 'economy',
    name: '经济路线',
    description: '综合考虑时间和费用',
    iconClass: 'css-icon economy'
  }
]

// 地图设置
const mapSettings = reactive({
  provider: 'amap',
  realTimeTraffic: true,
  satelliteView: false,
  show3DBuildings: false,
  navigationMode: 'fastest',
  avoidCongestion: true,
  avoidHighway: false,
  avoidToll: false,
  voiceNavigation: true,
  voicePackage: 'female_standard',
  voiceVolume: 80,
  detailVoice: true,
  autoVolume: true,
  highPrecisionLocation: true,
  locationUploadInterval: '10s',
  locationCorrection: true,
  showCoordinates: false,
  autoDownloadOffline: true,
  autoClearCache: true,
  cacheRetentionPeriod: '7d'
})

// 离线地图
const offlineMaps = ref([
  {
    id: 1,
    name: '珠海市',
    size: 45.6,
    lastUpdate: '2024-01-15'
  },
  {
    id: 2,
    name: '广州市',
    size: 128.3,
    lastUpdate: '2024-01-10'
  }
])

// 缓存大小
const cacheSize = ref(256.8)

// 加载地图设置
const loadMapSettings = async () => {
  try {
    // 模拟API调用
    // const response = await riderApi.getMapSettings()

    // 使用默认设置
    console.log('使用默认地图设置')
  } catch (error) {
    console.error('加载地图设置失败:', error)
    ElMessage.warning('加载设置失败，使用默认配置')
  }
}

// 选择地图提供商
const selectMapProvider = (provider) => {
  mapSettings.provider = provider
}

// 选择导航模式
const selectNavigationMode = (mode) => {
  mapSettings.navigationMode = mode
}

// 更新设置
const updateSetting = (key) => {
  ElMessage.success('设置已更新')
}

// 语音导航开关
const onVoiceNavigationToggle = (enabled) => {
  if (enabled && !mapSettings.voicePackage) {
    mapSettings.voicePackage = 'female_standard'
  }
}

// 更新离线地图
const updateOfflineMap = (mapId) => {
  ElMessage.info('正在检查更新...')
  setTimeout(() => {
    ElMessage.success('地图已是最新版本')
  }, 2000)
}

// 删除离线地图
const deleteOfflineMap = (mapId) => {
  ElMessageBox.confirm(
    '确定要删除这个离线地图吗？',
    '提示',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => {
    const index = offlineMaps.value.findIndex(map => map.id === mapId)
    if (index > -1) {
      offlineMaps.value.splice(index, 1)
      ElMessage.success('离线地图已删除')
    }
  }).catch(() => {
    // 用户取消
  })
}

// 下载更多地图
const downloadMoreMaps = () => {
  ElMessage.info('地图下载功能开发中...')
}

// 清除缓存
const clearCache = () => {
  ElMessageBox.confirm(
    '确定要清除地图缓存吗？这将占用流量重新加载数据。',
    '提示',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => {
    cacheSize.value = 0
    ElMessage.success('缓存已清除')
  }).catch(() => {
    // 用户取消
  })
}

// 保存设置
const saveSettings = async () => {
  try {
    // 模拟API调用
    // const response = await riderApi.updateMapSettings(mapSettings)

    ElMessage.success('地图设置已保存')
    router.go(-1)
  } catch (error) {
    console.error('保存地图设置失败:', error)
    ElMessage.error('保存失败，请重试')
  }
}

onMounted(() => {
  loadMapSettings()
})
</script>

<style scoped>
/* CSS图标 */
.css-icon {
  display: inline-block;
  width: 1em;
  height: 1em;
  position: relative;
  font-size: inherit;
  color: inherit;
}

/* 返回图标 */
.css-icon.back::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-40%, -50%) rotate(-45deg);
  width: 10px;
  height: 10px;
  border-left: 2px solid currentColor;
  border-bottom: 2px solid currentColor;
}

/* 高德地图图标 */
.css-icon.amap::before {
  content: '高';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 16px;
  font-weight: bold;
  color: #1E8E3E;
}

/* 百度地图图标 */
.css-icon.baidu::before {
  content: '百';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 16px;
  font-weight: bold;
  color: #2932E1;
}

/* 腾讯地图图标 */
.css-icon.tencent::before {
  content: '腾';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 16px;
  font-weight: bold;
  color: #00D4AA;
}

/* 最快路线图标 */
.css-icon.fastest::before {
  content: '⚡';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 16px;
}

/* 最短路线图标 */
.css-icon.shortest::before {
  content: '📏';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 16px;
}

/* 经济路线图标 */
.css-icon.economy::before {
  content: '💰';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 16px;
}

/* 下载图标 */
.css-icon.download::before {
  content: '⬇';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 16px;
}

.map-settings {
  background: #f5f5f5;
  min-height: 100vh;
  padding-bottom: 80px;
}

/* 顶部导航栏 */
.header-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 15px;
  background: #FFD700;
  color: #333;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 100;
}

.back-btn, .save-btn {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.back-btn {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  cursor: pointer;
  transition: all 0.3s ease;
}

.back-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

.save-btn {
  width: auto;
  padding: 0 15px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 20px;
  cursor: pointer;
}

.back-btn .css-icon {
  font-size: 20px;
  color: #333;
}

.save-btn span {
  font-size: 14px;
  font-weight: 500;
  color: #333;
}

.page-title {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
}

/* 偏好设置区块 */
.preference-section {
  margin: 70px 15px 15px;
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.section-title {
  margin: 0 0 20px 0;
  font-size: 16px;
  font-weight: 500;
  color: #333;
}

/* 地图提供商 */
.provider-label,
.mode-label,
.preference-label,
.voice-label,
.location-label,
.offline-label {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin-bottom: 15px;
}

.provider-options,
.mode-options {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.provider-option,
.mode-option {
  display: flex;
  align-items: center;
  padding: 16px;
  border: 2px solid #f0f0f0;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.provider-option:hover,
.mode-option:hover {
  border-color: #FFD700;
}

.provider-option.active,
.mode-option.active {
  border-color: #FFD700;
  background: #fffbf0;
}

.provider-icon,
.mode-icon {
  width: 44px;
  height: 44px;
  background: #f8f9fa;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 15px;
}

.provider-icon .css-icon,
.mode-icon .css-icon {
  font-size: 24px;
}

.provider-info,
.mode-content {
  flex: 1;
}

.provider-name,
.mode-name {
  font-size: 16px;
  font-weight: 500;
  color: #333;
  margin-bottom: 4px;
}

.provider-desc,
.mode-desc {
  font-size: 12px;
  color: #999;
}

.provider-radio {
  display: flex;
  align-items: center;
}

.radio-circle {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: 2px solid #ddd;
  position: relative;
  transition: all 0.3s ease;
}

.radio-circle.checked {
  border-color: #FFD700;
}

.radio-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #FFD700;
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%) scale(0);
  transition: transform 0.3s ease;
}

.radio-circle.checked .radio-dot {
  transform: translate(-50%, -50%) scale(1);
}

/* 地图显示 */
.map-display {
  margin-top: 25px;
}

.display-label {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin-bottom: 15px;
}

.display-options {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.display-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.display-item:last-child {
  border-bottom: none;
}

.display-info {
  flex: 1;
}

.display-title {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin-bottom: 4px;
}

.display-desc {
  font-size: 12px;
  color: #999;
}

/* 路线偏好 */
.route-preference {
  margin-top: 25px;
}

.preference-options {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.preference-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.preference-item:last-child {
  border-bottom: none;
}

.preference-info {
  flex: 1;
}

.preference-title {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin-bottom: 4px;
}

.preference-desc {
  font-size: 12px;
  color: #999;
}

/* 语音设置 */
.voice-settings,
.location-settings,
.offline-settings,
.cache-settings {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.voice-item,
.location-item,
.cache-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.voice-item:last-child,
.location-item:last-child,
.cache-item:last-child {
  border-bottom: none;
}

.voice-info,
.location-info,
.cache-info {
  flex: 1;
}

.voice-title,
.location-title,
.cache-title {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin-bottom: 4px;
}

.voice-desc,
.location-desc,
.cache-desc {
  font-size: 12px;
  color: #999;
}

.volume-control {
  flex: 1;
  margin-left: 15px;
}

/* 离线地图 */
.offline-maps {
  margin-top: 20px;
}

.offline-list {
  margin-top: 15px;
}

.offline-map-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 8px;
  margin-bottom: 10px;
}

.map-info {
  flex: 1;
}

.map-name {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin-bottom: 4px;
}

.map-size,
.map-update {
  font-size: 12px;
  color: #999;
}

.map-actions {
  display: flex;
  gap: 8px;
}

.download-btn {
  width: 100%;
  margin-top: 15px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.download-btn .css-icon {
  font-size: 16px;
}

/* 缓存设置 */
.cache-info {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 8px;
  margin-bottom: 15px;
}

.cache-stat {
  text-align: center;
}

.cache-size {
  font-size: 24px;
  font-weight: bold;
  color: #333;
  margin-bottom: 4px;
}

.cache-desc {
  font-size: 12px;
  color: #999;
}

/* 自定义组件样式 */
:deep(.el-input__wrapper) {
  border-radius: 8px;
  padding: 8px 12px;
}

:deep(.el-select) {
  width: 180px;
}

:deep(.el-slider) {
  flex: 1;
}

:deep(.el-switch__core) {
  background-color: #ddd;
}

:deep(.el-switch.is-checked .el-switch__core) {
  background-color: #FFD700;
}

/* 响应式设计 */
@media (max-width: 375px) {
  .preference-section {
    margin: 70px 10px 10px;
    padding: 15px;
  }

  .provider-option,
  .mode-option {
    padding: 12px;
  }

  .provider-icon,
  .mode-icon {
    width: 36px;
    height: 36px;
    margin-right: 12px;
  }

  .provider-icon .css-icon,
  .mode-icon .css-icon {
    font-size: 20px;
  }

  :deep(.el-select) {
    width: 100%;
  }

  .offline-map-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }

  .map-actions {
    width: 100%;
    justify-content: flex-end;
  }

  .cache-info {
    flex-direction: column;
    text-align: center;
    gap: 15px;
  }
}
</style>