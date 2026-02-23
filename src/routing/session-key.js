// PicoClaw: Session key routing utilities
// Compatible with OpenClaw UI format

/**
 * Parses agent session key for routing
 * @param {string} sessionKey - The session key to parse
 * @returns {Object} - Parsed routing information
 */
export function parseAgentSessionKey(sessionKey) {
  if (!sessionKey || typeof sessionKey !== 'string') {
    return {
      agentId: 'main',
      sessionId: 'default',
      channel: 'cli'
    };
  }

  // Format: agentId:sessionId:channel or agentId:sessionId
  const parts = sessionKey.split(':');
  
  if (parts.length === 3) {
    return {
      agentId: parts[0],
      sessionId: parts[1],
      channel: parts[2]
    };
  }
  
  if (parts.length === 2) {
    return {
      agentId: parts[0],
      sessionId: parts[1],
      channel: 'cli'
    };
  }

  return {
    agentId: 'main',
    sessionId: sessionKey,
    channel: 'cli'
  };
}