import type { V1Skill, V1MCPServer, V1Rule, V1GitKey } from '@maze/fabrication';
import type { CreateGitKeyData } from '../components/git-keys/constants';

function now(): string {
  return new Date().toISOString();
}

function generateTokenMask(token: string): string {
  if (token.length <= 8) return '****';
  return token.slice(0, 4) + '****' + token.slice(-4);
}

interface SkillApi {
  list(): Promise<V1Skill[]>;
  create(data: {
    name: string;
    description?: string;
    config?: Record<string, string>;
  }): Promise<V1Skill>;
  get(name: string): Promise<V1Skill>;
  update(
    name: string,
    data: { description?: string; config?: Record<string, string> },
  ): Promise<V1Skill>;
  delete(name: string): Promise<void>;
}

interface MCPServerApi {
  list(): Promise<V1MCPServer[]>;
  create(data: {
    name: string;
    type: string;
    command?: string;
    url?: string;
    args?: string[];
    env?: Record<string, string>;
  }): Promise<V1MCPServer>;
  get(name: string): Promise<V1MCPServer>;
  update(
    name: string,
    data: {
      type: string;
      command?: string;
      url?: string;
      args?: string[];
      env?: Record<string, string>;
    },
  ): Promise<V1MCPServer>;
  delete(name: string): Promise<void>;
}

interface RuleApi {
  list(): Promise<V1Rule[]>;
  create(data: { name: string; content?: string }): Promise<V1Rule>;
  get(name: string): Promise<V1Rule>;
  update(name: string, data: { content?: string }): Promise<V1Rule>;
  delete(name: string): Promise<void>;
}

interface GitKeyApi {
  list(): Promise<V1GitKey[]>;
  create(data: CreateGitKeyData): Promise<V1GitKey>;
  get(name: string): Promise<V1GitKey>;
  delete(name: string): Promise<void>;
}

class MockSkillApi implements SkillApi {
  private items = new Map<string, V1Skill>();

  list(): Promise<V1Skill[]> {
    return Promise.resolve(Array.from(this.items.values()));
  }
  create(data: {
    name: string;
    description?: string;
    config?: Record<string, string>;
  }): Promise<V1Skill> {
    if (this.items.has(data.name))
      return Promise.reject(new Error(`Skill "${data.name}" already exists`));
    const item: V1Skill = {
      name: data.name,
      description: data.description ?? '',
      config: data.config ?? {},
      createdAt: now(),
      updatedAt: now(),
    };
    this.items.set(data.name, item);
    return Promise.resolve(item);
  }
  get(name: string): Promise<V1Skill> {
    const item = this.items.get(name);
    if (!item) return Promise.reject(new Error(`Skill "${name}" not found`));
    return Promise.resolve(item);
  }
  update(
    name: string,
    data: { description?: string; config?: Record<string, string> },
  ): Promise<V1Skill> {
    const item = this.items.get(name);
    if (!item) return Promise.reject(new Error(`Skill "${name}" not found`));
    const updated: V1Skill = {
      ...item,
      description: data.description ?? item.description,
      config: data.config ?? item.config,
      updatedAt: now(),
    };
    this.items.set(name, updated);
    return Promise.resolve(updated);
  }
  delete(name: string): Promise<void> {
    if (!this.items.has(name)) return Promise.reject(new Error(`Skill "${name}" not found`));
    this.items.delete(name);
    return Promise.resolve();
  }
}

class MockMCPApi implements MCPServerApi {
  private items = new Map<string, V1MCPServer>();

