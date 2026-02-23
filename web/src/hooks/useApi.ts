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
  InstallSkillResponse 
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

  return {
    isLoading,
    request,
    sendChatMessage,
    getConfig,
    updateConfig,
    getModels,
    getSkills,
    getSkillDetail,
    searchSkills,
    installSkill,
  }
}