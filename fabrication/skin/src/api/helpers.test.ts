import { describe, it, expect } from 'vitest';
import { unwrapSdkResponse, unwrapVoidResponse } from './helpers';

// hey-api RequestResult 形态（throwOnError=false, responseStyle='fields'）：
// 成功 { data; error: undefined; request?; response? } / 失败 { data: undefined; error; request?; response? }
// error 是已 parse 的 response body（JSON 对象或字符串），response 是原始 Response（body 已被消费）。
describe('unwrapSdkResponse', () => {
  it('should unwrap successful response', async () => {
    const result = await unwrapSdkResponse(
      Promise.resolve({ data: { id: '123' }, error: undefined }),
    );
    expect(result.status).toBe('ok');
    expect(result.data).toEqual({ id: '123' });
  });

  it('should handle error with JSON body', async () => {
    const result = await unwrapSdkResponse(
      Promise.resolve({
        data: undefined,
        error: { code: 'conflict', message: 'Conflict detected' },
        response: new Response(null, { status: 409, statusText: 'Conflict' }),
      }),
    );
    expect(result.status).toBe('error');
    expect(result.code).toBe('conflict');
    expect(result.message).toBe('Conflict detected');
  });

  it('should parse reason from structured error response', async () => {
    const result = await unwrapSdkResponse(
      Promise.resolve({
        data: undefined,
        error: { code: 5, message: 'not found', reason: 'ARCHIVE_NOT_FOUND' },
        response: new Response(null, { status: 404, statusText: 'Not Found' }),
      }),
    );
    expect(result.status).toBe('error');
    expect(result.reason).toBe('ARCHIVE_NOT_FOUND');
    expect(result.details).toBeUndefined();
  });

  it('should parse details with fieldViolations', async () => {
    const result = await unwrapSdkResponse(
      Promise.resolve({
        data: undefined,
        error: {
          code: 3,
          message: 'invalid',
          reason: 'VALIDATION_FAILED',
          details: { fieldViolations: [{ field: 'name', description: 'required' }] },
        },
        response: new Response(null, { status: 400, statusText: 'Bad Request' }),
      }),
    );
    expect(result.status).toBe('error');
    expect(result.reason).toBe('VALIDATION_FAILED');
    expect(result.details?.fieldViolations).toEqual([{ field: 'name', description: 'required' }]);
  });

  it('should parse details with preconditionViolations and auto-fill conflicts', async () => {
    const result = await unwrapSdkResponse(
      Promise.resolve({
        data: undefined,
        error: {
          code: 9,
          message: 'conflict',
          reason: 'CONFIG_CONFLICT',
          details: {
            preconditionViolations: [
              { type: 'CONFIG_CONFLICT', subject: '/foo', description: 'abc' },
            ],
          },
        },
        response: new Response(null, { status: 400, statusText: 'Bad Request' }),
      }),
    );
    expect(result.status).toBe('error');
    expect(result.reason).toBe('CONFIG_CONFLICT');
    expect(result.details?.preconditionViolations).toEqual([
      { type: 'CONFIG_CONFLICT', subject: '/foo', description: 'abc' },
    ]);
    expect(result.conflicts).toEqual([{ path: '/foo', currentHash: 'abc' }]);
  });

  it('should handle error without reason or details', async () => {
    const result = await unwrapSdkResponse(
      Promise.resolve({
        data: undefined,
        error: { code: 13, message: 'internal error' },
        response: new Response(null, { status: 500, statusText: 'Internal Server Error' }),
      }),
    );
    expect(result.status).toBe('error');
    expect(result.reason).toBeUndefined();
    expect(result.details).toBeUndefined();
  });

  it('should fall back to HTTP status when error body is empty', async () => {
    const result = await unwrapSdkResponse(
      Promise.resolve({
        data: undefined,
        error: null,
        response: new Response(null, { status: 500, statusText: 'Internal Server Error' }),
      }),
    );
    expect(result.status).toBe('error');
    expect(result.message).toBe('HTTP 500');
  });

  it('should handle AbortError directly (reject path)', async () => {
    const abortErr = new DOMException('The operation was aborted', 'AbortError');
    const result = await unwrapSdkResponse(Promise.reject(abortErr));
    expect(result.status).toBe('error');
    expect(result.message).toBe('请求超时');
  });

  it('should handle generic Error (reject path)', async () => {
    const result = await unwrapSdkResponse(Promise.reject(new TypeError('Failed to fetch')));
    expect(result.status).toBe('error');
    expect(result.message).toBe('Failed to fetch');
  });

  it('should handle non-Error values (reject path)', async () => {
    // 测试非 Error 类型的 reject 值（如第三方库直接 reject 字符串），故意违反 prefer-promise-reject-errors
    const result = await unwrapSdkResponse(
      new Promise((_resolve, reject) => {
        // eslint-disable-next-line @typescript-eslint/prefer-promise-reject-errors
        reject('string error');
      }),
    );
    expect(result.status).toBe('error');
    expect(result.message).toBe('string error');
  });
});

describe('unwrapVoidResponse', () => {
  it('should return ok with undefined data', async () => {
    const result = await unwrapVoidResponse(Promise.resolve({ data: {}, error: undefined }));
    expect(result.status).toBe('ok');
    expect(result.data).toBeUndefined();
  });

  it('should handle errors', async () => {
    const result = await unwrapVoidResponse(Promise.reject(new Error('Network error')));
    expect(result.status).toBe('error');
    expect(result.message).toBe('Network error');
  });
});
