// 在浏览器控制台中运行这个脚本来调试前端

console.log("🔍 开始调试前端聊天功能...");

// 1. 检查应用实例
const appElement = document.querySelector('openclaw-app');
if (!appElement) {
    console.error("❌ 找不到应用元素");
} else {
    console.log("✅ 找到应用元素:", appElement);
    
    // 2. 检查连接状态
    console.log("📡 连接状态:", {
        connected: appElement.connected,
        client: !!appElement.client,
        settings: appElement.settings
    });
    
    // 3. 检查网关URL
    console.log("🌐 Gateway URL:", appElement.settings?.gatewayUrl);
    
    // 4. 检查是否使用PicoClaw适配器
    const isPicoclaw = appElement.settings?.gatewayUrl?.includes("18790");
    console.log("🔧 使用PicoClaw:", isPicoclaw);
    
    // 5. 检查聊天状态
    console.log("💬 聊天状态:", {
        chatRunId: appElement.chatRunId,
        chatStream: appElement.chatStream,
        lastError: appElement.lastError
    });
    
    // 6. 手动触发连接
    if (!appElement.connected && appElement.handleGatewayUrlConfirm) {
        console.log("🔄 尝试重新连接...");
        appElement.handleGatewayUrlConfirm();
    }
    
    // 7. 检查事件日志
    if (appElement.eventLog) {
        console.log("📋 事件日志:", appElement.eventLog.slice(0, 5));
    }
    
    // 8. 测试发送消息
    if (appElement.connected && appElement.handleSendChat) {
        console.log("📤 测试发送消息...");
        try {
            appElement.handleSendChat("Hello from debug script");
            console.log("✅ 消息发送成功");
        } catch (error) {
            console.error("❌ 消息发送失败:", error);
        }
    } else {
        console.error("❌ 无法发送消息 - 连接状态或方法不可用");
    }
}

// 9. 检查是否有全局错误
console.log("🔍 检查全局错误...");
if (window.performance && window.performance.getEntriesByType) {
    const resources = window.performance.getEntriesByType('resource');
    const failedRequests = resources.filter(r => r.status >= 400);
    if (failedRequests.length > 0) {
        console.warn("⚠️ 失败的请求:", failedRequests);
    }
}

// 10. 添加事件监听器来监控连接变化
if (appElement) {
    const originalConnected = Object.getOwnPropertyDescriptor(appElement.constructor.prototype, 'connected');
    if (originalConnected && originalConnected.set) {
        const originalSet = originalConnected.set;
        originalConnected.set = function(value) {
            console.log("🔄 连接状态变化:", value);
            originalSet.call(this, value);
        };
    }
}

console.log("🏁 调试脚本执行完成。请查看上述输出。");