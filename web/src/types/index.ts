export interface Message {
  id: string;
  content: string;
  role: 'user' | 'assistant';
  timestamp: Date;
  model?: string;
  thoughts?: Thought[];
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
  thoughts?: Thought[];
}

export interface Thought {
  type: 'tool_call' | 'tool_result' | 'thinking';
  timestamp: Date;
  content: string;
  tool_name?: string;
  args?: string;
  result?: string;
  duration?: number;
}

export interface Skill {
  name: string;
  path: string;
  source: string;
  description: string;
}

export interface SkillDetail extends Skill {
  content: string;
  metadata: Record<string, string>;
}

export interface SearchSkillsRequest {
  query: string;
  limit?: number;
}

export interface InstallSkillRequest {
  slug: string;
  registry: string;
  version?: string;
  force?: boolean;
}

export interface SearchSkillsResponse {
  query: string;
  results: any[];
}

export interface InstallSkillResponse {
  status: 'success' | 'error';
  message: string;
  result?: string;
}