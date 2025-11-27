import axios from "axios";

/**
 * 增强的 axios 实例
 * - baseURL: 使用相对路径 '/api'，在开发模式下 Vite 的 proxy 会将其转发到后端服务
 * - 请求拦截器：将从 localStorage 读取 token 并注入 Authorization 头
 * - 响应拦截器：打印后端错误信息并将错误抛出，方便上层处理
 * - 连接健康检查：在后端服务不可用时提供更好的错误反馈
 * - 指数退避重试：对网络错误进行智能重试
 *
 * 后端开发注意：建议统一返回标准 JSON 结构，例如 { code: 1, data: ..., msg: '' }
 */

// 连接状态跟踪
let connectionStatus = {
  isHealthy: true,
  lastCheck: Date.now(),
  consecutiveFailures: 0,
  maxRetries: 3
};

// 简单的健康检查端点
const HEALTH_CHECK_ENDPOINT = '/health';

// 指数退避重试函数
const getRetryDelay = (attempt) => {
  const baseDelay = 1000; // 1秒基础延迟
  const maxDelay = 10000; // 最大10秒延迟
  const delay = Math.min(baseDelay * Math.pow(2, attempt), maxDelay);
  return delay + Math.random() * 1000; // 添加随机抖动
};

// 健康检查函数
const checkHealth = async () => {
  try {
    // 使用轻量级的健康检查
    await axios.get('/api/health', {
      timeout: 3000,
      validateStatus: (status) => status < 500 // 接受4xx状态码
    });

    connectionStatus.isHealthy = true;
    connectionStatus.consecutiveFailures = 0;
    connectionStatus.lastCheck = Date.now();
    return true;
  } catch (error) {
    connectionStatus.isHealthy = false;
    connectionStatus.consecutiveFailures++;
    connectionStatus.lastCheck = Date.now();
    console.warn('Backend health check failed:', error.message);
    return false;
  }
};

// 创建 axios 实例
const service = axios.create({
  baseURL: '/api',
  timeout: 8000, // 增加超时时间以应对网络波动
  retry: 3, // 最大重试次数
  retryDelay: getRetryDelay,
  retryCondition: (error) => {
    // 只对网络错误和5xx错误重试
    return !error.response || error.response.status >= 500 || error.code === 'ECONNABORTED';
  }
});

// 请求拦截器
service.interceptors.request.use(
  async (config) => {
    // 在请求前检查连接状态
    const now = Date.now();
    const timeSinceLastCheck = now - connectionStatus.lastCheck;

    // 如果超过30秒或连接不健康，进行健康检查
    if (timeSinceLastCheck > 30000 || !connectionStatus.isHealthy) {
      await checkHealth();
    }

    // 如果连接不健康且连续失败次数过多，抛出错误
    if (!connectionStatus.isHealthy && connectionStatus.consecutiveFailures > 2) {
      console.warn(`Backend connection unstable (${connectionStatus.consecutiveFailures} consecutive failures)`);
      // 在请求头中添加连接状态信息，供后端参考
      config.headers['X-Connection-Status'] = 'unstable';
    }

    // 添加重试信息到请求头
    if (config.__retryCount) {
      config.headers['X-Retry-Count'] = config.__retryCount;
    }

    // 这里可以统一加 token
    const token = localStorage.getItem("token");
    if (token) {
      // token may already include the 'Bearer ' prefix (backend returns it that way).
      // Avoid duplicating 'Bearer ' (which would produce 'Bearer Bearer ...').
      config.headers.Authorization = token.startsWith('Bearer ') ? token : `Bearer ${token}`;
    }

    return config;
  },
  (error) => Promise.reject(error)
);

// 响应拦截器
service.interceptors.response.use(
  (response) => {
    // 成功响应，重置连接状态
    connectionStatus.isHealthy = true;
    connectionStatus.consecutiveFailures = 0;
    connectionStatus.lastCheck = Date.now();

    // 直接返回后端数据，而不是axios响应对象
    // 这样前端代码就能直接得到 { code: 1, data: ..., msg: "" } 格式
    return response.data;
  },
  async (error) => {
    // 增强的错误处理
    const config = error.config;

    // 如果没有配置信息，直接拒绝
    if (!config) {
      console.error("API Error (no config):", error.message);
      return Promise.reject(error);
    }

    // 初始化重试计数
    config.__retryCount = config.__retryCount || 0;

    // 详细错误信息记录
    console.error("API Error Details:", {
      message: error.message,
      code: error.code,
      status: error.response?.status,
      statusText: error.response?.statusText,
      url: config.url,
      method: config.method,
      retryCount: config.__retryCount,
      connectionStatus: connectionStatus.isHealthy ? 'healthy' : 'unhealthy',
      consecutiveFailures: connectionStatus.consecutiveFailures
    });

    // 检查是否应该重试
    const shouldRetry = (
      config.retry !== false && // 允许重试
      config.__retryCount < config.retry && // 未超过重试次数
      (error.code === 'ECONNABORTED' || // 超时
       error.code === 'NETWORK_ERROR' || // 网络错误
       !error.response || // 无响应
       error.response.status >= 500) // 服务器错误
    );

    if (shouldRetry) {
      config.__retryCount += 1;

      // 计算重试延迟
      const delay = getRetryDelay(config.__retryCount - 1);

      console.warn(`Retrying request (${config.__retryCount}/${config.retry}) to ${config.url} after ${delay}ms`);

      // 等待延迟后重试
      await new Promise(resolve => setTimeout(resolve, delay));

      return service(config);
    }

    // 更新连接状态
    if (error.code === 'ECONNABORTED' || !error.response) {
      connectionStatus.isHealthy = false;
      connectionStatus.consecutiveFailures++;
    }

    return Promise.reject(error);
  }
);

// 导出连接状态和健康检查函数
export const connectionHealth = {
  status: connectionStatus,
  check: checkHealth,
  isHealthy: () => connectionStatus.isHealthy && connectionStatus.consecutiveFailures < 3,
  reset: () => {
    connectionStatus.isHealthy = true;
    connectionStatus.consecutiveFailures = 0;
    connectionStatus.lastCheck = Date.now();
  }
};

// 网络状态检测工具
export const networkUtils = {
  // 检查网络连接状态
  isOnline: () => navigator.onLine,

  // 获取网络信息（如果支持）
  getConnectionInfo: () => {
    if ('connection' in navigator) {
      const conn = navigator.connection;
      return {
        effectiveType: conn.effectiveType,
        downlink: conn.downlink,
        rtt: conn.rtt,
        saveData: conn.saveData
      };
    }
    return null;
  },

  // 判断网络是否适合API调用
  isGoodConnection: () => {
    if (!navigator.onLine) return false;

    const connInfo = networkUtils.getConnectionInfo();
    if (!connInfo) return true; // 如果没有网络信息API，假设连接良好

    // 避免在慢速网络（2g）或保存数据模式下频繁API调用
    return connInfo.effectiveType !== '2g' && !connInfo.saveData;
  },

  // 计算建议的API超时时间
  getRecommendedTimeout: () => {
    const connInfo = networkUtils.getConnectionInfo();
    if (!connInfo) return 8000;

    const baseTimeout = 5000;
    const networkMultiplier = {
      'slow-2g': 4,
      '2g': 3,
      '3g': 2,
      '4g': 1.5
    };

    const multiplier = networkMultiplier[connInfo.effectiveType] || 2;
    return baseTimeout * multiplier;
  }
};

export default service;