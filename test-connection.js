// 测试 PicoClaw 连接的脚本

// 模拟 PicoClaw 适配器
class PicoClawAdapter {
  constructor(url) {
    this.url = url;
    this.connected = false;
    this.messageId = 0;
    this.lastSeq = 0;
  }

  async connect() {
    // 转换 ws:// 到 http:// 
    let httpUrl = this.url.replace(/^ws:\/\//, "http://").replace(/^wss:\/\//, "https://");
    if (!httpUrl.endsWith("/api")) {
      httpUrl = httpUrl.replace(/\/gateway$/, "") + "/api";
    }

    console.log("Connecting to:", httpUrl);

    try {
      // 测试连接
      const response = await fetch(httpUrl, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          id: 'connect-test',
          method: 'connect',
          params: {}
        })
      });

      const result = await response.json();
      console.log("Connect result:", result);
      
      if (result.ok) {
        this.connected = true;
        console.log("✅ Connected successfully!");
        return true;
      } else {
        console.log("❌ Connection failed:", result.error);
        return false;
      }
    } catch (error) {
      console.error("❌ Connection error:", error);
      return false;
    }
  }

  async sendMessage(message) {
    if (!this.connected) {
      console.log("❌ Not connected");
      return;
    }

    let httpUrl = this.url.replace(/^ws:\/\//, "http://").replace(/^wss:\/\//, "https://");
    if (!httpUrl.endsWith("/api")) {
      httpUrl = httpUrl.replace(/\/gateway$/, "") + "/api";
    }

    try {
      const response = await fetch(httpUrl, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          id: String(++this.messageId),
          method: 'chat.send',
          params: {
            sessionKey: 'test-session',
            message: message,
            deliver: true,
            idempotencyKey: 'test-' + this.messageId
          }
        })
      });

      const result = await response.json();
      console.log("Message result:", result);
      return result;
    } catch (error) {
      console.error("Message error:", error);
    }
  }

  async pollEvents() {
    if (!this.connected) return;

    let httpUrl = this.url.replace(/^ws:\/\//, "http://").replace(/^wss:\/\//, "https://");
    if (!httpUrl.endsWith("/api")) {
      httpUrl = httpUrl.replace(/\/gateway$/, "") + "/api";
    }

    try {
      const response = await fetch(httpUrl, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          id: String(++this.messageId),
          method: 'events.poll',
          params: { lastSeq: this.lastSeq }
        })
      });

      const result = await response.json();
      console.log("Events result:", result);
      
      if (result.ok && result.payload && result.payload.events) {
        for (const event of result.payload.events) {
          if (event.seq > this.lastSeq) {
            this.lastSeq = event.seq;
          }
          console.log("📨 Event:", event.event, event.payload);
        }
      }
    } catch (error) {
      console.error("Poll error:", error);
    }
  }
}

// 测试连接
async function testConnection() {
  const gatewayUrl = "ws://localhost:18790/gateway"; // 默认的UI URL
  
  console.log("🔄 Testing PicoClaw connection to:", gatewayUrl);
  
  const adapter = new PicoClawAdapter(gatewayUrl);
  
  // 连接
  const connected = await adapter.connect();
  
  if (connected) {
    // 发送测试消息
    console.log("\n📤 Sending test message...");
    await adapter.sendMessage("你好，这是测试消息");
    
    // 等待一秒，然后轮询事件
    setTimeout(async () => {
      console.log("\n🔄 Polling for events...");
      await adapter.pollEvents();
    }, 2000);
  }
}

// 运行测试
testConnection();