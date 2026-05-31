// gRPC-Web 连接配置
// 注意: 当前使用 REST API 作为过渡，后续迁移至 gRPC-Web

const API_BASE = '/api';
const GRPC_BASE = 'http://localhost:9090';

// 通用 fetch 封装
async function grpcFetch<T>(
  service: string,
  method: string,
  body?: Record<string, unknown>
): Promise<T> {
  const url = `${GRPC_BASE}/${service}/${method}`;
  const response = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: body ? JSON.stringify(body) : undefined,
  });

  if (!response.ok) {
    throw new Error(`gRPC request failed: ${response.status}`);
  }

  return response.json();
}

// REST API 兼容层 (过渡期间使用)
async function restFetch<T>(
  path: string,
  options?: RequestInit
): Promise<T> {
  const url = `${API_BASE}${path}`;
  const response = await fetch(url, {
    headers: {
      'Content-Type': 'application/json',
    },
    ...options,
  });

  if (!response.ok) {
    throw new Error(`REST request failed: ${response.status}`);
  }

  return response.json();
}

export { grpcFetch, restFetch, API_BASE, GRPC_BASE };
