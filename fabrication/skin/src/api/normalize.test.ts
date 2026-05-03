import { describe, it, expect } from 'vitest';
import { normalizeTemplate } from './normalize';
import type { V1SessionTemplate } from './gen/models';

describe('normalizeTemplate', () => {
  it('should fill defaults when null', () => {
    const tpl = {
      id: 'test',
      name: 'test',
      defaults: null,
      sessionSchema: null,
    } as unknown as V1SessionTemplate;
    const result = normalizeTemplate(tpl);
    expect(result.defaults).toEqual({ env: {}, files: [] });
    expect(result.sessionSchema).toEqual({ envDefs: [], fileDefs: [] });
  });

  it('should fill defaults when undefined', () => {
    const tpl = { id: 'test', name: 'test' } as unknown as V1SessionTemplate;
    const result = normalizeTemplate(tpl);
    expect(result.defaults).toEqual({ env: {}, files: [] });
    expect(result.sessionSchema).toEqual({ envDefs: [], fileDefs: [] });
  });

  it('should preserve existing values', () => {
    const tpl = {
      id: 'test',
      name: 'test',
      defaults: { env: { KEY: 'val' }, files: [{ path: 'test', content: 'hi' }] },
      sessionSchema: {
        envDefs: [{ key: 'K', label: 'K', required: true, placeholder: '', sensitive: false }],
        fileDefs: [],
      },
    } as unknown as V1SessionTemplate;
    const result = normalizeTemplate(tpl);
    expect(result.defaults).toEqual(tpl.defaults);
    expect(result.sessionSchema).toEqual(tpl.sessionSchema);
  });
});
