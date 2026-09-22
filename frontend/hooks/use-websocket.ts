'use client';

import { useEffect, useRef, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { Product, StockUpdateEvent } from '../types';

const WS_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws';

export function useWebSocketSync() {
  const queryClient = useQueryClient();
  const [isConnected, setIsConnected] = useState(false);
  const [lastEvent, setLastEvent] = useState<StockUpdateEvent | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    let reconnectTimeout: NodeJS.Timeout;

    const connect = () => {
      try {
        const ws = new WebSocket(WS_URL);
        wsRef.current = ws;

        ws.onopen = () => {
          setIsConnected(true);
          console.log('[WebSocket] Live sync connected to POS server');
        };

        ws.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data);
            if (data.type === 'stock_update') {
              setLastEvent(data);

              // Interceptively patch TanStack Query 'products' cache live on screen
              queryClient.setQueryData(['products'], (oldProducts: Product[] | undefined) => {
                if (!oldProducts) return [];
                return oldProducts.map((p) =>
                  p.sku === data.sku ? { ...p, stock_quantity: data.new_stock } : p
                );
              });
            } else if (data.type === 'order_status_update' || data.type === 'ORDER_STATUS_UPDATE') {
              queryClient.invalidateQueries({ queryKey: ['orderHistory'] });
              queryClient.invalidateQueries({ queryKey: ['kitchenOrders'] });
            }
          } catch (err) {
            console.error('[WebSocket] Error parsing message payload:', err);
          }
        };

        ws.onclose = () => {
          setIsConnected(false);
          console.log('[WebSocket] Connection closed. Attempting reconnect in 3s...');
          reconnectTimeout = setTimeout(connect, 3000);
        };

        ws.onerror = (error) => {
          console.error('[WebSocket] Error:', error);
          ws.close();
        };
      } catch (err) {
        console.error('[WebSocket] Dial error:', err);
        reconnectTimeout = setTimeout(connect, 5000);
      }
    };

    connect();

    return () => {
      if (reconnectTimeout) clearTimeout(reconnectTimeout);
      if (wsRef.current) {
        wsRef.current.onclose = null;
        wsRef.current.close();
      }
    };
  }, [queryClient]);

  return { isConnected, lastEvent };
}
