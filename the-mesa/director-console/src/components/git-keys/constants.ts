export const TOKEN_TYPE_DISPLAY: Record<string, string> = {
  PERSONAL_ACCESS_TOKEN: 'PAT',
  SSH_KEY: 'SSH',
};

export interface CreateGitKeyData {
  name: string;
  token: string;
  tokenType: string;
  host?: string;
}
