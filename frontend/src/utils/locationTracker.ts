/**
 * 骑手定位追踪器
 * 负责获取骑手位置并定时上报到后端
 */

interface LocationPosition {
  latitude: number;
  longitude: number;
  accuracy?: number;
  timestamp: number;
}

interface LocationUpdateCallback {
  (position: LocationPosition, error?: string): void;
}

class LocationTracker {
  private static instance: LocationTracker;
  private watchId: number | null = null;
  private reportInterval: number | null = null;
  private lastReportedPosition: LocationPosition | null = null;
  private currentPosition: LocationPosition | null = null;
  private isTracking = false;
  private hasPermission = false;
  private callbacks: LocationUpdateCallback[] = [];
  private reportCount = 0;
  private consecutiveErrors = 0;

  // 配置参数
  private readonly REPORT_INTERVAL = 15000; // 15秒上报一次
  private readonly MIN_DISTANCE = 20; // 最小移动距离（米）
  private readonly MAX_CONSECUTIVE_ERRORS = 3; // 最大连续错误次数

  private constructor() {}

  static getInstance(): LocationTracker {
    if (!LocationTracker.instance) {
      LocationTracker.instance = new LocationTracker();
    }
    return LocationTracker.instance;
  }

  /**
   * 开始定位追踪
   */
  async startTracking(): Promise<boolean> {
    if (this.isTracking) {
      console.log('Location tracking already started');
      return true;
    }

    try {
      // 检查浏览器支持
      if (!navigator.geolocation) {
        throw new Error('浏览器不支持定位功能');
      }

      // 请求定位权限
      const permission = await this.requestLocationPermission();
      if (!permission) {
        throw new Error('定位权限被拒绝');
      }

      // 开始监听位置变化
      this.watchId = navigator.geolocation.watchPosition(
        (position) => this.handlePositionSuccess(position),
        (error) => this.handlePositionError(error),
        {
          enableHighAccuracy: true,
          timeout: 10000,
          maximumAge: 60000 // 1分钟内的缓存位置可用
        }
      );

      // 启动定时上报
      this.startPeriodicReport();

      this.isTracking = true;
      this.hasPermission = true;
      console.log('Location tracking started successfully');

      // 发送状态更新事件，但不发送位置（因为还没获取到）
      this.updateGlobalStatus(this.currentPosition);

      // 尝试获取一次当前位置
      try {
        const initialPosition = await this.getCurrentLocationOnce();
        console.log('🎯 [定位追踪] 获取初始位置成功:', initialPosition);
      } catch (error) {
        console.log('⚠️ [定位追踪] 初始位置获取失败，等待自动定位:', error);
      }

      return true;

    } catch (error) {
      console.error('Failed to start location tracking:', error);
      this.isTracking = false;
      this.hasPermission = false;
      this.notifyCallbacks(null, error instanceof Error ? error.message : '定位启动失败');
      return false;
    }
  }

  /**
   * 停止定位追踪
   */
  stopTracking(): void {
    if (this.watchId !== null) {
      navigator.geolocation.clearWatch(this.watchId);
      this.watchId = null;
    }

    if (this.reportInterval !== null) {
      clearInterval(this.reportInterval);
      this.reportInterval = null;
    }

    this.isTracking = false;
    console.log('Location tracking stopped');
  }

  /**
   * 请求定位权限
   */
  private async requestLocationPermission(): Promise<boolean> {
    try {
      // 尝试获取一次当前位置来检查权限
      await new Promise<GeolocationPosition>((resolve, reject) => {
        navigator.geolocation.getCurrentPosition(resolve, reject, {
          enableHighAccuracy: true,
          timeout: 5000
        });
      });
      return true;
    } catch (error) {
      console.warn('Location permission check failed:', error);
      return false;
    }
  }

  /**
   * 处理位置获取成功
   */
  private handlePositionSuccess(position: GeolocationPosition): void {
    const newLocation: LocationPosition = {
      latitude: position.coords.latitude,
      longitude: position.coords.longitude,
      accuracy: position.coords.accuracy,
      timestamp: Date.now()
    };

    // 检查位置是否有效
    if (!this.isValidLocation(newLocation)) {
      console.warn('Invalid location data received');
      return;
    }

    this.currentPosition = newLocation;
    this.consecutiveErrors = 0; // 重置错误计数

    // 更新全局状态
    this.updateGlobalStatus(newLocation);
  }

  /**
   * 处理位置获取错误
   */
  private handlePositionError(error: GeolocationPositionError): void {
    console.error('Location error:', error);

    this.consecutiveErrors++;

    let errorMessage = '定位失败';
    switch (error.code) {
      case error.PERMISSION_DENIED:
        errorMessage = '定位权限被拒绝';
        this.hasPermission = false;
        break;
      case error.POSITION_UNAVAILABLE:
        errorMessage = '无法获取位置信息';
        break;
      case error.TIMEOUT:
        errorMessage = '定位超时';
        break;
    }

    // 如果连续错误过多，停止追踪
    if (this.consecutiveErrors >= this.MAX_CONSECUTIVE_ERRORS) {
      console.warn(`Too many consecutive errors (${this.consecutiveErrors}), stopping tracking`);
      this.stopTracking();
    }

    this.updateGlobalStatus(null, errorMessage);
  }

