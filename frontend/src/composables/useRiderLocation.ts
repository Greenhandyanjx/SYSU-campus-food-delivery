import { ref, onMounted, onUnmounted } from 'vue';
import { ElMessage } from 'element-plus';
import { riderApi } from '@/api/rider';

export interface LocationData {
  latitude: number;
  longitude: number;
  address?: string;
}

const useRiderLocation = () => {
  const isTracking = ref(false);
  const currentPosition = ref<LocationData | null>(null);
  const locationError = ref<string | null>(null);
  let watchId: number | null = null;
  let reportInterval: NodeJS.Timeout | null = null;
  let lastReportedPosition: LocationData | null = null;

  // 距离计算函数（单位：米）
  const calculateDistance = (pos1: LocationData, pos2: LocationData): number => {
    const R = 6371e3; // 地球半径（米）
    const φ1 = (pos1.latitude * Math.PI) / 180;
    const φ2 = (pos2.latitude * Math.PI) / 180;
    const Δφ = ((pos2.latitude - pos1.latitude) * Math.PI) / 180;
    const Δλ = ((pos2.longitude - pos1.longitude) * Math.PI) / 180;

    const a = Math.sin(Δφ / 2) * Math.sin(Δφ / 2) +
              Math.cos(φ1) * Math.cos(φ2) *
              Math.sin(Δλ / 2) * Math.sin(Δλ / 2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));

    return R * c;
  };

  // 上报位置到服务器
  const reportLocation = async (location: LocationData) => {
    try {
      await riderApi.updateLocation(location);
      lastReportedPosition = { ...location };

      // 在控制台显示成功上报的信息
      console.log('📍 [位置上报成功]', {
        时间: new Date().toLocaleTimeString(),
        纬度: location.latitude,
        经度: location.longitude,
        地址: location.address || '未提供'
      });
    } catch (error: any) {
      // 在控制台显示失败信息
      console.error('❌ [位置上报失败]', {
        时间: new Date().toLocaleTimeString(),
        尝试上报的位置: location,
        错误信息: error?.response?.data?.msg || error.message || '未知错误'
      });

      // 只在第一次失败时提示，避免刷屏
      if (!locationError.value) {
        locationError.value = error?.response?.data?.msg || '位置上报失败';
        ElMessage.warning(locationError.value);
      }
    }
  };

  // 获取当前位置
  const getCurrentPosition = (): Promise<LocationData> => {
    return new Promise((resolve, reject) => {
      if (!navigator.geolocation) {
        reject(new Error('浏览器不支持定位功能'));
        return;
      }

      navigator.geolocation.getCurrentPosition(
        (position) => {
          const location: LocationData = {
            latitude: position.coords.latitude,
            longitude: position.coords.longitude,
          };
          currentPosition.value = location;
          resolve(location);
        },
        (error) => {
          let errorMsg = '获取位置失败';
          switch (error.code) {
            case error.PERMISSION_DENIED:
              errorMsg = '定位权限被拒绝，请在浏览器设置中允许定位';
              break;
            case error.POSITION_UNAVAILABLE:
              errorMsg = '无法获取位置信息';
              break;
            case error.TIMEOUT:
              errorMsg = '定位请求超时';
              break;
          }
          locationError.value = errorMsg;
          reject(new Error(errorMsg));
        },
        {
          enableHighAccuracy: true,
          timeout: 10000,
          maximumAge: 30000, // 允许使用30秒内的缓存位置
        }
      );
    });
  };

  // 发送定位状态事件
  const emitLocationStatus = () => {
    window.dispatchEvent(new CustomEvent('rider:locationStatus', {
      detail: {
        isTracking: isTracking.value,
        error: locationError.value,
        location: currentPosition.value,
      }
    }));
  };

  // 发送位置更新事件
  const emitLocationUpdate = () => {
    if (currentPosition.value) {
      window.dispatchEvent(new CustomEvent('rider:locationUpdate', {
        detail: {
          location: currentPosition.value,
          timestamp: Date.now()
        }
      }));
    }
  };

  // 开始位置追踪
  const startLocationTracking = async () => {
    console.log('🚀 [开始位置追踪]', {
      时间: new Date().toLocaleTimeString(),
      说明: '骑手位置追踪已启动，将每15秒上报一次位置'
    });

    try {
      // 首先获取当前位置
      const initialLocation = await getCurrentPosition();
      console.log('📍 [获取初始位置成功]', {
        时间: new Date().toLocaleTimeString(),
        纬度: initialLocation.latitude,
        经度: initialLocation.longitude
      });

      await reportLocation(initialLocation);

      // 设置位置监听
      if (navigator.geolocation) {
        watchId = navigator.geolocation.watchPosition(
          async (position) => {
            const location: LocationData = {
              latitude: position.coords.latitude,
              longitude: position.coords.longitude,
            };
            currentPosition.value = location;
            emitLocationUpdate();

            console.log('🔄 [位置更新]', {
              时间: new Date().toLocaleTimeString(),
              新位置: {
                纬度: location.latitude,
                经度: location.longitude
              }
            });

            // 如果位置变化超过20米，立即上报
            if (lastReportedPosition) {
              const distance = calculateDistance(location, lastReportedPosition);
              if (distance > 20) {
                console.log('🏃 [位置变化超过阈值]', {
                  时间: new Date().toLocaleTimeString(),
                  距离变化: `${Math.round(distance)}米`,
                  阈值: '20米',
                  说明: '立即上报新位置'
                });
                await reportLocation(location);
              }
            }
          },
          (error) => {
            console.error('Location watch error:', error);
            locationError.value = '位置监听出错';
            emitLocationStatus();
          },
          {
            enableHighAccuracy: true,
            timeout: 10000,
            maximumAge: 15000, // 允许使用15秒内的缓存位置
          }
        );
      }

      // 设置定时上报（每15秒）
      reportInterval = setInterval(async () => {
        if (currentPosition.value) {
          console.log('⏰ [定时上报位置]', {
            时间: new Date().toLocaleTimeString(),
            说明: '15秒定时上报',
            当前位置: {
              纬度: currentPosition.value.latitude,
              经度: currentPosition.value.longitude
            }
          });
          await reportLocation(currentPosition.value);
        } else {
          console.log('⚠️ [定时上报跳过]', {
            时间: new Date().toLocaleTimeString(),
            说明: '暂无当前位置数据'
          });
        }
      }, 15000);

      isTracking.value = true;
      locationError.value = null;
      emitLocationStatus();
    } catch (error: any) {
      locationError.value = error.message;
      ElMessage.error(error.message);
      emitLocationStatus();
    }
  };

  // 停止位置追踪
  const stopLocationTracking = () => {
    if (watchId !== null) {
      navigator.geolocation.clearWatch(watchId);
      watchId = null;
    }

    if (reportInterval !== null) {
      clearInterval(reportInterval);
      reportInterval = null;
    }

    isTracking.value = false;
    emitLocationStatus();
  };

  // 检查定位权限
  const checkLocationPermission = async (): Promise<boolean> => {
    if (!navigator.geolocation) {
      ElMessage.error('浏览器不支持定位功能');
      return false;
    }

    if ('permissions' in navigator) {
      try {
        const result = await navigator.permissions.query({ name: 'geolocation' });
        if (result.state === 'denied') {
          ElMessage.error('定位权限被拒绝，请在浏览器设置中允许定位');
          return false;
        }
      } catch (error) {
        console.warn('Cannot check location permission:', error);
      }
    }

    return true;
  };

  // 组件挂载时不自动开始追踪，改为手动控制
  // onMounted(() => {
  //   // 延迟1秒后开始，确保页面加载完成
  //   setTimeout(() => {
  //     checkLocationPermission().then((hasPermission) => {
  //       if (hasPermission) {
  //         startLocationTracking();
  //       }
  //     });
  //   }, 1000);
  // });

  // 组件卸载时清理
  onUnmounted(() => {
    stopLocationTracking();
  });

  return {
    isTracking,
    currentPosition,
    locationError,
    startLocationTracking,
    stopLocationTracking,
    checkLocationPermission,
    getCurrentPosition,
  };
};

export default useRiderLocation;