import type { GatewayBrowserClientOptions, GatewayEventFrame, GatewayResponseFrame } from "./gateway.ts";

export class PicoClawAdapter {
  private opts: GatewayBrowserClientOptions;
  private connected = false;
  private closed = false;
  private messageId = 0;
  private lastSeq = 0;
  private polling = false;

  constructor(opts: GatewayBrowserClientOptions) {
    this.opts = opts;
  }

  start() {
    console.log("[PicoClaw] Starting adapter, URL:", this.opts.url);
    this.closed = false;
    this.polling = false;
    // Convert ws:// to http:// for REST API
    let httpUrl = this.opts.url.replace(/^ws:\/\//, "http://").replace(/^wss:\/\//, "https://");
    
    // Use /api endpoint instead of WebSocket
    if (!httpUrl.endsWith("/api")) {
      httpUrl = httpUrl.replace(/\/gateway$/, "") + "/api";
    }

    console.log("[PicoClaw] Converted to HTTP URL:", httpUrl);

    // Start connection process
    this.connect(httpUrl);
  }

  stop() {
    console.log("[PicoClaw] Stopping adapter...");
    this.closed = true;
    this.connected = false;
    this.polling = false;
  }

  get isConnectionOpen() {
    return this.connected && !this.closed;
  }

  stop() {
    this.closed = true;
    this.connected = false;
  }

  get isConnectionOpen() {
    return this.connected;
  }

  private async connect(httpUrl: string) {
    console.log("[PicoClaw] Connecting to:", httpUrl);
    try {
      // Send connect request (use direct request to avoid recursion)
      const helloResponse = await this.makeDirectRequest("connect", {
        token: this.opts.token,
        clientInfo: {
          name: this.opts.clientName,
          mode: this.opts.mode,
        },
      });

      if (helloResponse.ok) {
        console.log("[PicoClaw] Connected successfully:", helloResponse.payload);
        this.connected = true;
        this.opts.onHello?.(helloResponse.payload as any);
        
        // Start polling for events
        console.log("[PicoClaw] Starting event polling...");
        this.startPolling();
      } else {
        console.error("[PicoClaw] Connect failed:", helloResponse.error);
      }
    } catch (error) {
      console.error("[PicoClaw] Connection failed:", error);
      this.opts.onClose?.({ code: 1006, reason: "Connection failed" });
    }
  }

  private async startPolling() {
    if (this.closed || !this.connected || this.polling) return;
    
    this.polling = true;

    try {
      // Initial poll to get all events
      console.log("[PicoClaw] Initial polling, starting from seq 0");
      
      const initialResponse = await this.makeDirectRequest("events.poll", {
        lastSeq: 0,
      });
      
      if (initialResponse.ok && initialResponse.payload) {
        const events = (initialResponse.payload as any).events || [];
        console.log("[PicoClaw] Initial poll received", events.length, "events");
        
        // Process initial events to set correct lastSeq
        for (const event of events) {
          if (event.seq > this.lastSeq) {
            this.lastSeq = event.seq;
          }
          
          console.log("[PicoClaw] Processing initial event:", event.event);
          // Mark initial events as historical by adding a flag
          const payload = event.payload ? {
            ...event.payload,
            _isHistorical: true
          } : event.payload;
          
          this.opts.onEvent?.({
            type: event.type || "event",
            event: event.event,
            payload: payload,
            seq: event.seq,
          });
        }
      }

      // Continue with regular polling
      while (!this.closed && this.connected) {
        console.log("[PicoClaw] Polling events, lastSeq:", this.lastSeq);
        
        try {
          const response = await this.makeDirectRequest("events.poll", {
            lastSeq: this.lastSeq,
          });

          if (response.ok && response.payload) {
            const events = (response.payload as any).events || [];
            console.log("[PicoClaw] Received", events.length, "events");
            for (const event of events) {
              if (event.seq > this.lastSeq) {
                this.lastSeq = event.seq;
              }
              
              console.log("[PicoClaw] Processing event:", event.event, event.payload);
              // Convert to the format expected by the frontend
              this.opts.onEvent?.({
                type: event.type || "event",
                event: event.event,
                payload: event.payload,
                seq: event.seq,
              });
            }
          } else {
            console.log("[PicoClaw] No events in response");
          }
        } catch (error) {
          console.error("[PicoClaw] Event polling failed:", error);
        }

        // Wait before next poll (3 seconds)
        await new Promise(resolve => setTimeout(resolve, 3000));
      }
    } finally {
      this.polling = false;
    }
  }

  private async makeDirectRequest<T = any>(method: string, params?: any): Promise<GatewayResponseFrame> {
    if (this.closed) {
      throw new Error("Adapter is stopped");
    }

    let httpUrl = this.opts.url.replace(/^ws:\/\//, "http://").replace(/^wss:\/\//, "https://");
    if (!httpUrl.endsWith("/api")) {
      httpUrl = httpUrl.replace(/\/gateway$/, "") + "/api";
    }

    const messageId = `req_${++this.messageId}`;
    const postData = JSON.stringify({
      id: messageId,
      method: method,
      params: params,
    });

    console.log("[PicoClaw] Making direct request:", method, "params:", JSON.stringify(params));
    console.log("[PicoClaw] Request URL:", httpUrl);
    console.log("[PicoClaw] Request body:", postData);

    try {
      const response = await fetch(httpUrl, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: postData,
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      console.log("[PicoClaw] Direct response:", JSON.stringify(data));

      return data as GatewayResponseFrame;
    } catch (error) {
      console.error("[PicoClaw] Direct request failed:", error);
      throw error;
    }
  }

  async request<T = any>(method: string, params?: any): Promise<GatewayResponseFrame> {
    if (this.closed) {
      throw new Error("Adapter is stopped");
    }

    // Auto-connect if not connected
    if (!this.connected) {
      console.log("[PicoClaw] Auto-connecting before request:", method);
      let httpUrl = this.opts.url.replace(/^ws:\/\//, "http://").replace(/^wss:\/\//, "https://");
      if (!httpUrl.endsWith("/api")) {
        httpUrl = httpUrl.replace(/\/gateway$/, "") + "/api";
      }
      await this.connect(httpUrl);
    }

    const messageId = String(++this.messageId);
    
    let httpUrl = this.opts.url.replace(/^ws:\/\//, "http://").replace(/^wss:\/\//, "https://");
    if (!httpUrl.endsWith("/api")) {
      httpUrl = httpUrl.replace(/\/gateway$/, "") + "/api";
    }

    const response = await fetch(httpUrl, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        id: messageId,
        method,
        params,
      }),
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    const result = await response.json() as GatewayResponseFrame;
    console.log("[PicoClaw] Request response:", result);
    
    if (!result.ok) {
      console.error("[PicoClaw] Request failed:", result.error);
    }
    
    return result;
  }

  sendEvent(event: string, payload?: any) {
    // PicoClaw doesn't support client-side events via HTTP
    console.log("Event sent (not supported by HTTP adapter):", event, payload);
  }
}

// Mock for the original GatewayBrowserClient interface
export function createPicoClawClient(opts: GatewayBrowserClientOptions) {
  console.log("[PicoClaw] Creating PicoClaw client with URL:", opts.url);
  const adapter = new PicoClawAdapter(opts);
  
  return {
    start: () => {
      console.log("[PicoClaw] Starting PicoClaw adapter...");
      return adapter.start();
    },
    stop: () => {
      console.log("[PicoClaw] Stopping PicoClaw adapter...");
      return adapter.stop();
    },
    get connected() { 
      console.log("[PicoClaw] Connection status check:", adapter.isConnectionOpen);
      return adapter.isConnectionOpen; 
    },
    request: <T = any>(method: string, params?: any) => {
      console.log("[PicoClaw] Making request:", method, params);
      return adapter.request<T>(method, params);
    },
    sendEvent: (event: string, payload?: any) => adapter.sendEvent(event, payload),
  };
}