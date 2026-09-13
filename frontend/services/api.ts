import axios from 'axios';
import { Product, Order, CheckoutRequest, ReceiptAudit } from '../types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

export const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
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
