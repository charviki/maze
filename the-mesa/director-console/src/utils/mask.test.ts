import { describe, it, expect } from 'vitest';
import { maskEnvValue, maskFileContent } from '@maze/fabrication';

describe('maskEnvValue', () => {
  it('当 sensitive=true 时，应对值进行脱敏', () => {
    expect(maskEnvValue('ANY_KEY', 'my-secret-value', true)).toBe('my-s****alue');
  });

  it('当 key 包含 PASSWORD 时，即使 sensitive 未设置也应脱敏', () => {
    expect(maskEnvValue('DB_PASSWORD', 'superlongpassword')).toBe('supe****word');
  });

  it('当 key 包含 SECRET 时，即使 sensitive 未设置也应脱敏', () => {
    expect(maskEnvValue('APP_SECRET', 'mysecretvalue123')).toBe('myse****e123');
  });

  it('当 key 包含 TOKEN 时，即使 sensitive 未设置也应脱敏', () => {
    expect(maskEnvValue('AUTH_TOKEN', 'abcdefghijk')).toBe('abcd****hijk');
  });

  it('当 key 包含 KEY 时，即使 sensitive 未设置也应脱敏', () => {
    expect(maskEnvValue('API_KEY', 'longapikey12345')).toBe('long****2345');
  });

  it('当 key 包含 API（但不含敏感词）时，不应脱敏', () => {
    expect(maskEnvValue('API_URL', 'http://localhost:3000')).toBe('http://localhost:3000');
  });

  it('当 key 和 sensitive 都不敏感时，返回原始值', () => {
    expect(maskEnvValue('HOST', 'localhost')).toBe('localhost');
  });

  it('短值（不足8位）应全部脱敏为 ****', () => {
    expect(maskEnvValue('PASSWORD', 'short', true)).toBe('****');
  });

  it('空字符串应原样返回', () => {
    expect(maskEnvValue('PASSWORD', '', true)).toBe('');
  });

  it('包含 AUTH 的 key 应被脱敏', () => {
    expect(maskEnvValue('AUTH_HEADER', 'bearer-token-value')).toBe('bear****alue');
  });

  it('包含 CREDENTIAL 的 key 应被脱敏', () => {
    expect(maskEnvValue('AWS_CREDENTIAL', 'AKIAIOSFODNN7EXAMPLE')).toBe('AKIA****MPLE');
  });

  it('包含 PRIVATE_KEY 的 key 应被脱敏', () => {
    // 前4后4规则: "-----BEGIN RSA PRIVATE KEY-----" -> 前4 "----" + **** + 后4 "----"
    expect(maskEnvValue('SSH_PRIVATE_KEY', '-----BEGIN RSA PRIVATE KEY-----')).toBe('----****----');
  });
});

describe('maskFileContent', () => {
  it('对 JSON 文件中的敏感字段进行脱敏', () => {
    const content = JSON.stringify({
      username: 'admin',
      password: 'mysecretpassword',
      api_key: 'sk-1234567890abcdef',
    });
    const result = maskFileContent('config.json', content);
    const parsed = JSON.parse(result);

    expect(parsed.username).toBe('admin');
    expect(parsed.password).toBe('myse****word');
    expect(parsed.api_key).toBe('sk-1****cdef');
  });

  it('对嵌套 JSON 中的敏感字段进行递归脱敏', () => {
    const content = JSON.stringify({
      database: {
        host: 'localhost',
        token: 'secret-token-value',
      },
    });
    const result = maskFileContent('db.json', content);
    const parsed = JSON.parse(result);

    expect(parsed.database.host).toBe('localhost');
    expect(parsed.database.token).toBe('secr****alue');
  });

  it('对非 JSON 文本中的 key=value 格式进行脱敏', () => {
    const content = 'PASSWORD=supersecret123\nHOST=localhost';
    const result = maskFileContent('config.env', content);

    expect(result).toContain('****');
    expect(result).toContain('localhost');
  });

  it('对非 JSON 文本中的 key: value 格式进行脱敏', () => {
    const content = 'secret: my-secret-value-here\nname: test';
    const result = maskFileContent('config.yaml', content);

    expect(result).toContain('****');
    expect(result).toContain('test');
  });

  it('空内容应原样返回', () => {
    expect(maskFileContent('file.txt', '')).toBe('');
  });

  it('无效 JSON 内容应原样返回（降级为文本处理）', () => {
    const content = 'this is not valid json { broken';
    const result = maskFileContent('config.json', content);

    // 降级到 maskTextContent 处理，原文不含敏感 key=value 模式，应原样返回
    expect(result).toBe(content);
  });
});
