const SENSITIVE_KEYWORDS = [
  'password', 'passwd', 'secret', 'token', 'api_key', 'apikey',
  'auth', 'credential', 'private_key', 'access_key', 'session_key',
  'private', 'key',
];

function isSensitiveKey(key: string): boolean {
  const lower = key.toLowerCase();
  return SENSITIVE_KEYWORDS.some(kw => lower.includes(kw));
}

// 保留前4后4，不足8位全掩码
function maskString(val: string): string {
  if (val.length <= 8) return '****';
  return val.slice(0, 4) + '****' + val.slice(-4);
}

// 环境变量值脱敏：优先用 sensitive 标记，其次用 key 名规则
export function maskEnvValue(key: string, value: string, sensitive?: boolean): string {
  if (!value) return value;
  if (sensitive || isSensitiveKey(key)) {
    return maskString(value);
  }
  return value;
}

// JSON 内容脱敏：递归扫描所有 key-value，对敏感 key 的 value 做脱敏
function maskJsonValue(obj: unknown): unknown {
  if (typeof obj === 'string') return obj;
  if (Array.isArray(obj)) return obj.map(item => maskJsonValue(item));
  if (obj !== null && typeof obj === 'object') {
    const result: Record<string, unknown> = {};
    for (const [k, v] of Object.entries(obj as Record<string, unknown>)) {
      if (typeof v === 'string' && isSensitiveKey(k)) {
        result[k] = maskString(v);
      } else {
        result[k] = maskJsonValue(v);
      }
    }
    return result;
  }
  return obj;
}

function maskJsonContent(content: string): string {
  try {
    const parsed = JSON.parse(content);
    const masked = maskJsonValue(parsed);
    return JSON.stringify(masked, null, 2);
  } catch {
    return content;
  }
}

// 通用文本脱敏：正则匹配 key=value / key: value 等常见模式
function maskTextContent(content: string): string {
  // 匹配 "key": "value" / key=value / key: value 格式中 key 包含敏感词的行
  return content.replace(
    /(["']?)([\w.-]+)\1\s*[:=]\s*(["']?)(.+?)\3/gi,
    (match, q1, key, q2, value) => {
      if (isSensitiveKey(key)) {
        return `${q1}${key}${q1} ${match.includes(':') ? ':' : '='} ${q2}${maskString(value)}${q2}`;
      }
      return match;
    }
  );
}

// 文件内容脱敏（根据文件类型智能处理）
export function maskFileContent(path: string, content: string): string {
  if (!content) return content;
  if (path.endsWith('.json')) return maskJsonContent(content);
  return maskTextContent(content);
}
