// PicoClaw: Simple device auth implementation
// Compatible with OpenClaw UI format

/**
 * Builds device authentication payload
 * @param {Object} params - Authentication parameters
 * @returns {Object} - Device auth payload
 */
export function buildDeviceAuthPayload(params = {}) {
  return {
    deviceId: params.deviceId || 'picoclaw-device-' + Math.random().toString(36).substr(2, 9),
    timestamp: Date.now(),
    version: '1.0.0',
    ...params
  };
}