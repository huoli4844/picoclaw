// PicoClaw: Simple reasoning tags implementation
// Compatible with OpenClaw UI format

/**
 * Strips reasoning tags from text content
 * @param {string} text - The input text containing reasoning tags
 * @returns {string} - Text with reasoning tags removed
 */
export function stripReasoningTagsFromText(text) {
  if (!text || typeof text !== 'string') {
    return text;
  }

  // Remove reasoning content between special tags
  // Examples: <thinking>...</thinking>, <reasoning>...</reasoning>
  return text
    .replace(/<thinking>[\s\S]*?<\/thinking>/gi, '')
    .replace(/<reasoning>[\s\S]*?<\/reasoning>/gi, '')
    .replace(/<scratchpad>[\s\S]*?<\/scratchpad>/gi, '')
    .replace(/<analysis>[\s\S]*?<\/analysis>/gi, '')
    .trim();
}

/**
 * Legacy export for compatibility
 */
export const stripThinkingTags = stripReasoningTagsFromText;