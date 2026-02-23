export interface Message {
  id: string;
  content: string;
  role: 'user' | 'assistant';
  timestamp: Date;
  model?: string;
}

export interface Model {
  model_name: string;
  model: string;
  api_key?: string;
  api_base?: string;
  provider?: string;
}

export interface Config {
  model_list: Model[];
  agents: {
    defaults: {
      model: string;
      max_tokens?: number;
      temperature?: number;
      max_tool_iterations?: number;
    };
  };
}

export interface ChatState {
  messages: Message[];
  isLoading: boolean;
  selectedModel: string;
  models: Model[];
}

export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
}

export interface ChatRequest {
  message: string;
  model: string;
  stream?: boolean;
}

export interface ChatResponse {
  message: string;
  model: string;
  timestamp: Date;
}