import { describe, it, expect } from 'vitest';
import { unwrapSdkResponse, unwrapVoidResponse } from './helpers';
import { ResponseError } from './gen/runtime';

describe('unwrapSdkResponse', () => {
  it('should unwrap successful response', async () => {
    const result = await unwrapSdkResponse(Promise.resolve({ id: '123' }));
    expect(result.status).toBe('ok');
    expect(result.data).toEqual({ id: '123' });
  });

  it('should handle ResponseError with JSON body', async () => {
    const errorResponse = new Response(
      JSON.stringify({ code: 'conflict', message: 'Conflict detected' }),
      { status: 409, statusText: 'Conflict' },
    );
    const error = new ResponseError(errorResponse, 'Response returned an error code');
    const result = await unwrapSdkResponse(Promise.reject(error));
    expect(result.status).toBe('error');
    expect(result.code).toBe('conflict');
    expect(result.message).toBe('Conflict detected');
  });

  it('should handle ResponseError with empty body', async () => {
    const errorResponse = new Response(null, { status: 500, statusText: 'Internal Server Error' });
    const error = new ResponseError(errorResponse, 'Response returned an error code');
    const result = await unwrapSdkResponse(Promise.reject(error));
    expect(result.status).toBe('error');
    expect(result.message).toBe('HTTP 500');
  });

  it('should handle AbortError directly', async () => {
    const abortErr = new DOMException('The operation was aborted', 'AbortError');
    const result = await unwrapSdkResponse(Promise.reject(abortErr));
    expect(result.status).toBe('error');
    expect(result.message).toBe('请求超时');
  });

  it('should handle generic Error', async () => {
    const result = await unwrapSdkResponse(Promise.reject(new TypeError('Failed to fetch')));
    expect(result.status).toBe('error');
    expect(result.message).toBe('Failed to fetch');
  });

  it('should handle non-Error values', async () => {
    // 测试非 Error 类型的 reject 值（如第三方库直接 reject 字符串），故意违反 prefer-promise-reject-errors
    // eslint-disable-next-line @typescript-eslint/prefer-promise-reject-errors
    const result = await unwrapSdkResponse(new Promise((_resolve, reject) => { reject('string error'); }));
    expect(result.status).toBe('error');
    expect(result.message).toBe('string error');
  });
});

describe('unwrapVoidResponse', () => {
  it('should return ok with undefined data', async () => {
    const result = await unwrapVoidResponse(Promise.resolve({}));
    expect(result.status).toBe('ok');
    expect(result.data).toBeUndefined();
  });

  it('should handle errors', async () => {
    const result = await unwrapVoidResponse(Promise.reject(new Error('Network error')));
    expect(result.status).toBe('error');
    expect(result.message).toBe('Network error');
  });
});
