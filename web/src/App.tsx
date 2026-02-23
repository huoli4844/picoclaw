import { useState, useEffect, useRef } from 'react'
import { ChatMessage } from './components/ChatMessage'
import { ChatInput } from './components/ChatInput'
import { ModelSelector } from './components/ModelSelector'
import { TypingIndicator } from './components/TypingIndicator'
import { ModelSettings } from './components/settings/ModelSettings'
import { ScrollArea } from './components/ui/scroll-area'
import { useApi } from './hooks/useApi'
import { Message, Model } from './types'
import { Brain } from 'lucide-react'

function App() {
  const [messages, setMessages] = useState<Message[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [models, setModels] = useState<Model[]>([])
  const [selectedModel, setSelectedModel] = useState('')
  const [isSettingsOpen, setIsSettingsOpen] = useState(false)
  const [isInitialized, setIsInitialized] = useState(false)
  
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const { sendChatMessage, getConfig, updateConfig } = useApi()

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  useEffect(() => {
    const initializeApp = async () => {
      try {
        const configResult = await getConfig()
        if (configResult.success && configResult.data) {
          const config = configResult.data
          setModels(config.model_list || [])
          
          if (config.model_list && config.model_list.length > 0) {
            const defaultModel = config.agents?.defaults?.model || config.model_list[0].model_name
            setSelectedModel(defaultModel)
          }
        }
        setIsInitialized(true)
      } catch (error) {
        console.error('Failed to initialize app:', error)
        // 设置默认模型以便测试
        setModels([
          {
            model_name: 'gpt-4',
            model: 'openai/gpt-4',
            api_key: ''
          }
        ])
        setSelectedModel('gpt-4')
        setIsInitialized(true)
      }
    }

    initializeApp()
  }, [getConfig])

  const handleSendMessage = async (content: string) => {
    const userMessage: Message = {
      id: Date.now().toString(),
      content,
      role: 'user',
      timestamp: new Date(),
      model: selectedModel
    }

    setMessages(prev => [...prev, userMessage])
    setIsLoading(true)

    try {
      const response = await sendChatMessage({
        message: content,
        model: selectedModel,
        stream: false
      })

      if (response.success && response.data) {
        const assistantMessage: Message = {
          id: (Date.now() + 1).toString(),
          content: response.data.message,
          role: 'assistant',
          timestamp: new Date(response.data.timestamp),
          model: response.data.model
        }
        setMessages(prev => [...prev, assistantMessage])
      } else {
        const errorMessage: Message = {
          id: (Date.now() + 1).toString(),
          content: `错误: ${response.error || '未知错误'}`,
          role: 'assistant',
          timestamp: new Date(),
          model: selectedModel
        }
        setMessages(prev => [...prev, errorMessage])
      }
    } catch (error) {
      const errorMessage: Message = {
        id: (Date.now() + 1).toString(),
        content: `网络错误: ${error instanceof Error ? error.message : '未知错误'}`,
        role: 'assistant',
        timestamp: new Date(),
        model: selectedModel
      }
      setMessages(prev => [...prev, errorMessage])
    } finally {
      setIsLoading(false)
    }
  }

  const handleModelsChange = async (newModels: Model[]) => {
    setModels(newModels)
    
    if (newModels.length === 0) {
      setSelectedModel('')
      return
    }

    // 更新配置
    try {
      await updateConfig({
        model_list: newModels,
        agents: {
          defaults: {
            model: newModels[0].model_name
          }
        }
      })
    } catch (error) {
      console.error('Failed to update models:', error)
    }

    // 如果当前选择的模型不在新列表中，选择第一个
    if (!newModels.find(m => m.model_name === selectedModel)) {
      setSelectedModel(newModels[0].model_name)
    }
  }

  if (!isInitialized) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="flex items-center gap-2">
          <Brain className="w-6 h-6 animate-pulse" />
          <span>正在初始化...</span>
        </div>
      </div>
    )
  }

  return (
    <div className="chat-container">
      {/* Header */}
      <header className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="max-w-4xl mx-auto px-4 py-4">
          <div className="flex items-center gap-3">
            <Brain className="w-8 h-8 text-primary" />
            <div>
              <h1 className="text-xl font-semibold">PicoClaw Web</h1>
              <p className="text-sm text-muted-foreground">轻量级 AI 助手</p>
            </div>
          </div>
        </div>
      </header>

      {/* Model Selector */}
      <ModelSelector
        models={models}
        selectedModel={selectedModel}
        onModelChange={setSelectedModel}
        onOpenSettings={() => setIsSettingsOpen(true)}
      />

      {/* Messages */}
      <ScrollArea className="flex-1">
        <div className="chat-messages max-w-4xl mx-auto">
          {messages.length === 0 ? (
            <div className="flex items-center justify-center h-64 text-muted-foreground">
              <div className="text-center">
                <Brain className="w-12 h-12 mx-auto mb-4 opacity-50" />
                <p>开始对话吧！</p>
                <p className="text-sm mt-2">我是 PicoClaw，您的轻量级 AI 助手</p>
              </div>
            </div>
          ) : (
            <>
              {messages.map((message) => (
                <ChatMessage key={message.id} message={message} />
              ))}
              {isLoading && <TypingIndicator />}
            </>
          )}
          <div ref={messagesEndRef} />
        </div>
      </ScrollArea>

      {/* Input */}
      <ChatInput
        onSendMessage={handleSendMessage}
        isLoading={isLoading}
        disabled={!selectedModel || models.length === 0}
      />

      {/* Settings Dialog */}
      <ModelSettings
        isOpen={isSettingsOpen}
        onClose={() => setIsSettingsOpen(false)}
        models={models}
        onModelsChange={handleModelsChange}
      />
    </div>
  )
}

export default App