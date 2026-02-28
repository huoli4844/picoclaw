import { useState, useCallback, useEffect, useRef } from 'react'

// Toast通知函数
const showToast = (message: string, type: 'success' | 'error' | 'info' = 'info') => {
  if (window.dispatchEvent) {
    window.dispatchEvent(new CustomEvent('toast', { detail: { message, type } }))
  }
}

// 简单的防抖函数实现
function debounce<T extends (...args: any[]) => any>(func: T, wait: number): T {
  let timeout: ReturnType<typeof setTimeout>
  return ((...args: Parameters<T>) => {
    clearTimeout(timeout)
    timeout = setTimeout(() => func(...args), wait)
  }) as T
}
import { Conversation, Message, CreateConversationRequest, UpdateConversationRequest } from '../types/conversation'
import { useApi } from './useApi'
import { ChatResponse } from '../types'

interface UseConversationsProps {
  selectedModel?: string
}

interface UseConversationsReturn {
  conversations: Conversation[]
  activeConversationId: string
  activeConversation: Conversation | undefined
  isLoading: boolean
  isSaving: (conversationId: string) => boolean
  
  createConversation: () => Promise<string>
  selectConversation: (id: string) => Promise<void>
  loadConversation: (id: string) => Promise<void>
  closeConversation: (id: string) => void
  deleteConversation: (id: string) => void
  renameConversation: (id: string, newTitle: string) => void
  sendMessage: (content: string) => Promise<void>
}

