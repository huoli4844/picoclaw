// PicoClaw: Tool policy implementation
// Compatible with OpenClaw UI format

/**
 * Tool policy types
 */
export const TOOL_POLICY_TYPES = {
  ALLOW_ALL: 'allow_all',
  DENY_ALL: 'deny_all',
  APPROVAL_REQUIRED: 'approval_required',
  WHITELIST: 'whitelist',
  BLACKLIST: 'blacklist'
};

/**
 * Default tool policies
 */
export const DEFAULT_TOOL_POLICIES = {
  dangerous_tools: {
    policy: TOOL_POLICY_TYPES.APPROVAL_REQUIRED,
    tools: ['exec', 'write_file', 'delete_file']
  },
  safe_tools: {
    policy: TOOL_POLICY_TYPES.ALLOW_ALL,
    tools: ['read_file', 'list_dir', 'web_search']
  }
};

/**
 * Checks if a tool requires approval
 * @param {string} toolName - Name of the tool
 * @param {Object} policies - Tool policies to check against
 * @returns {boolean} - Whether approval is required
 */
export function requiresApproval(toolName, policies = DEFAULT_TOOL_POLICIES) {
  for (const policy of Object.values(policies)) {
    if (policy.tools.includes(toolName)) {
      return policy.policy === TOOL_POLICY_TYPES.APPROVAL_REQUIRED;
    }
  }
  return false;
}

/**
 * Normalizes tool name for consistent comparison
 * @param {string} toolName - The tool name to normalize
 * @returns {string} - Normalized tool name
 */
export function normalizeToolName(toolName) {
  if (!toolName || typeof toolName !== 'string') {
    return '';
  }
  return toolName.toLowerCase().trim().replace(/[^a-z0-9_]/g, '_');
}

/**
 * Expands tool groups into individual tools
 * @param {Array} toolGroups - Array of tool groups
 * @returns {Array} - Expanded list of tools
 */
export function expandToolGroups(toolGroups) {
  if (!Array.isArray(toolGroups)) {
    return [];
  }
  
  const tools = [];
  for (const group of toolGroups) {
    if (typeof group === 'string') {
      tools.push(group);
    } else if (group && group.tools) {
      tools.push(...group.tools);
    }
  }
  return tools;
}

/**
 * Resolves tool profile policy
 * @param {string} toolName - Name of the tool
 * @param {Object} profile - Tool profile
 * @returns {Object} - Resolved policy
 */
export function resolveToolProfilePolicy(toolName, profile = {}) {
  const normalizedTool = normalizeToolName(toolName);
  const defaultPolicy = {
    allowed: true,
    requiresApproval: false,
    timeout: 30000
  };

  if (profile[normalizedTool]) {
    return { ...defaultPolicy, ...profile[normalizedTool] };
  }

  return defaultPolicy;
}