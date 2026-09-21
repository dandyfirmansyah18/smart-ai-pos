declare global {
  interface Window {
    snap?: {
      pay: (
        token: string,
        options?: {
          onSuccess?: (result: any) => void;
          onPending?: (result: any) => void;
          onError?: (result: any) => void;
          onClose?: () => void;
        }
      ) => void;
    };
  }
}

export enum OrderStatus {
  UNPAID = 'UNPAID',
  PENDING = 'PENDING',
  PREPARING = 'PREPARING',
  READY = 'READY',
  SERVED = 'SERVED',
  COMPLETED = 'COMPLETED',
  CANCELLED = 'CANCELLED',
  EXPIRED = 'EXPIRED',
  REFUNDED = 'REFUNDED',
}

export enum PaymentStatus {
  PENDING = 'PENDING',
  SUCCESS = 'SUCCESS',
  FAILED = 'FAILED',
  EXPIRED = 'EXPIRED',
}

export enum PaymentMethod {
  CASH = 'CASH',
  MIDTRANS = 'MIDTRANS',
  QRIS = 'QRIS',
  DEBIT = 'DEBIT',
}

export enum PaymentGateway {
  MIDTRANS = 'MIDTRANS',
  CASH = 'CASH',
}

export enum CashShiftStatus {
  OPEN = 'OPEN',
  CLOSED = 'CLOSED',
}

export enum UserRole {
  ADMIN = 'ADMIN',
  CASHIER = 'CASHIER',
  KITCHEN = 'KITCHEN',
  WAREHOUSE = 'WAREHOUSE',
}

export interface User {
  id: string;
  username: string;
  role: UserRole;
  full_name: string;
  created_at: string;
  updated_at: string;
}

export interface Product {
  id: string;
  sku: string;
  name: string;
  description: string;
  price: number;
  stock_quantity: number;
  created_at: string;
  updated_at: string;
}

export interface OrderItem {
  id: string;
  order_id: string;
  product_id: string;
  sku: string;
  name: string;
  quantity: number;
  unit_price: number;
}

export interface Order {
  id: string;
  transaction_id: string;
  total_amount: number;
  status: OrderStatus;
  idempotency_key: string;
  items: OrderItem[];
  created_at: string;
  updated_at: string;
}

export interface OrderPayment {
  id: string;
  order_id: string;
  payment_method: PaymentMethod;
  gateway: PaymentGateway;
  gateway_transaction_id: string;
  amount: number;
  status: PaymentStatus;
  snap_token?: string;
  snap_redirect_url?: string;
  raw_response?: string;
  created_at: string;
  updated_at: string;
}

export interface CashShift {
  id: string;
  user_id: string;
  status: CashShiftStatus;
  opening_cash: number;
  closing_cash: number;
  expected_cash: number;
  total_cash_sales: number;
  total_qris_sales?: number;
  total_debit_sales?: number;
  notes: string;
  opened_at: string;
  closed_at?: string;
}

export interface CheckoutItem {
  sku: string;
  quantity: number;
}

export interface CheckoutRequest {
  idempotency_key: string;
  payment_method?: PaymentMethod;
  items: CheckoutItem[];
}

export interface CreatePaymentChargeRequest {
  order_id: string;
  amount: number;
  payment_method: PaymentMethod;
}

export interface OpenCashShiftRequest {
  opening_cash: number;
  notes?: string;
}

export interface OpenCashShiftResponse {
  message: string;
  shift_id: string;
}

export interface CloseCashShiftRequest {
  closing_cash: number;
  notes?: string;
}

export interface CloseCashShiftResponse {
  message: string;
  opening_cash: number;
  total_sales: number;
  expected_cash: number;
  closing_cash: number;
  difference: number;
}

export interface OCRItem {
  name: string;
  quantity: number;
  price: number;
}

export interface ReceiptOCRResult {
  merchant_name: string;
  date: string;
  items: OCRItem[];
  total_amount: number;
}

export interface ReceiptAudit {
  id: string;
  merchant_name: string;
  receipt_date: string;
  total_amount: number;
  raw_ocr_json: ReceiptOCRResult;
  created_at: string;
}

export interface StockUpdateEvent {
  type: string;
  sku: string;
  new_stock: number;
  updated_at: string;
}

export interface OrderStatusUpdateEvent {
  type: string;
  id: string;
  status: OrderStatus;
  updated_at: string;
}

export interface CartItem {
  product: Product;
  quantity: number;
}
