import { useState, useCallback, useEffect, useRef } from 'react'
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

  // 实时保存对话到后端 - 确保每个对话独立保存
  const saveConversationToBackend = useCallback(async (conversationId: string, currentConversation?: Conversation) => {
    try {
      let conversation: Conversation | undefined
      
      if (currentConversation) {
        // 如果传入了当前对话数据，直接使用
        conversation = currentConversation
      } else {
        // 否则从最新状态中获取指定对话
        conversation = conversations.find(conv => conv.id === conversationId)
      }
      
      if (!conversation) {
        console.warn(`Conversation ${conversationId} not found for saving`)
        return
      }

      console.log(`Saving conversation ${conversationId}:`, {
        title: conversation.title,
        messageCount: conversation.messages.length,
        messages: conversation.messages.map(m => ({ id: m.id, role: m.role, content: m.content.substring(0, 50) + '...' }))
      })

      const request: UpdateConversationRequest = {
        messages: conversation.messages,
        title: conversation.title
      }
      
      await updateConversation(conversationId, request)
      console.log(`Successfully saved conversation ${conversationId} with ${conversation.messages.length} messages`)
    } catch (error) {
      console.error(`Failed to save conversation ${conversationId} to backend:`, error)
    }
  }, [conversations, updateConversation])

  // 为每个对话创建独立的防抖器，避免跨对话冲突
  const saveTimeoutsRef = useRef<Map<string, ReturnType<typeof setTimeout>>>(new Map())
  
  const debouncedSaveConversation = useCallback((conversationId: string, currentConversation?: Conversation) => {
    // 清除该对话之前的超时
    const existingTimeout = saveTimeoutsRef.current.get(conversationId)
    if (existingTimeout) {
      clearTimeout(existingTimeout)
    }
    
    // 设置新的超时
    const timeout = setTimeout(() => {
      saveConversationToBackend(conversationId, currentConversation)
      saveTimeoutsRef.current.delete(conversationId)
    }, 1000)
    
    saveTimeoutsRef.current.set(conversationId, timeout)
  }, [saveConversationToBackend])

  // 组件卸载时保存所有对话并清理防抖器
  useEffect(() => {
    return () => {
      // 清除所有防抖超时
      saveTimeoutsRef.current.forEach((timeout, conversationId) => {
        clearTimeout(timeout)
        // 立即保存该对话
        const conversation = conversations.find(c => c.id === conversationId)
        if (conversation) {
          saveConversationToBackend(conversationId, conversation)
        }
      })
      saveTimeoutsRef.current.clear()
    }
  }, [conversations, saveConversationToBackend])

  const createConversation = useCallback(async (): Promise<string> => {
    return await createNewConversation()
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
    // 先找到要关闭的对话
    const conversationToClose = conversations.find(conv => conv.id === id)
    if (!conversationToClose) return

    // 清除该对话的防抖器并立即保存
    const existingTimeout = saveTimeoutsRef.current.get(id)
    if (existingTimeout) {
      clearTimeout(existingTimeout)
      saveTimeoutsRef.current.delete(id)
    }
    
    // 立即保存该对话
    await saveConversationToBackend(id, conversationToClose)

    // 从列表中移除该对话
    setConversations(prev => {
      const newConversations = prev.filter(conv => conv.id !== id)
      
      // 如果关闭的是当前活跃的对话，切换到第一个对话
      if (id === activeConversationId && newConversations.length > 0) {
        setActiveConversationId(newConversations[0].id)
      } else if (newConversations.length === 0) {
        // 如果没有其他对话，清空活跃对话ID
        setActiveConversationId('')
      }
      
      return newConversations
    })
  }, [conversations, activeConversationId, saveConversationToBackend])

  const deleteConversation = useCallback(async (id: string) => {
    try {
      const result = await deleteConversationApi(id)
      if (result.success) {
        setConversations(prev => {
          const newConversations = prev.filter(conv => conv.id !== id)
          
          // 如果删除的是当前活跃的对话，切换到第一个对话
          if (id === activeConversationId && newConversations.length > 0) {
            setActiveConversationId(newConversations[0].id)
          }
          
          return newConversations
        })
      } else {
        throw new Error(result.error || 'Failed to delete conversation')
      }
    } catch (error) {
      console.error('Failed to delete conversation:', error)
    }
  }, [deleteConversationApi, activeConversationId])

  const renameConversation = useCallback(async (id: string, newTitle: string) => {
    try {
      const request: UpdateConversationRequest = { title: newTitle }
      const result = await updateConversation(id, request)
      if (result.success && result.data) {
        setConversations(prev =>
          prev.map(conv =>
            conv.id === id
              ? { ...conv, title: newTitle, updatedAt: new Date(result.data?.updatedAt || new Date()) }
              : conv
          )
        )
      } else {
        throw new Error(result.error || 'Failed to rename conversation')
      }
    } catch (error) {
      console.error('Failed to rename conversation:', error)
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
    
    // 触发实时保存，传入更新后的对话数据
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
    
    // 触发实时保存，传入更新后的对话数据
    if (updatedConversation) {
      debouncedSaveConversation(conversationId, updatedConversation)
    }
  }, [debouncedSaveConversation])



  const sendMessage = useCallback(async (content: string) => {
    if (!activeConversationId) return

    setIsLoading(true)

    // 获取当前对话的模型，如果没有则使用全局选中的模型
    const currentConversationModel = conversations.find(c => c.id === activeConversationId)?.model || selectedModel
    
    // 创建唯一的消息ID，包含对话ID和随机数
    const generateMessageId = (role: string) => {
      return `${activeConversationId}_${role}_${Date.now()}_${Math.random().toString(36).substring(2, 11)}`
    }
    
    // 创建用户消息
    const userMessage: Message = {
      id: generateMessageId('user'),
      content,
      role: 'user',
      timestamp: new Date(),
      model: currentConversationModel
    }

    addMessage(activeConversationId, userMessage)

    // 创建助手消息
    const assistantMessage: Message = {
      id: generateMessageId('assistant'),
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
          let updatedConversation: Conversation | undefined
          
          setConversations(prev => {
            const updated = prev.map(conv => {
              if (conv.id === activeConversationId) {
                updatedConversation = {
                    ...conv,
                    messages: conv.messages.map(msg =>
                      msg.id === assistantMessage.id
                        ? { 
                            ...msg, 
                            thoughts: [...(msg.thoughts || []), thought]
                          }
                        : msg
                    ),
                    updatedAt: new Date()
                  }
                return updatedConversation
              }
              return conv
            })
            
            return updated
          })
          
          // 在思考过程中也触发保存，使用更长的防抖时间
          if (updatedConversation) {
            setTimeout(() => {
              saveConversationToBackend(activeConversationId, updatedConversation)
            }, 2000) // 2秒后保存
          }
        },
        // onComplete
        async (response: ChatResponse) => {
          let finalConversation: Conversation | undefined
          
          setConversations(prev => {
            const newConversations = prev.map(conv => {
              if (conv.id === activeConversationId) {
                finalConversation = {
                  ...conv,
                  messages: conv.messages.map(msg =>
                    msg.id === assistantMessage.id
                      ? { ...msg, content: response.message, timestamp: response.timestamp }
                      : msg
                  ),
                  updatedAt: new Date()
                }
                return finalConversation
              }
              return conv
            })
            return newConversations
          })

          // 立即保存完成的对话
          if (finalConversation) {
            try {
              // 如果是新对话的第一条消息，使用用户消息的前30个字符作为标题
              if (finalConversation.messages.length <= 2 && finalConversation.title === '新对话') {
                const title = content.length > 30 
                  ? content.substring(0, 30) + '...' 
                  : content
                
                // 更新标题并保存
                const updatedWithTitle = { ...finalConversation, title }
                await updateConversation(activeConversationId, {
                  title,
                  messages: updatedWithTitle.messages
                })
                
                // 更新本地状态
                setConversations(prev =>
                  prev.map(conv =>
                    conv.id === activeConversationId ? updatedWithTitle : conv
                  )
                )
              } else {
                // 保存对话到后端
                await saveConversationToBackend(activeConversationId, finalConversation)
              }
            } catch (error) {
              console.error('Failed to save conversation:', error)
            }
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
                      msg.id === assistantMessage.id
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
      updateMessage(activeConversationId, assistantMessage.id, {
        content: `网络错误: ${error instanceof Error ? error.message : '未知错误'}`
      })
    } finally {
      setIsLoading(false)
    }
  }, [activeConversationId, conversations, sendStreamingChatMessage, addMessage, updateMessage, renameConversation, saveConversationToBackend, debouncedSaveConversation])

  const activeConversation = conversations.find(conv => conv.id === activeConversationId)

  return {
    conversations,
    activeConversationId,
    activeConversation,
    isLoading,
    createConversation,
    selectConversation,
    loadConversation,
    closeConversation,
    deleteConversation,
    renameConversation,
    sendMessage
  }
}