export function useConversations({ selectedModel = 'gpt-4' }: UseConversationsProps = {}): UseConversationsReturn {
  const { sendStreamingChatMessage, getConversation, createConversation: createConversationApi, updateConversation, deleteConversation: deleteConversationApi } = useApi()
  const [conversations, setConversations] = useState<Conversation[]>([])
  const [activeConversationId, setActiveConversationId] = useState<string>('')
  const [isLoading, setIsLoading] = useState(true)
  const [savingConversationIds, setSavingConversationIds] = useState<Set<string>>(new Set())

  // 创建新对话的后端版本
  const createNewConversation = useCallback(async (): Promise<string> => {
    try {
      const request: CreateConversationRequest = {
        model: selectedModel
      }
      const result = await createConversationApi(request)
      if (result.success && result.data) {
        const newConv = {
          ...result.data,
          createdAt: new Date(result.data.createdAt),
          updatedAt: new Date(result.data.updatedAt)
        }
        setConversations(prev => [newConv, ...prev])
        setActiveConversationId(newConv.id)
        return newConv.id
      } else {
        throw new Error(result.error || 'Failed to create conversation')
      }
    } catch (error) {
      console.error('Failed to create conversation:', error)
      throw error
    }
  }, [createConversationApi, selectedModel])

  // 启动时不自动创建对话，等待用户操作
  useEffect(() => {
    const initializeApp = async () => {
      try {
        // 不自动创建对话，只设置加载完成
        setIsLoading(false)
      } catch (error) {
        console.error('Failed to initialize app:', error)
        setIsLoading(false)
      }
    }

    initializeApp()
  }, [])

  // 实时保存对话到后端
  const saveConversationToBackend = useCallback(async (conversationId: string, conversationData?: Conversation) => {
    try {
      // 使用传入的conversationData或从最新状态中获取
      let conversation: Conversation | undefined
      if (conversationData) {
        conversation = conversationData
      } else {
        // 使用函数式更新来获取最新状态
        setConversations(prev => {
          const conv = prev.find(c => c.id === conversationId)
          if (!conv) return prev
          conversation = conv
          return prev
        })
      }
      
      if (!conversation) {
        console.warn(`Conversation ${conversationId} not found for saving`)
        return
      }

      // 添加到保存状态
      setSavingConversationIds(prev => new Set([...prev, conversationId]))

      const request: UpdateConversationRequest = {
        messages: conversation.messages,
        title: conversation.title
      }
      
      const result = await updateConversation(conversationId, request)
      
      // 保存完成，从保存状态中移除
      setSavingConversationIds(prev => {
        const newSet = new Set(prev)
        newSet.delete(conversationId)
        return newSet
      })
      
      if (result.success) {
        console.log(`对话 "${conversation.title}" 自动保存成功`)
        // 可选：显示成功通知（为了避免过于频繁，这里注释掉）
        // showToast(`对话 "${conversation.title}" 已保存`, 'success')
      } else {
        throw new Error(result.error || 'Save failed')
      }
    } catch (error) {
      console.error('Failed to save conversation to backend:', error)
      
      // 保存失败，从保存状态中移除
      setSavingConversationIds(prev => {
        const newSet = new Set(prev)
        newSet.delete(conversationId)
        return newSet
      })
      
      // 显示错误通知
      const errorMessage = error instanceof Error ? error.message : '保存失败'
      showToast(`保存对话失败: ${errorMessage}`, 'error')
    }
  }, [updateConversation])

  // 保存状态管理
  const saveTimeoutsRef = useRef<Map<string, ReturnType<typeof setTimeout>>>(new Map())

  // 优化的防抖保存机制
  const debouncedSaveConversation = useCallback((conversationId: string, conversation?: Conversation) => {
    // 清除之前的超时
    const existingTimeout = saveTimeoutsRef.current.get(conversationId)
    if (existingTimeout) {
      clearTimeout(existingTimeout)
    }
    
    // 设置新的超时
    const timeout = setTimeout(() => {
      saveConversationToBackend(conversationId, conversation)
      saveTimeoutsRef.current.delete(conversationId)
    }, 1000)
    
    saveTimeoutsRef.current.set(conversationId, timeout)
  }, [saveConversationToBackend])

  // 立即保存（用于重要操作）
  const immediateSaveConversation = useCallback((conversationId: string, conversation?: Conversation) => {
    // 清除防抖超时
    const existingTimeout = saveTimeoutsRef.current.get(conversationId)
    if (existingTimeout) {
      clearTimeout(existingTimeout)
      saveTimeoutsRef.current.delete(conversationId)
    }
    
    saveConversationToBackend(conversationId, conversation)
  }, [saveConversationToBackend])

  // 组件卸载时保存所有对话
  useEffect(() => {
    return () => {
      // 清除所有超时并保存对话
      saveTimeoutsRef.current.forEach((timeout, conversationId) => {
        clearTimeout(timeout)
        saveConversationToBackend(conversationId)
      })
      saveTimeoutsRef.current.clear()
    }
  }, [saveConversationToBackend])

  const createConversation = useCallback(async (): Promise<string> => {
    try {
      const id = await createNewConversation()
      showToast('新对话创建成功', 'success')
      return id
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '创建对话失败'
      showToast(`创建对话失败: ${errorMessage}`, 'error')
      throw error
    }
  }, [createNewConversation])

  const loadConversation = useCallback(async (id: string) => {
    try {
      const result = await getConversation(id)
      if (result.success && result.data) {
        const conversation = {
          ...result.data,
          createdAt: new Date(result.data.createdAt),
          updatedAt: new Date(result.data.updatedAt),
          messages: result.data.messages.map(msg => ({
            ...msg,
            timestamp: new Date(msg.timestamp)
          }))
        }
        
        setConversations(prev => {
          // 检查是否已经存在
          const exists = prev.some(conv => conv.id === id)
          if (exists) {
            // 如果已存在，切换到该对话
            setActiveConversationId(id)
            return prev
          } else {
            // 如果不存在，添加到列表并切换
            setActiveConversationId(id)
            return [conversation, ...prev]
          }
        })
      } else {
        throw new Error(result.error || 'Failed to load conversation')
      }
    } catch (error) {
      console.error('Failed to load conversation:', error)
    }
  }, [getConversation])

  const selectConversation = useCallback(async (id: string) => {
    // 首先切换到选中的对话
    setActiveConversationId(id)
    
    // 检查当前对话是否存在消息内容，如果没有则重新加载
    const currentConversation = conversations.find(conv => conv.id === id)
    if (currentConversation && (!currentConversation.messages || currentConversation.messages.length === 0)) {
      try {
        await loadConversation(id)
      } catch (error) {
        console.error('Failed to load conversation details:', error)
      }
    }
  }, [conversations, loadConversation])

  const closeConversation = useCallback(async (id: string) => {
    let conversationToClose: Conversation | undefined
    
    setConversations(prev => {
      // 先找到要关闭的对话
      const conv = prev.find(conv => conv.id === id)
      if (!conv) return prev
      conversationToClose = conv
      
      const newConversations = prev.filter(conv => conv.id !== id)
      
      // 如果关闭的是当前活跃的对话，切换到第一个对话
      if (id === activeConversationId && newConversations.length > 0) {
        setActiveConversationId(newConversations[0].id)
      }
      
      return newConversations
    })
    
    // 立即保存对话到后端
    if (conversationToClose) {
      immediateSaveConversation(id, conversationToClose)
    }
  }, [activeConversationId, immediateSaveConversation])

  const deleteConversation = useCallback(async (id: string) => {
    try {
      const result = await deleteConversationApi(id)
      if (result.success) {
        setConversations(prev => {
          const conversationToDelete = prev.find(conv => conv.id === id)
          const newConversations = prev.filter(conv => conv.id !== id)
          
          // 如果删除的是当前活跃的对话，切换到第一个对话
          if (id === activeConversationId && newConversations.length > 0) {
            setActiveConversationId(newConversations[0].id)
          }
          
          return newConversations
        })
        
        showToast('对话删除成功', 'success')
      } else {
        throw new Error(result.error || 'Failed to delete conversation')
      }
    } catch (error) {
      console.error('Failed to delete conversation:', error)
      const errorMessage = error instanceof Error ? error.message : '删除对话失败'
      showToast(`删除对话失败: ${errorMessage}`, 'error')
    }
  }, [deleteConversationApi, activeConversationId])

  const renameConversation = useCallback(async (id: string, newTitle: string) => {
    try {
      const request: UpdateConversationRequest = { title: newTitle }
      const result = await updateConversation(id, request)
      if (result.success && result.data) {
        let updatedConv: Conversation | undefined
        
        setConversations(prev =>
          prev.map(conv => {
            if (conv.id === id) {
              updatedConv = { 
                ...conv, 
                title: newTitle, 
                updatedAt: new Date(result.data?.updatedAt || new Date()) 
              }
              return updatedConv
            }
            return conv
          })
        )
        
        showToast('对话重命名成功', 'success')
      } else {
        throw new Error(result.error || 'Failed to rename conversation')
      }
    } catch (error) {
      console.error('Failed to rename conversation:', error)
      const errorMessage = error instanceof Error ? error.message : '重命名对话失败'
      showToast(`重命名对话失败: ${errorMessage}`, 'error')
    }
  }, [updateConversation])



  const addMessage = useCallback((conversationId: string, message: Message) => {
    let updatedConversation: Conversation | undefined
    
    setConversations(prev =>
      prev.map(conv => {
        if (conv.id === conversationId) {
          updatedConversation = { 
            ...conv, 
            messages: [...conv.messages, message],
            updatedAt: new Date()
          }
          return updatedConversation
        }
        return conv
      })
    )
    
    // 立即触发防抖保存，传入更新后的对话数据
    if (updatedConversation) {
      debouncedSaveConversation(conversationId, updatedConversation)
    }
  }, [debouncedSaveConversation])

  const updateMessage = useCallback((conversationId: string, messageId: string, updates: Partial<Message>) => {
    let updatedConversation: Conversation | undefined
    
    setConversations(prev =>
      prev.map(conv => {
        if (conv.id === conversationId) {
          updatedConversation = {
            ...conv,
            messages: conv.messages.map(msg =>
              msg.id === messageId ? { ...msg, ...updates } : msg
            ),
            updatedAt: new Date()
          }
          return updatedConversation
        }
        return conv
      })
    )
    
    // 立即触发防抖保存，传入更新后的对话数据
    if (updatedConversation) {
      debouncedSaveConversation(conversationId, updatedConversation)
    }
  }, [debouncedSaveConversation])



  const sendMessage = useCallback(async (content: string) => {
    if (!activeConversationId) return

    setIsLoading(true)

    // 获取当前对话的模型，如果没有则使用全局选中的模型
    const currentConversationModel = conversations.find(c => c.id === activeConversationId)?.model || selectedModel
    
    // 创建用户消息
    const userMessage: Message = {
      id: Date.now().toString(),
      content,
      role: 'user',
      timestamp: new Date(),
      model: currentConversationModel
    }

    addMessage(activeConversationId, userMessage)

    // 创建助手消息
    const assistantMessageId = (Date.now() + 1).toString()
    const assistantMessage: Message = {
      id: assistantMessageId,
      content: '',
      role: 'assistant',
      timestamp: new Date(),
      model: currentConversationModel,
      thoughts: []
    }

    addMessage(activeConversationId, assistantMessage)

    try {
      const currentConv = conversations.find(c => c.id === activeConversationId)
      const modelToUse = currentConv?.model || selectedModel
      
      await sendStreamingChatMessage(
        {
          message: content,
          model: modelToUse,
          stream: true,
          conversationId: activeConversationId
        },
        // onThought
        (thought: any) => {
          let updatedConv: Conversation | undefined
          
          setConversations(prev => {
            const updated = prev.map(conv => {
              if (conv.id === activeConversationId) {
                updatedConv = {
                  ...conv,
                  messages: conv.messages.map(msg =>
                    msg.id === assistantMessageId
                      ? { 
                          ...msg, 
                          thoughts: [...(msg.thoughts || []), thought]
                        }
                      : msg
                  ),
                  updatedAt: new Date()
                }
                return updatedConv
              }
              return conv
            })
            
            return updated
          })
          
          // 在思考过程中触发保存，使用更长的防抖时间避免频繁保存
          if (updatedConv) {
            setTimeout(() => {
              debouncedSaveConversation(activeConversationId, updatedConv)
            }, 2000)
          }
        },
        // onComplete
        async (response: ChatResponse) => {
          let updatedConv: Conversation | undefined
          
          setConversations(prev => {
            const newConversations = prev.map(conv => {
              if (conv.id === activeConversationId) {
                updatedConv = {
                  ...conv,
                  messages: conv.messages.map(msg =>
                    msg.id === assistantMessageId
                      ? { ...msg, content: response.message, timestamp: response.timestamp }
                      : msg
                  ),
                  updatedAt: new Date()
                }
                return updatedConv
              }
              return conv
            })
            return newConversations
          })

          // 如果是第一条消息，自动生成标题
          if (updatedConv && updatedConv.messages.length <= 2 && updatedConv.title.startsWith('新对话')) {
            const title = content.length > 30 
              ? content.substring(0, 30) + '...' 
              : content
            
            // 立即保存包含新标题的对话
            immediateSaveConversation(activeConversationId, {
              ...updatedConv,
              title
            })
          } else if (updatedConv) {
            // 立即保存完成的对话
            immediateSaveConversation(activeConversationId, updatedConv)
          }
        },
        // onError
        (error: string) => {
          setConversations(prev =>
            prev.map(conv =>
              conv.id === activeConversationId
                ? {
                    ...conv,
                    messages: conv.messages.map(msg =>
                      msg.id === assistantMessageId
                        ? { ...msg, content: `错误: ${error}` }
                        : msg
                    ),
                    updatedAt: new Date()
                  }
                : conv
            )
          )
        }
      )
    } catch (error) {
      updateMessage(activeConversationId, assistantMessageId, {
        content: `网络错误: ${error instanceof Error ? error.message : '未知错误'}`
      })
    } finally {
      setIsLoading(false)
    }
  }, [activeConversationId, conversations, sendStreamingChatMessage, addMessage, updateMessage, renameConversation])

  const activeConversation = conversations.find(conv => conv.id === activeConversationId)

  return {
    conversations,
    activeConversationId,
    activeConversation,
    isLoading,
    isSaving: (conversationId: string) => savingConversationIds.has(conversationId),
    createConversation,
    selectConversation,
    loadConversation,
    closeConversation,
    deleteConversation,
    renameConversation,
    sendMessage
  }
}