  list(): Promise<V1MCPServer[]> {
    return Promise.resolve(Array.from(this.items.values()));
  }
  create(data: {
    name: string;
    type: string;
    command?: string;
    url?: string;
    args?: string[];
    env?: Record<string, string>;
  }): Promise<V1MCPServer> {
    if (this.items.has(data.name))
      return Promise.reject(new Error(`MCPServer "${data.name}" already exists`));
    const item: V1MCPServer = {
      name: data.name,
      type: data.type,
      command: data.command ?? '',
      url: data.url ?? '',
      args: data.args ?? [],
      env: data.env ?? {},
      createdAt: now(),
      updatedAt: now(),
    };
    this.items.set(data.name, item);
    return Promise.resolve(item);
  }
  get(name: string): Promise<V1MCPServer> {
    const item = this.items.get(name);
    if (!item) return Promise.reject(new Error(`MCPServer "${name}" not found`));
    return Promise.resolve(item);
  }
  update(
    name: string,
    data: {
      type: string;
      command?: string;
      url?: string;
      args?: string[];
      env?: Record<string, string>;
    },
  ): Promise<V1MCPServer> {
    const item = this.items.get(name);
    if (!item) return Promise.reject(new Error(`MCPServer "${name}" not found`));
    const updated: V1MCPServer = {
      ...item,
      type: data.type ?? item.type,
      command: data.command ?? item.command,
      url: data.url ?? item.url,
      args: data.args ?? item.args,
      env: data.env ?? item.env,
      updatedAt: now(),
    };
    this.items.set(name, updated);
    return Promise.resolve(updated);
  }
  delete(name: string): Promise<void> {
    if (!this.items.has(name)) return Promise.reject(new Error(`MCPServer "${name}" not found`));
    this.items.delete(name);
    return Promise.resolve();
  }
}

class MockRuleApi implements RuleApi {
  private items = new Map<string, V1Rule>();

  list(): Promise<V1Rule[]> {
    return Promise.resolve(Array.from(this.items.values()));
  }
  create(data: { name: string; content?: string }): Promise<V1Rule> {
    if (this.items.has(data.name))
      return Promise.reject(new Error(`Rule "${data.name}" already exists`));
    const item: V1Rule = {
      name: data.name,
      content: data.content ?? '',
      createdAt: now(),
      updatedAt: now(),
    };
    this.items.set(data.name, item);
    return Promise.resolve(item);
  }
  get(name: string): Promise<V1Rule> {
    const item = this.items.get(name);
    if (!item) return Promise.reject(new Error(`Rule "${name}" not found`));
    return Promise.resolve(item);
  }
  update(name: string, data: { content?: string }): Promise<V1Rule> {
    const item = this.items.get(name);
    if (!item) return Promise.reject(new Error(`Rule "${name}" not found`));
    const updated: V1Rule = { ...item, content: data.content ?? item.content, updatedAt: now() };
    this.items.set(name, updated);
    return Promise.resolve(updated);
  }
  delete(name: string): Promise<void> {
    if (!this.items.has(name)) return Promise.reject(new Error(`Rule "${name}" not found`));
    this.items.delete(name);
    return Promise.resolve();
  }
}

class MockGitKeyApi implements GitKeyApi {
  private items = new Map<string, V1GitKey>();

  list(): Promise<V1GitKey[]> {
    return Promise.resolve(Array.from(this.items.values()));
  }
  create(data: CreateGitKeyData): Promise<V1GitKey> {
    if (this.items.has(data.name))
      return Promise.reject(new Error(`GitKey "${data.name}" already exists`));
    const item: V1GitKey = {
      name: data.name,
      tokenMask: generateTokenMask(data.token),
      tokenType: data.tokenType,
      host: data.host,
      createdAt: now(),
    };
    this.items.set(data.name, item);
    return Promise.resolve(item);
  }
  get(name: string): Promise<V1GitKey> {
    const item = this.items.get(name);
    if (!item) return Promise.reject(new Error(`GitKey "${name}" not found`));
    return Promise.resolve(item);
  }
  delete(name: string): Promise<void> {
    if (!this.items.has(name)) return Promise.reject(new Error(`GitKey "${name}" not found`));
    this.items.delete(name);
    return Promise.resolve();
  }
}

export interface FabricationApi {
  skills: SkillApi;
  mcpServers: MCPServerApi;
  rules: RuleApi;
  gitKeys: GitKeyApi;
}

export function createMockFabricationApi(): FabricationApi {
  return {
    skills: new MockSkillApi(),
    mcpServers: new MockMCPApi(),
    rules: new MockRuleApi(),
    gitKeys: new MockGitKeyApi(),
  };
}