  /**
   * 验证位置数据是否有效
   */
  private isValidLocation(location: LocationPosition): boolean {
    // 检查经纬度是否在合理范围内
    const { latitude, longitude, accuracy } = location;

    // 经纬度范围检查
    if (latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180) {
      return false;
    }

    // 精度检查（如果精度超过1公里，可能数据不准确）
    if (accuracy && accuracy > 1000) {
      console.warn(`Low accuracy location: ${accuracy}m`);
    }

    return true;
  }

  /**
   * 计算两个位置之间的距离（米）
   */
  private calculateDistance(pos1: LocationPosition, pos2: LocationPosition): number {
    const R = 6371000; // 地球半径（米）
    const dLat = this.toRadians(pos2.latitude - pos1.latitude);
    const dLon = this.toRadians(pos2.longitude - pos1.longitude);
    const a = Math.sin(dLat / 2) * Math.sin(dLat / 2) +
      Math.cos(this.toRadians(pos1.latitude)) * Math.cos(this.toRadians(pos2.latitude)) *
      Math.sin(dLon / 2) * Math.sin(dLon / 2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
    return R * c;
  }

  private toRadians(degrees: number): number {
    return degrees * (Math.PI / 180);
  }

  /**
   * 启动定时上报
   */
  private startPeriodicReport(): void {
    this.reportInterval = window.setInterval(() => {
      if (this.currentPosition) {
        this.reportLocation(this.currentPosition);
      }
    }, this.REPORT_INTERVAL);
  }

  /**
   * 上报位置到后端
   */
  private async reportLocation(position: LocationPosition): Promise<void> {
    // 检查是否需要上报（距离变化或首次上报）
    if (this.lastReportedPosition) {
      const distance = this.calculateDistance(this.lastReportedPosition, position);
      if (distance < this.MIN_DISTANCE) {
        return; // 距离变化太小，不上报
      }
    }

    try {
      const token = localStorage.getItem('token');
      console.log('📍 [位置上报] 准备上报位置:', {
        latitude: position.latitude,
        longitude: position.longitude,
        hasToken: !!token,
        tokenLength: token?.length || 0
      });

      const response = await fetch('/api/rider/location', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
          latitude: position.latitude,
          longitude: position.longitude,
          // address 字段可选，暂时不上报地址
        })
      });

      const responseData = await response.json();

      if (!response.ok) {
        console.error('❌ [位置上报] 服务器返回错误:', {
          status: response.status,
          statusText: response.statusText,
          data: responseData
        });
        throw new Error(`HTTP ${response.status}: ${responseData.msg || 'Unknown error'}`);
      }

      this.lastReportedPosition = position;
      this.reportCount++;
      console.log(`✅ [位置上报] 成功上报 (${this.reportCount}):`, {
        position: position,
        serverResponse: responseData
      });

    } catch (error) {
      console.error('❌ [位置上报] 失败:', error);

      // 只在第一次失败时提示用户
      if (this.reportCount === 0) {
        console.warn('⚠️ [位置上报] 首次位置上报失败，可能是网络问题或服务器错误');
      }
    }
  }

  /**
   * 更新全局状态
   */
  private updateGlobalStatus(position: LocationPosition | null, error?: string): void {
    // 发送全局事件，让其他组件知道定位状态
    window.dispatchEvent(new CustomEvent('rider:locationStatus', {
      detail: {
        isTracking: this.isTracking,
        hasPermission: this.hasPermission,
        position: position,
        error: error,
        accuracy: position?.accuracy
      }
    }));

    // 如果有位置更新，也发送位置更新事件
    if (position) {
      window.dispatchEvent(new CustomEvent('rider:locationUpdate', {
        detail: {
          position: position,
          accuracy: position.accuracy,
          timestamp: position.timestamp
        }
      }));
    }

    // 通知回调
    this.notifyCallbacks(position, error);
  }

  /**
   * 添加状态变化回调
   */
  addCallback(callback: LocationUpdateCallback): void {
    this.callbacks.push(callback);
  }

  /**
   * 移除状态变化回调
   */
  removeCallback(callback: LocationUpdateCallback): void {
    const index = this.callbacks.indexOf(callback);
    if (index > -1) {
      this.callbacks.splice(index, 1);
    }
  }

  /**
   * 通知所有回调
   */
  private notifyCallbacks(position: LocationPosition | null, error?: string): void {
    this.callbacks.forEach(callback => {
      try {
        callback(position!, error);
      } catch (e) {
        console.error('Callback error:', e);
      }
    });
  }

  /**
   * 获取当前位置
   */
  getCurrentPosition(): LocationPosition | null {
    return this.currentPosition;
  }

  /**
   * 获取定位状态
   */
  getStatus() {
    return {
      isTracking: this.isTracking,
      hasPermission: this.hasPermission,
      currentPosition: this.currentPosition,
      reportCount: this.reportCount,
      consecutiveErrors: this.consecutiveErrors
    };
  }

  /**
   * 手动获取一次位置
   */
  async getCurrentLocationOnce(): Promise<LocationPosition> {
    return new Promise((resolve, reject) => {
      navigator.geolocation.getCurrentPosition(
        (position) => {
          const location: LocationPosition = {
            latitude: position.coords.latitude,
            longitude: position.coords.longitude,
            accuracy: position.coords.accuracy,
            timestamp: Date.now()
          };
          resolve(location);
        },
        (error) => {
          reject(new Error(`定位失败: ${error.message}`));
        },
        {
          enableHighAccuracy: true,
          timeout: 10000
        }
      );
    });
  }
}

// 导出单例实例
export const locationTracker = LocationTracker.getInstance();

// 导出类型
export type { LocationPosition, LocationUpdateCallback };

export default locationTracker;