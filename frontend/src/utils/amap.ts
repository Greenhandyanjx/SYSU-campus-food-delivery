/**
 * 高德地图加载器
 * 负责单次加载高德地图 SDK，避免重复加载
 */

interface AMapLoaderOptions {
  key?: string;
  securityCode?: string;
  version?: string;
  plugins?: string[];
}

class AMapLoader {
  private static instance: AMapLoader;
  private isLoaded = false;
  private isLoading = false;
  private loadPromise: Promise<void> | null = null;
  private loadCallbacks: ((amap: any) => void)[] = [];

  private constructor() {}

  static getInstance(): AMapLoader {
    if (!AMapLoader.instance) {
      AMapLoader.instance = new AMapLoader();
    }
    return AMapLoader.instance;
  }

  /**
   * 加载高德地图 SDK
   */
  async load(options: AMapLoaderOptions = {}): Promise<any> {
    // 如果已经加载完成，直接返回
    if (this.isLoaded && window.AMap) {
      console.log('✅ AMap already loaded, plugins available:', Object.keys(window.AMap).filter(key => key.startsWith('AMap')));
      return window.AMap;
    }

    // 如果正在加载中，等待加载完成
    if (this.isLoading && this.loadPromise) {
      console.log('⏳ AMap loading in progress...');
      return this.loadPromise.then(() => window.AMap);
    }

    // 检查是否已经有脚本标签但没有完成加载
    const existingScript = document.querySelector('script[src*="webapi.amap.com"]');
    if (existingScript && !window.AMap) {
      console.log('AMap script exists but not loaded, removing and reloading...');
      existingScript.remove();
    }

    // 开始加载
    this.isLoading = true;
    this.loadPromise = this.loadScript(options);

    try {
      await this.loadPromise;
      this.isLoaded = true;
      this.isLoading = false;

      // 执行所有等待的回调
      this.loadCallbacks.forEach(callback => callback(window.AMap));
      this.loadCallbacks = [];

      return window.AMap;
    } catch (error) {
      this.isLoading = false;
      this.loadPromise = null;
      throw error;
    }
  }

  /**
   * 添加加载完成回调
   */
  onLoad(callback: (amap: any) => void): void {
    if (this.isLoaded && window.AMap) {
      callback(window.AMap);
    } else {
      this.loadCallbacks.push(callback);
    }
  }

  private async loadScript(options: AMapLoaderOptions): Promise<void> {
    return new Promise((resolve, reject) => {
      // 检查是否已经存在脚本标签
      const existingScript = document.querySelector('script[src*="webapi.amap.com"]');
      if (existingScript) {
        console.log('AMap script already found, checking if loaded...');

        // 如果 AMap 对象已经存在，直接返回
        if (window.AMap) {
          console.log('AMap object already available');
          this.isLoaded = true;
          resolve();
          return;
        }

        // 移除未加载完成的脚本
        console.log('Removing incomplete AMap script...');
        existingScript.remove();
      }

      // 调试环境变量
      console.log('Environment variables:', {
        VITE_AMAP_KEY: import.meta.env.VITE_AMAP_KEY,
        VITE_AMAP_KEY_BACKUP: import.meta.env.VITE_AMAP_KEY_BACKUP,
        VITE_AMAP_SECURITY_CODE: import.meta.env.VITE_AMAP_SECURITY_CODE
      });

      // 设置安全配置，支持备用密钥
      const key = options.key || import.meta.env.VITE_AMAP_KEY || import.meta.env.VITE_AMAP_KEY_BACKUP || 'e3064e9e20ff62d8ebb59d24d634c179';
      const securityCode = options.securityCode || import.meta.env.VITE_AMAP_SECURITY_CODE;

      console.log('Using AMap configuration:', {
        key: key,
        securityCode: securityCode ? `${securityCode.substring(0, 8)}...` : 'not set',
        isBackupKey: key === import.meta.env.VITE_AMAP_KEY_BACKUP
      });

      // 设置全局安全配置（如果有安全代码）
      if (securityCode) {
        window._AMapSecurityConfig = {
          securityJsCode: securityCode,
        };
        console.log('Security code configured');
      } else {
        console.log('No security code provided');
      }

      // 高德地图2.0版本 - 不预先加载插件，而是动态加载
      console.log('Loading AMap SDK v2.0 (plugins will be loaded dynamically)');

      // 创建脚本标签 - 使用最新版本 2.0，不预先加载插件
      const script = document.createElement('script');
      script.id = 'amap-sdk';
      script.type = 'text/javascript';
      script.src = `https://webapi.amap.com/maps?v=2.0&key=${key}`;
      script.charset = 'utf-8';
      script.async = true;

      console.log('Loading AMap SDK from:', script.src);

      script.onload = () => {
        console.log('AMap SDK loaded successfully');
        setTimeout(() => {
          if (window.AMap) {
            console.log('✅ AMap object available:', typeof window.AMap);
            const availablePlugins = Object.keys(window.AMap).filter(key => key.startsWith('AMap'));
            console.log('📦 Available plugins:', availablePlugins);

            if (availablePlugins.includes('AMap.Geocoder')) {
              console.log('✅ AMap.Geocoder plugin is loaded');
            } else {
              console.error('❌ AMap.Geocoder plugin NOT loaded');
            }
          } else {
            console.error('❌ AMap script loaded but window.AMap not found');
          }
        }, 500); // 增加延迟时间确保插件加载完成
        resolve();
      };

      script.onerror = (error) => {
        console.error('Failed to load AMap SDK with key', key, ':', error);

        // 如果使用的是主密钥且存在备用密钥，尝试使用备用密钥
        if (key === import.meta.env.VITE_AMAP_KEY && import.meta.env.VITE_AMAP_KEY_BACKUP) {
          console.log('Attempting to load with backup key...');

          // 移除失败的脚本
          script.remove();

          // 用备用密钥重新创建脚本
          const backupScript = document.createElement('script');
          backupScript.id = 'amap-sdk-backup';
          backupScript.type = 'text/javascript';
          // 更新备用密钥的插件列表
          const backupPlugins = [
            'AMap.Geocoder',
            'AMap.PlaceSearch',
            'AMap.ToolBar',
            'AMap.Scale',
            'AMap.HawkEye',
            'AMap.MapType',
            'AMap.Geolocation',
            'AMap.Marker',
            'AMap.InfoWindow'
          ];
          backupScript.src = `https://webapi.amap.com/maps?v=2.0&key=${import.meta.env.VITE_AMAP_KEY_BACKUP}&plugin=${backupPlugins.join(',')}`;
          backupScript.charset = 'utf-8';
          backupScript.async = true;

          backupScript.onload = () => {
            console.log('AMap SDK loaded successfully with backup key');
            setTimeout(() => {
              if (window.AMap) {
                console.log('AMap object available with backup key:', typeof window.AMap);
              }
            }, 100);
            resolve();
          };

          backupScript.onerror = (backupError) => {
            console.error('Failed to load AMap SDK with backup key:', backupError);
            reject(new Error('Failed to load AMap SDK with all available keys'));
          };

          document.head.appendChild(backupScript);
        } else {
          reject(new Error('Failed to load AMap SDK with all available keys'));
        }
      };

      document.head.appendChild(script);
    });
  }

  /**
   * 检查是否已加载
   */
  isLoadedSuccessfully(): boolean {
    return this.isLoaded && !!window.AMap;
  }
}

// 导出单例实例
export const amapLoader = AMapLoader.getInstance();

// 导出类型声明
declare global {
  interface Window {
    AMap: any;
    _AMapSecurityConfig: {
      securityJsCode: string;
    };
  }
}

export default amapLoader;