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
  status: 'PENDING' | 'PREPARING' | 'READY' | 'SERVED' | 'COMPLETED' | 'CANCELLED' | 'REFUNDED';
  idempotency_key: string;
  items: OrderItem[];
  created_at: string;
  updated_at: string;
}

export interface CheckoutItem {
  sku: string;
  quantity: number;
}

export interface CheckoutRequest {
  idempotency_key: string;
  items: CheckoutItem[];
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
  status: string;
  updated_at: string;
}

export interface CartItem {
  product: Product;
  quantity: number;
}
