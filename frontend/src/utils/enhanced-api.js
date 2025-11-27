import request, { connectionHealth, networkUtils } from './request';

/**
 * 增强的API包装器，提供更好的错误处理和用户反馈
 * - 自动网络检测和健康检查
 * - 智能重试和超时配置
 * - 详细的错误分类和用户友好提示
 * - 演示数据回退机制
 */

// 错误类型定义
const ERROR_TYPES = {
  NETWORK: 'network',
  TIMEOUT: 'timeout',
  SERVER: 'server',
  CLIENT: 'client',
  AUTH: 'auth',
  NOT_FOUND: 'not_found',
  VALIDATION: 'validation'
};

// 错误消息映射
const ERROR_MESSAGES = {
  [ERROR_TYPES.NETWORK]: '网络连接异常，请检查网络设置',
  [ERROR_TYPES.TIMEOUT]: '请求超时，请稍后重试',
  [ERROR_TYPES.SERVER]: '服务器暂时不可用，请稍后重试',
  [ERROR_TYPES.CLIENT]: '请求参数错误',
  [ERROR_TYPES.AUTH]: '身份验证失败，请重新登录',
  [ERROR_TYPES.NOT_FOUND]: '请求的资源不存在',
  [ERROR_TYPES.VALIDATION]: '数据格式错误'
};

class EnhancedAPI {
  constructor(baseApi) {
    this.baseApi = baseApi;
    this.lastRequestTime = {};
    this.requestCache = new Map();
    this.cacheTimeout = 30000; // 30秒缓存
  }

  // 错误分类
  classifyError(error) {
    if (!error.response) {
      // 网络错误
      if (error.code === 'ECONNABORTED') {
        return ERROR_TYPES.TIMEOUT;
      }
      return ERROR_TYPES.NETWORK;
    }

    const status = error.response.status;
    if (status >= 500) return ERROR_TYPES.SERVER;
    if (status === 401 || status === 403) return ERROR_TYPES.AUTH;
    if (status === 404) return ERROR_TYPES.NOT_FOUND;
    if (status >= 400 && status < 500) return ERROR_TYPES.CLIENT;

    return ERROR_TYPES.NETWORK;
  }

  // 获取用户友好的错误消息
  getErrorMessage(error, context = '') {
    const errorType = this.classifyError(error);
    const baseMessage = ERROR_MESSAGES[errorType] || '未知错误';

    // 如果后端返回了具体错误消息，优先使用
    const backendMessage = error.response?.data?.msg;
    if (backendMessage) {
      return backendMessage;
    }

    // 添加上下文信息
    if (context) {
      return `${context}: ${baseMessage}`;
    }

    return baseMessage;
  }

  // 检查是否应该使用演示数据
  shouldUseDemoData(error) {
    const errorType = this.classifyError(error);
    const isNetworkRelated = [ERROR_TYPES.NETWORK, ERROR_TYPES.TIMEOUT, ERROR_TYPES.SERVER].includes(errorType);
    const connectionUnhealthy = !connectionHealth.isHealthy();

    return isNetworkRelated && connectionUnhealthy;
  }

  // 生成缓存键
  getCacheKey(url, params = {}) {
    return `${url}:${JSON.stringify(params)}`;
  }

  // 检查缓存
  getCachedData(cacheKey) {
    const cached = this.requestCache.get(cacheKey);
    if (cached && Date.now() - cached.timestamp < this.cacheTimeout) {
      return cached.data;
    }
    return null;
  }

  // 设置缓存
  setCacheData(cacheKey, data) {
    this.requestCache.set(cacheKey, {
      data,
      timestamp: Date.now()
    });
  }

