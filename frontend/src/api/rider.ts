import { http } from "@/utils/http";

// 聊天消息类型
export type ChatMessage = {
  id: number;
  from_base_id: number;
  merchant_id: number;
  user_base_id: number;
  content: string;
  type: string; // text/image/other
  status: string; // sent/delivered/read
  created_at: string;
  delivered_at?: string;
  read_at?: string;
};

// 会话列表项类型
export type ChatSession = {
  merchant_id: number;
  merchant_name?: string;
  last_message: string;
  last_at: string;
  unread_count: number;
};

export type ApiResp<T> = { code: number | string; msg: string; data: T };

export type RiderMe = {
  id: number;
  name: string;
  avatar: string;
  phone: string;
  rating: number;
  completedOrders: number;
  isOnline: boolean;
};

export type RiderOrderItem = {
  id: number;
  restaurant: string;
  pickupAddress: string;
  customer: string;
  deliveryAddress: string;
  distance: number;
  estimatedFee: number;
  estimatedTime: number;
  createdAt: string;
  status: number;

  // 聊天功能字段 - 后端现在提供这些数据
  merchantId: number;        // 商家ID
  userId: number;            // 下单用户的ID
  userBaseId: number;        // 用户的base_user_id，用于聊天

  acceptedAt?: string | null;
  pickupAt?: string | null;
  deliverAt?: string | null;
  finishAt?: string | null;
};
export type RiderWallet = {
  id: number;
  riderId: number;
  balance: number;
  frozenAmount: number;
  totalIncome: number;
};

export type RiderIncomeRecord = {
  id: number;
  riderId: number;
  orderId: number;
  amount: number;
  type: string; // order | bonus | adjustment
  remark: string;
  createdAt: string;
};

export type RiderWithdraw = {
  id: number;
  riderId: number;
  amount: number;
  account: string;
  status: string; // pending | success | failed
  appliedAt: string;
  processedAt?: string | null;
};

export type RiderStat = {
  newCount: number;
  ongoingCount: number;
  historyCount: number;
  completedCount: number;
  todayIncome: number;
  monthIncome: number;
};
export const riderApi = {
  getMe() {
    return http.get<ApiResp<RiderMe>>("/rider/me");
  },
  updateOnline(isOnline: boolean) {
    return http.post<ApiResp<{ success: boolean }>>("/rider/online", { isOnline });
  },
  getStat() {
    return http.get<ApiResp<RiderStat>>("/rider/stat");
  },

  getNewOrders() {
    return http.get<ApiResp<RiderOrderItem[]>>("/rider/orders/new");
  },
  acceptOrder(id: number) {
    return http.post<ApiResp<{ success: boolean }>>(`/rider/orders/${id}/accept`);
  },
  pickupOrder(id: number) {
    return http.post<ApiResp<{ success: boolean }>>(`/rider/orders/${id}/pickup`);
  },
  deliverOrder(id: number) {
    return http.post<ApiResp<{ success: boolean }>>(`/rider/orders/${id}/deliver`);
  },

  getOngoing() {
    return http.get<ApiResp<RiderOrderItem[]>>("/rider/orders/ongoing");
  },
  getHistory() {
    return http.get<ApiResp<RiderOrderItem[]>>("/rider/orders/history");
  },
    getWallet() {
    return http.get<ApiResp<RiderWallet>>("/rider/wallet");
  },
  getIncome(params?: { page?: number; size?: number }) {
    return http.get<ApiResp<RiderIncomeRecord[]>>("/rider/income", { params });
  },
  applyWithdraw(payload: { amount: number; account: string }) {
    return http.post<ApiResp<{ success: boolean; withdrawId?: number }>>("/rider/withdraw", payload);
  },
  getWithdraws() {
    return http.get<ApiResp<RiderWithdraw[]>>("/rider/withdraws");
  },

  // 聊天相关API - 骑手使用用户端的聊天接口
  getChatHistory(merchantId: number, userBaseId?: number) {
    const params: any = { merchantId };
    if (userBaseId) params.userBaseId = userBaseId;
    return http.get<ApiResp<ChatMessage[]>>("/chat/history", { params });
  },

  getChatSessions() {
    return http.get<ApiResp<ChatSession[]>>("/user/chats");
  },

  markChatAsRead(merchantId: number) {
    return http.post<ApiResp<{ success: boolean }>>("/user/chats/mark_read", { merchant_id: merchantId });
  },

  sendMessage(payload: {
    merchant_id: number;
    user_base_id: number;
    content: string;
    type: string;
  }) {
    // WebSocket发送消息，这里提供方法但不直接发送
    // 实际发送通过chatClientWebSocket
    return Promise.resolve({ success: true, payload });
  },

  // 定位上报API
  updateLocation(payload: {
    latitude: number;
    longitude: number;
    address?: string;
  }) {
    return http.post<ApiResp<{ success: boolean }>>("/rider/location", payload);
  },
};
