import type { V1SessionTemplate, V1ConfigLayer, V1SessionSchema } from './gen/models';

const emptySchema: Required<V1SessionSchema> = { envDefs: [], fileDefs: [] };
const emptyDefaults: Required<V1ConfigLayer> = { env: {}, files: [] };

// normalizeTemplate 保证 defaults 和 sessionSchema 一定存在，消除 optional 链
export type NormalizedTemplate = Omit<V1SessionTemplate, 'defaults' | 'sessionSchema'> & {
  defaults: Required<V1ConfigLayer>;
  sessionSchema: Required<V1SessionSchema>;
};

export function normalizeTemplate(tpl: V1SessionTemplate): NormalizedTemplate {
  return {
    ...tpl,
    defaults: tpl.defaults ? { env: tpl.defaults.env ?? {}, files: tpl.defaults.files ?? [] } : emptyDefaults,
    sessionSchema: tpl.sessionSchema ? { envDefs: tpl.sessionSchema.envDefs ?? [], fileDefs: tpl.sessionSchema.fileDefs ?? [] } : emptySchema,
  };
}