  // 增强的请求方法
  async enhancedRequest(config, options = {}) {
    const {
      context = '',
      useCache = false,
      fallbackData = null,
      showRetryButton = true,
      customErrorHandler = null
    } = options;

    const { url, method = 'GET', params = {} } = config;
    const cacheKey = this.getCacheKey(url, params);

    // 检查缓存
    if (method === 'GET' && useCache) {
      const cachedData = this.getCachedData(cacheKey);
      if (cachedData) {
        console.log(`Using cached data for ${url}`);
        return { success: true, data: cachedData, cached: true };
      }
    }

    // 检查网络状态
    if (!networkUtils.isOnline()) {
      const offlineError = new Error('设备处于离线状态');
      offlineError.type = ERROR_TYPES.NETWORK;
      offlineError.isOffline = true;

      if (fallbackData) {
        return { success: true, data: fallbackData, fallback: true, error: offlineError };
      }

      throw offlineError;
    }

    // 检查连接质量
    if (!networkUtils.isGoodConnection()) {
      console.warn('Network connection quality is poor, requests may be slow');
    }

    try {
      console.log(`Making ${method} request to ${url}`, {
        connectionHealthy: connectionHealth.isHealthy(),
        networkQuality: networkUtils.getConnectionInfo(),
        retryCount: config.__retryCount || 0
      });

      const response = await request({
        ...config,
        timeout: networkUtils.getRecommendedTimeout()
      });

      // 缓存成功响应
      if (method === 'GET' && useCache) {
        this.setCacheData(cacheKey, response);
      }

      return {
        success: true,
        data: response,
        cached: false
      };

    } catch (error) {
      const errorType = this.classifyError(error);
      error.type = errorType;
      error.userMessage = this.getErrorMessage(error, context);
      error.shouldRetry = [ERROR_TYPES.NETWORK, ERROR_TYPES.TIMEOUT, ERROR_TYPES.SERVER].includes(errorType);
      error.canUseDemo = this.shouldUseDemoData(error);

      console.error('Enhanced API Error:', {
        url,
        method,
        errorType,
        userMessage: error.userMessage,
        shouldRetry: error.shouldRetry,
        canUseDemo: error.canUseDemo,
        connectionStatus: connectionHealth.status,
        networkInfo: networkUtils.getConnectionInfo()
      });

      // 自定义错误处理器
      if (customErrorHandler) {
        const handled = customErrorHandler(error);
        if (handled !== undefined) {
          return handled;
        }
      }

      // 演示数据回退
      if (fallbackData && error.canUseDemo) {
        console.warn(`Using fallback data for ${url} due to ${errorType}`);
        return {
          success: true,
          data: fallbackData,
          fallback: true,
          error
        };
      }

      throw error;
    }
  }

  // 包装API方法
  wrapApiMethod(apiMethod, options = {}) {
    return async (...args) => {
      try {
        const config = apiMethod(...args);
        return await this.enhancedRequest(config, options);
      } catch (error) {
        return {
          success: false,
          error,
          data: null
        };
      }
    };
  }

  // 批量请求
  async batchRequests(requests, options = {}) {
    const results = [];
    const { failFast = false } = options;

    for (const requestConfig of requests) {
      try {
        const result = await this.enhancedRequest(requestConfig, options);
        results.push(result);

        if (failFast && !result.success) {
          break;
        }
      } catch (error) {
        results.push({
          success: false,
          error,
          data: null
        });

        if (failFast) {
          break;
        }
      }
    }

    return results;
  }

  // 重置所有状态
  reset() {
    connectionHealth.reset();
    this.requestCache.clear();
    this.lastRequestTime = {};
  }

  // 获取连接状态摘要
  getConnectionSummary() {
    return {
      healthy: connectionHealth.isHealthy(),
      consecutiveFailures: connectionHealth.status.consecutiveFailures,
      lastCheck: connectionHealth.status.lastCheck,
      networkInfo: networkUtils.getConnectionInfo(),
      isOnline: networkUtils.isOnline(),
      isGoodConnection: networkUtils.isGoodConnection(),
      recommendedTimeout: networkUtils.getRecommendedTimeout()
    };
  }
}

// 创建增强API实例
const enhancedAPI = new EnhancedAPI();

// 导出工具和实例
export {
  EnhancedAPI,
  enhancedAPI,
  ERROR_TYPES,
  ERROR_MESSAGES,
  connectionHealth,
  networkUtils
};

export default enhancedAPI;