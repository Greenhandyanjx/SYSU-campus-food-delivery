<template>
  <div class="test-map-page">
    <h1>地图测试页面</h1>
    <div class="test-info">
      <p>API Key: {{ apiKey ? `${apiKey.substring(0, 8)}...` : '未设置' }}</p>
      <p>Security Code: {{ securityCode ? `${securityCode.substring(0, 8)}...` : '未设置' }}</p>
      <p>AMap loaded: {{ amapLoaded }}</p>
    </div>

    <el-button @click="loadMap" type="primary">加载地图</el-button>
    <el-button @click="testGeocode" type="success">测试地理编码</el-button>

    <div id="test-map-container" class="map-container"></div>

    <div class="test-output">
      <h3>测试输出:</h3>
      <pre>{{ testOutput }}</pre>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import amapLoader from '@/utils/amap';

const apiKey = ref('');
const securityCode = ref('');
const amapLoaded = ref(false);
const testOutput = ref('');

onMounted(() => {
  // 检查环境变量
  apiKey.value = import.meta.env.VITE_AMAP_KEY || '';
  securityCode.value = import.meta.env.VITE_AMAP_SECURITY_CODE || '';

  testOutput.value += `环境变量检查:\n`;
  testOutput.value += `API Key: ${apiKey.value ? apiKey.value.substring(0, 8) + '...' : '未设置'}\n`;
  testOutput.value += `Security Code: ${securityCode.value ? securityCode.value.substring(0, 8) + '...' : '未设置'}\n`;
  testOutput.value += `Backup API Key: ${import.meta.env.VITE_AMAP_KEY_BACKUP ? import.meta.env.VITE_AMAP_KEY_BACKUP.substring(0, 8) + '...' : '未设置'}\n\n`;

  // 检查 AMap 是否已加载
  if (window.AMap) {
    amapLoaded.value = true;
    testOutput.value += `AMap 已加载: ${typeof window.AMap}\n`;
  } else {
    testOutput.value += `AMap 未加载\n`;
  }
});

const loadMap = async () => {
  testOutput.value += '\n=== 开始加载地图 ===\n';

  try {
    const AMap = await amapLoader.load();
    amapLoaded.value = true;

    testOutput.value += `地图加载成功: ${typeof AMap}\n`;

    // 创建测试地图
    const mapContainer = document.getElementById('test-map-container');
    if (mapContainer) {
      const map = new AMap.Map(mapContainer, {
        zoom: 16,
        center: [113.299, 23.099], // 中山大学
        viewMode: '2D'
      });

      const marker = new AMap.Marker({
        position: [113.299, 23.099],
        title: '测试标记'
      });
      map.add(marker);

      testOutput.value += '测试地图创建成功\n';
    }
  } catch (error) {
    testOutput.value += `地图加载失败: ${error}\n`;
  }
};

const testGeocode = async () => {
  if (!window.AMap) {
    testOutput.value += '\n错误: AMap 未加载，请先加载地图\n';
    return;
  }

  testOutput.value += '\n=== 开始测试地理编码 ===\n';

  try {
    const AMap = window.AMap;
    AMap.plugin(['AMap.Geocoder'], () => {
      const geocoder = new AMap.Geocoder({
        city: '广州'
      });

      geocoder.getLocation('中山大学', (status: string, result: any) => {
        if (status === 'complete' && result.geocodes.length > 0) {
          const location = result.geocodes[0].location;
          testOutput.value += `地理编码成功:\n`;
          testOutput.value += `地址: 中山大学\n`;
          testOutput.value += `经度: ${location.lng}\n`;
          testOutput.value += `纬度: ${location.lat}\n`;
        } else {
          testOutput.value += `地理编码失败: ${status}\n`;
        }
      });
    });
  } catch (error) {
    testOutput.value += `地理编码测试失败: ${error}\n`;
  }
};
</script>

<style scoped lang="scss">
.test-map-page {
  padding: 20px;
  max-width: 800px;
  margin: 0 auto;
}

.test-info {
  background: #f5f5f5;
  padding: 15px;
  border-radius: 8px;
  margin: 20px 0;
}

.test-info p {
  margin: 5px 0;
  font-family: monospace;
}

.map-container {
  width: 100%;
  height: 400px;
  border: 1px solid #ddd;
  border-radius: 8px;
  margin: 20px 0;
}

.test-output {
  background: #f8f8f8;
  padding: 15px;
  border-radius: 8px;
  margin-top: 20px;
}

.test-output pre {
  background: #2d2d2d;
  color: #fff;
  padding: 10px;
  border-radius: 4px;
  overflow-x: auto;
  white-space: pre-wrap;
}

h1 {
  color: #333;
}

h3 {
  color: #555;
  margin-bottom: 10px;
}
</style>