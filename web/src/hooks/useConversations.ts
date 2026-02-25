import { useState, useCallback, useEffect } from 'react'

// 简单的防抖函数实现
function debounce<T extends (...args: any[]) => any>(func: T, wait: number): T {
  let timeout: NodeJS.Timeout
  return ((...args: Parameters<T>) => {
    clearTimeout(timeout)
    timeout = setTimeout(() => func(...args), wait)
  }) as T
}
import { Conversation, Message, CreateConversationRequest, UpdateConversationRequest } from '../types/conversation'
import { useApi } from './useApi'
import { ChatResponse, Thought } from '../types'

interface UseConversationsProps {
  selectedModel?: string
}

interface UseConversationsReturn {
  conversations: Conversation[]
  activeConversationId: string
  activeConversation: Conversation | undefined
  isLoading: boolean
  
  createConversation: () => string
  selectConversation: (id: string) => Promise<void>
  loadConversation: (id: string) => Promise<void>
  closeConversation: (id: string) => void
  deleteConversation: (id: string) => void
  renameConversation: (id: string, newTitle: string) => void
  sendMessage: (content: string) => Promise<void>
}

export function useConversations({ selectedModel = 'gpt-4' }: UseConversationsProps = {}): UseConversationsReturn {
  const { sendStreamingChatMessage, getConversations, getConversation, createConversation: createConversationApi, updateConversation, deleteConversation: deleteConversationApi } = useApi()
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

  // 启动时加载历史对话
  useEffect(() => {
    const loadConversations = async () => {
      try {
        const result = await getConversations()
        if (result.success && result.data) {
          // 转换日期字符串为Date对象
          const conversationsWithDates = result.data.map(conv => ({
            ...conv,
            createdAt: new Date(conv.createdAt),
            updatedAt: new Date(conv.updatedAt),
            messages: conv.messages.map(msg => ({
              ...msg,
              timestamp: new Date(msg.timestamp)
            }))
          }))
          setConversations(conversationsWithDates)
          
          // 如果有对话，选择最新的一个
          if (conversationsWithDates.length > 0) {
            setActiveConversationId(conversationsWithDates[0].id)
          } else {
            // 如果没有历史对话，创建一个新的
            await createNewConversation()
          }
        }
      } catch (error) {
        console.error('Failed to load conversations:', error)
      } finally {
        setIsLoading(false)
      }
    }

    loadConversations()
  }, [getConversations, createNewConversation])

  // 实时保存对话到后端
  const saveConversationToBackend = useCallback(async (conversationId: string) => {
    try {
      const conversation = conversations.find(conv => conv.id === conversationId)
      if (!conversation) return

      const request: UpdateConversationRequest = {
        messages: conversation.messages,
        title: conversation.title
      }
      await updateConversation(conversationId, request)
    } catch (error) {
      console.error('Failed to save conversation to backend:', error)
    }
  }, [conversations, updateConversation])

  // 使用防抖来避免频繁保存
  const debouncedSaveConversation = useCallback(
    debounce((conversationId: string) => {
      saveConversationToBackend(conversationId)
    }, 1000), // 1秒防抖
    [saveConversationToBackend]
  )

  // 组件卸载时保存所有对话
  useEffect(() => {
    return () => {
      // 保存所有修改过的对话
      conversations.forEach(conv => {
        saveConversationToBackend(conv.id)
      })
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

  const closeConversation = useCallback((id: string) => {
    setConversations(prev => {
      const newConversations = prev.filter(conv => conv.id !== id)
      
      // 如果关闭的是当前活跃的对话，切换到第一个对话
      if (id === activeConversationId && newConversations.length > 0) {
        setActiveConversationId(newConversations[0].id)
      }
      
      return newConversations
    })
  }, [activeConversationId])

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
              ? { ...conv, title: newTitle, updatedAt: new Date(result.data.updatedAt) }
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
    setConversations(prev =>
      prev.map(conv =>
        conv.id === conversationId
          ? { 
              ...conv, 
              messages: [...conv.messages, message],
              updatedAt: new Date()
            }
          : conv
      )
    )
    
    // 触发实时保存
    debouncedSaveConversation(conversationId)
  }, [debouncedSaveConversation])

  const updateMessage = useCallback((conversationId: string, messageId: string, updates: Partial<Message>) => {
    setConversations(prev =>
      prev.map(conv =>
        conv.id === conversationId
          ? {
              ...conv,
              messages: conv.messages.map(msg =>
                msg.id === messageId ? { ...msg, ...updates } : msg
              ),
              updatedAt: new Date()
            }
          : conv
      )
    )
    
    // 触发实时保存
    debouncedSaveConversation(conversationId)
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
          setConversations(prev => {
            const updated = prev.map(conv =>
              conv.id === activeConversationId
                ? {
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
                : conv
            )
            
            // 在思考过程中也触发保存，但使用更长的防抖时间
            setTimeout(() => {
              saveConversationToBackend(activeConversationId)
            }, 2000) // 2秒后保存
            
            return updated
          })
        },
        // onComplete
        async (response: ChatResponse) => {
          setConversations(prev => {
            const newConversations = prev.map(conv => {
              if (conv.id === activeConversationId) {
                const updatedConv = {
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

          // 异步保存对话到后端
          try {
            const currentConv = conversations.find(c => c.id === activeConversationId)
            if (currentConv) {
              const updatedConv = {
                ...currentConv,
                messages: currentConv.messages.map(msg =>
                  msg.id === assistantMessageId
                    ? { ...msg, content: response.message, timestamp: response.timestamp }
                    : msg
                ),
                updatedAt: new Date()
              }

              // 如果是新对话的第一条消息，使用用户消息的前30个字符作为标题
              if (currentConv.messages.length <= 2 && currentConv.title === '新对话') {
                const title = content.length > 30 
                  ? content.substring(0, 30) + '...' 
                  : content
                const request: UpdateConversationRequest = {
                  title,
                  messages: updatedConv.messages
                }
                await updateConversation(activeConversationId, request)
              } else {
                // 保存对话到后端
                const request: UpdateConversationRequest = {
                  messages: updatedConv.messages
                }
                await updateConversation(activeConversationId, request)
              }
            }
          } catch (error) {
            console.error('Failed to save conversation:', error)
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
    createConversation,
    selectConversation,
    loadConversation,
    closeConversation,
    deleteConversation,
    renameConversation,
    sendMessage
  }
}