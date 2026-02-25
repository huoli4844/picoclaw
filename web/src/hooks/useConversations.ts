import { useState, useCallback } from 'react'
import { Conversation, Message } from '../types/conversation'
import { useApi } from './useApi'
import { ChatResponse } from '../types'

interface UseConversationsReturn {
  conversations: Conversation[]
  activeConversationId: string
  activeConversation: Conversation | undefined
  isLoading: boolean
  
  createConversation: () => string
  selectConversation: (id: string) => void
  deleteConversation: (id: string) => void
  renameConversation: (id: string, newTitle: string) => void
  sendMessage: (content: string) => Promise<void>
}

export function useConversations(): UseConversationsReturn {
  const { sendStreamingChatMessage } = useApi()
  const [conversations, setConversations] = useState<Conversation[]>(() => {
    // 初始化时创建一个默认对话
    const defaultConversation: Conversation = {
      id: 'default-' + Date.now(),
      title: '新对话',
      messages: [],
      createdAt: new Date(),
      updatedAt: new Date(),
      model: 'gpt-4'
    }
    return [defaultConversation]
  })
  const [activeConversationId, setActiveConversationId] = useState<string>(() => conversations[0]?.id || '')
  const [isLoading, setIsLoading] = useState(false)

  const createConversation = useCallback(() => {
    const newConversation: Conversation = {
      id: 'conv-' + Date.now(),
      title: `对话 ${conversations.length + 1}`,
      messages: [],
      createdAt: new Date(),
      updatedAt: new Date(),
      model: 'gpt-4' // 默认模型，可以后续配置
    }

    setConversations(prev => [...prev, newConversation])
    setActiveConversationId(newConversation.id)
    return newConversation.id
  }, [conversations.length])

  const selectConversation = useCallback((id: string) => {
    setActiveConversationId(id)
  }, [])

  const deleteConversation = useCallback((id: string) => {
    setConversations(prev => {
      const newConversations = prev.filter(conv => conv.id !== id)
      
      // 如果删除的是当前活跃的对话，切换到第一个对话
      if (id === activeConversationId && newConversations.length > 0) {
        setActiveConversationId(newConversations[0].id)
      }
      
      return newConversations
    })
  }, [activeConversationId])

  const renameConversation = useCallback((id: string, newTitle: string) => {
    setConversations(prev =>
      prev.map(conv =>
        conv.id === id
          ? { ...conv, title: newTitle, updatedAt: new Date() }
          : conv
      )
    )
  }, [])



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
  }, [])

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
  }, [])

  const sendMessage = useCallback(async (content: string) => {
    if (!activeConversationId) return

    setIsLoading(true)

    // 创建用户消息
    const userMessage: Message = {
      id: Date.now().toString(),
      content,
      role: 'user',
      timestamp: new Date(),
      model: conversations.find(c => c.id === activeConversationId)?.model || 'gpt-4'
    }

    addMessage(activeConversationId, userMessage)

    // 创建助手消息
    const assistantMessageId = (Date.now() + 1).toString()
    const assistantMessage: Message = {
      id: assistantMessageId,
      content: '',
      role: 'assistant',
      timestamp: new Date(),
      model: conversations.find(c => c.id === activeConversationId)?.model || 'gpt-4',
      thoughts: []
    }

    addMessage(activeConversationId, assistantMessage)

    try {
      const currentConv = conversations.find(c => c.id === activeConversationId)
      
      await sendStreamingChatMessage(
        {
          message: content,
          model: currentConv?.model || 'gpt-4',
          stream: true
        },
        // onThought
        (thought: any) => {
          setConversations(prev =>
            prev.map(conv =>
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
          )
        },
        // onComplete
        (response: ChatResponse) => {
          setConversations(prev =>
            prev.map(conv => {
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
                
                // 如果是新对话的第一条消息，使用用户消息的前30个字符作为标题
                if (conv.messages.length <= 2 && conv.title === '新对话') {
                  const title = content.length > 30 
                    ? content.substring(0, 30) + '...' 
                    : content
                  return { ...updatedConv, title }
                }
                
                return updatedConv
              }
              return conv
            })
          )
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
    deleteConversation,
    renameConversation,
    sendMessage
  }
}