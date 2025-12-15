import { http } from "@/utils/http";

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
export const riderApi = {
  getMe() {
    return http.get<ApiResp<RiderMe>>("/rider/me");
  },
  updateOnline(isOnline: boolean) {
    return http.post<ApiResp<{ success: boolean }>>("/rider/online", { isOnline });
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
};
