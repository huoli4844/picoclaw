// PicoClaw: Client info protocol
// Compatible with OpenClaw UI format

export const GATEWAY_CLIENT_MODES = {
  WEB: 'web',
  CLI: 'cli',
  API: 'api'
};

export const GATEWAY_CLIENT_NAMES = {
  OPENCLAW_CONTROL: 'openclaw-control-ui',
  PICOCLAW_CONTROL: 'picoclaw-control-ui',
  CLI: 'picoclaw-cli'
};

export const GATEWAY_CAPABILITIES = {
  CHAT: 'chat',
  STATUS: 'status',
  CONFIG: 'config',
  TOOLS: 'tools'
};