import { useState, useCallback } from 'react'
import { 
  ApiResponse, 
  ChatRequest, 
  ChatResponse, 
  Model, 
  Config, 
  Skill, 
  SkillDetail,
  SearchSkillsRequest, 
  InstallSkillRequest,
  SearchSkillsResponse,
  InstallSkillResponse,
  McpServer,
  McpSearchRequest,
  McpSearchResponse,
  McpInstallRequest,
  McpInstallResponse
} from '@/types'

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api'

export function useApi() {
  const [isLoading, setIsLoading] = useState(false)

  const request = useCallback(async <T = any>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> => {
    try {
      const response = await fetch(`${API_BASE_URL}${endpoint}`, {
        headers: {
          'Content-Type': 'application/json',
          ...options.headers,
        },
        ...options,
      })

      const data = await response.json()

      if (!response.ok) {
        return {
          success: false,
          error: data.error || `HTTP error! status: ${response.status}`,
        }
      }

      return {
        success: true,
        data,
      }
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error',
      }
    }
  }, [])

  const sendChatMessage = useCallback(async (
    chatRequest: ChatRequest
  ): Promise<ApiResponse<ChatResponse>> => {
    setIsLoading(true)
    try {
      const result = await request<ChatResponse>('/chat', {
        method: 'POST',
        body: JSON.stringify(chatRequest),
      })
      return result
    } finally {
      setIsLoading(false)
    }
  }, [request])

  const sendStreamingChatMessage = useCallback(
    (
      chatRequest: ChatRequest,
      onThought: (thought: any) => void,
      onComplete: (response: ChatResponse) => void,
      onError: (error: string) => void
    ) => {
      setIsLoading(true)
      
      const streamingRequest = { ...chatRequest, stream: true }
      
      fetch(`${API_BASE_URL}/chat`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(streamingRequest),
      })
      .then(response => {
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`)
        }
        
        const reader = response.body?.getReader()
        if (!reader) {
          throw new Error('Response body is not readable')
        }
        
        const decoder = new TextDecoder()
        let buffer = ''
        
        const readStream = async () => {
          try {
            while (true) {
              const { done, value } = await reader.read()
              if (done) {
                // 流结束，确保loading状态被重置
                setIsLoading(false)
                break
              }
              
              buffer += decoder.decode(value, { stream: true })
              const lines = buffer.split('\n')
              buffer = lines.pop() || ''
              
              for (const line of lines) {
                if (line.startsWith('data: ')) {
                  try {
                    const data = JSON.parse(line.slice(6))
                    
                    switch (data.type) {
                      case 'thought':
                        onThought(data.thought)
                        break
                      case 'complete':
                        console.log('Received complete event:', data)
                        onComplete({
                          message: data.message,
                          model: data.model,
                          timestamp: new Date(data.timestamp),
                          thoughts: [] // 思考过程已经通过 onThought 实时推送
                        })
                        setIsLoading(false)
                        return
                      default:
                        console.warn('Unknown SSE event type:', data.type)
                    }
                  } catch (parseError) {
                    console.error('Error parsing SSE message:', parseError, line)
                  }
                }
              }
            }
          } catch (error) {
            console.error('Stream reading error:', error)
            onError('读取服务器响应失败')
            setIsLoading(false)
          }
        }
        
        readStream()
      })
      .catch(error => {
        console.error('Streaming request error:', error)
        onError('连接服务器失败')
        setIsLoading(false)
      })
      
      // 返回清理函数
      return () => {
        setIsLoading(false)
      }
    },
    []
  )

  const getConfig = useCallback(async (): Promise<ApiResponse<Config>> => {
    return request<Config>('/config')
  }, [request])

  const updateConfig = useCallback(async (
    config: Partial<Config>
  ): Promise<ApiResponse> => {
    return request('/config', {
      method: 'PUT',
      body: JSON.stringify(config),
    })
  }, [request])

  const getModels = useCallback(async (): Promise<ApiResponse<Model[]>> => {
    return request<Model[]>('/models')
  }, [request])

  const getSkills = useCallback(async (): Promise<ApiResponse<Skill[]>> => {
    return request<Skill[]>('/skills')
  }, [request])

  const getSkillDetail = useCallback(async (name: string): Promise<ApiResponse<SkillDetail>> => {
    return request<SkillDetail>(`/skills/${name}`)
  }, [request])

  const searchSkills = useCallback(async (requestObj: SearchSkillsRequest): Promise<ApiResponse<SearchSkillsResponse>> => {
    return request<SearchSkillsResponse>('/skills/search', {
      method: 'POST',
      body: JSON.stringify(requestObj),
    })
  }, [request])

  const installSkill = useCallback(async (requestObj: InstallSkillRequest): Promise<ApiResponse<InstallSkillResponse>> => {
    return request<InstallSkillResponse>('/skills/install', {
      method: 'POST',
      body: JSON.stringify(requestObj),
    })
  }, [request])

  // MCP-related API methods
  const getMcpServers = useCallback(async (): Promise<ApiResponse<McpServer[]>> => {
    return request<McpServer[]>('/mcp/servers')
  }, [request])

  const searchMcpServers = useCallback(async (requestObj: McpSearchRequest): Promise<ApiResponse<McpSearchResponse>> => {
    return request<McpSearchResponse>('/mcp/search', {
      method: 'POST',
      body: JSON.stringify(requestObj),
    })
  }, [request])

  const installMcpServer = useCallback(async (requestObj: McpInstallRequest): Promise<ApiResponse<McpInstallResponse>> => {
    return request<McpInstallResponse>('/mcp/install', {
      method: 'POST',
      body: JSON.stringify(requestObj),
    })
  }, [request])

  const uninstallMcpServer = useCallback(async (serverId: string): Promise<ApiResponse> => {
    return request(`/mcp/servers/${serverId}`, {
      method: 'DELETE',
    })
  }, [request])

  const getMcpServerDetail = useCallback(async (serverId: string): Promise<ApiResponse<McpServer>> => {
    return request<McpServer>(`/mcp/servers/${serverId}`)
  }, [request])

  return {
    isLoading,
    request,
    sendChatMessage,
    sendStreamingChatMessage,
    getConfig,
    updateConfig,
    getModels,
    getSkills,
    getSkillDetail,
    searchSkills,
    installSkill,
    // MCP methods
    getMcpServers,
    searchMcpServers,
    installMcpServer,
    uninstallMcpServer,
    getMcpServerDetail,
  }
}