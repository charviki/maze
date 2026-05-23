import type { FabricationItem } from './Sidebar';
import { SkillList } from './skills/SkillList';
import { MCPList } from './mcp/MCPList';
import { RuleList } from './rules/RuleList';
import { GitKeyList } from './git-keys/GitKeyList';
import { controllerApi } from '../api/controller';

const skillApi = {
  list: controllerApi.listSkills,
  create: controllerApi.createSkill,
  update: controllerApi.updateSkill,
  delete: (name: string) => controllerApi.deleteSkill(name).then(() => {}),
};

const mcpApi = {
  list: controllerApi.listMCPServers,
  create: controllerApi.createMCPServer,
  update: controllerApi.updateMCPServer,
  delete: (name: string) => controllerApi.deleteMCPServer(name).then(() => {}),
};

const ruleApi = {
  list: controllerApi.listRules,
  create: controllerApi.createRule,
  update: controllerApi.updateRule,
  delete: (name: string) => controllerApi.deleteRule(name).then(() => {}),
};

const gitKeyApi = {
  list: controllerApi.listGitKeys,
  create: controllerApi.createGitKey,
  delete: (name: string) => controllerApi.deleteGitKey(name).then(() => {}),
};

export function FabricationPanel({ item }: { item: FabricationItem }) {
  return (
    <div className="flex-1 w-full flex min-w-0">
      {item === 'skills' && <SkillList api={skillApi} />}
      {item === 'mcp-servers' && <MCPList api={mcpApi} />}
      {item === 'rules' && <RuleList api={ruleApi} />}
      {item === 'git-keys' && <GitKeyList api={gitKeyApi} />}
    </div>
  );
}
