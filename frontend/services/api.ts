import axios from 'axios';
import {
  Product,
  Order,
  CheckoutRequest,
  ReceiptAudit,
  OrderPayment,
  CreatePaymentChargeRequest,
  CashShift,
  OpenCashShiftRequest,
  OpenCashShiftResponse,
  CloseCashShiftRequest,
  CloseCashShiftResponse,
  OrderStatus,
} from '../types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

export const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

api.interceptors.request.use((config) => {
  if (typeof window !== 'undefined') {
    const token = localStorage.getItem('pos_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  }
  return config;
});

export const fetchProducts = async (): Promise<Product[]> => {
  const response = await api.get<Product[]>('/products');
  return response.data;
};

export const fetchProductBySKU = async (sku: string): Promise<Product> => {
  const response = await api.get<Product>(`/products/${sku}`);
  return response.data;
};

export const checkoutOrder = async (payload: CheckoutRequest): Promise<Order> => {
  const response = await api.post<Order>('/orders/checkout', payload, {
    headers: {
      'X-Idempotency-Key': payload.idempotency_key,
    },
  });
  return response.data;
};

export const createPaymentCharge = async (payload: CreatePaymentChargeRequest): Promise<OrderPayment> => {
  const response = await api.post<OrderPayment>('/payments/charge', payload);
  return response.data;
};

export const getPaymentByOrderID = async (orderId: string): Promise<OrderPayment> => {
  const response = await api.get<OrderPayment>(`/payments/order/${orderId}`);
  return response.data;
};

export const notifyPaymentStatus = async (payload: {
  order_id: string;
  transaction_status: string;
  status_code?: string;
  gross_amount?: string;
}): Promise<void> => {
  await api.post('/payments/webhook', payload);
};

export const getCurrentCashShift = async (): Promise<CashShift> => {
  const response = await api.get<CashShift>('/cash-shifts/current');
  return response.data;
};

export const openCashShift = async (payload: OpenCashShiftRequest): Promise<OpenCashShiftResponse> => {
  const response = await api.post<OpenCashShiftResponse>('/cash-shifts/open', payload);
  return response.data;
};

export const closeCashShift = async (payload: CloseCashShiftRequest): Promise<CloseCashShiftResponse> => {
  const response = await api.post<CloseCashShiftResponse>('/cash-shifts/close', payload);
  return response.data;
};

export const uploadReceiptScan = async (file: File): Promise<ReceiptAudit> => {
  const formData = new FormData();
  formData.append('image', file);

  const response = await api.post<ReceiptAudit>('/receipts/scan', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
  return response.data;
};

export const fetchReceiptAudits = async (): Promise<ReceiptAudit[]> => {
  const response = await api.get<ReceiptAudit[]>('/receipts/audits');
  return response.data;
};

export const fetchKitchenOrders = async (): Promise<Order[]> => {
  const response = await api.get<Order[]>('/kitchen/orders');
  return response.data;
};

export const updateOrderStatus = async (orderId: string, status: OrderStatus): Promise<void> => {
  await api.patch(`/kitchen/orders/${orderId}/status`, { status });
};

export interface SyncStatus {
  is_online: boolean;
  pending_orders: number;
  pending_shifts: number;
  pending_payments: number;
  last_synced_at?: string;
}

export const fetchSyncStatus = async (): Promise<SyncStatus> => {
  const response = await api.get<SyncStatus>('/sync/status');
  return response.data;
};

export const triggerSync = async (): Promise<void> => {
  await api.post('/sync/trigger', {});
};

export const fetchOrderHistory = async (): Promise<Order[]> => {
  const response = await api.get<Order[]>('/orders/history');
  return response.data;
};
