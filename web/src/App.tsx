import { useState, useEffect, useRef } from 'react'
import { ChatMessage } from './components/ChatMessage'
import { ChatInput } from './components/ChatInput'
import { TypingIndicator } from './components/TypingIndicator'
import { Sidebar } from './components/layout/Sidebar'
import { Header } from './components/layout/Header'
import { SettingsPage } from './components/settings/SettingsPage'
import { SkillsPage } from './components/skills/SkillsPage'
import { McpPage } from './components/mcp/McpPage'
import { ScrollArea } from './components/ui/scroll-area'
import { ThemeProvider } from './contexts/ThemeContext'
import { useApi } from './hooks/useApi'
import { Message, Model, ChatResponse } from './types'
import { Brain } from 'lucide-react'

function App() {
  const [messages, setMessages] = useState<Message[]>([])
  const [models, setModels] = useState<Model[]>([])
  const [selectedModel, setSelectedModel] = useState('')
  const [currentView, setCurrentView] = useState<'chat' | 'skills' | 'settings' | 'mcp'>('chat')
  const [isSidebarOpen, setIsSidebarOpen] = useState(false)
  const [isInitialized, setIsInitialized] = useState(false)
  
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const { sendStreamingChatMessage, getConfig, updateConfig, isLoading } = useApi()

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
            const defaultModelFromConfig = config.agents?.defaults?.model
            // 尝试匹配 model 字段，如果找不到则使用 model_name
            const matchingModel = config.model_list.find(m => 
              m.model === defaultModelFromConfig || m.model_name === defaultModelFromConfig
            )
            setSelectedModel(matchingModel ? matchingModel.model_name : config.model_list[0].model_name)
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

    // 创建一个空的助手消息，用于实时更新思考过程和最终内容
    const assistantMessageId = (Date.now() + 1).toString()
    const assistantMessage: Message = {
      id: assistantMessageId,
      content: '',
      role: 'assistant',
      timestamp: new Date(),
      model: selectedModel,
      thoughts: []
    }
    
    setMessages(prev => [...prev, assistantMessage])

    try {
      sendStreamingChatMessage(
        {
          message: content,
          model: selectedModel,
          stream: true
        },
        // onThought: 接收到思考过程时的回调
        (thought: any) => {
          setMessages(prev => 
            prev.map(msg => 
              msg.id === assistantMessageId 
                ? { ...msg, thoughts: [...(msg.thoughts || []), thought] }
                : msg
            )
          )
        },
        // onComplete: 接收到最终回复时的回调
        (response: ChatResponse) => {
          console.log('App.tsx: Received completion:', response)
          setMessages(prev => 
            prev.map(msg => 
              msg.id === assistantMessageId 
                ? { ...msg, content: response.message, timestamp: response.timestamp }
                : msg
            )
          )
        },
        // onError: 出错时的回调
        (error: string) => {
          setMessages(prev => 
            prev.map(msg => 
              msg.id === assistantMessageId 
                ? { ...msg, content: `错误: ${error}` }
                : msg
            )
          )
        }
      )
    } catch (error) {
      setMessages(prev => 
        prev.map(msg => 
          msg.id === assistantMessageId 
            ? { ...msg, content: `网络错误: ${error instanceof Error ? error.message : '未知错误'}` }
            : msg
        )
      )
    }
  }

  const handleModelsChange = async (newModels: Model[]) => {
    console.log('handleModelsChange called with:', newModels.map(m => ({
      name: m.model_name,
      hasApiKey: !!m.api_key,
      apiKeyLength: m.api_key?.length || 0
    })))
    
    setModels(newModels)
    
    if (newModels.length === 0) {
      setSelectedModel('')
      return
    }

    // 更新配置，保持当前的默认模型
    try {
      const configUpdate = {
        model_list: newModels,
        agents: {
          defaults: {
            model: selectedModel // 使用当前选中的模型作为默认模型
          }
        }
      }
      console.log('Updating config with:', configUpdate)
      
      await updateConfig(configUpdate)
      console.log('Config updated successfully')
    } catch (error) {
      console.error('Failed to update models:', error)
    }

    // 如果当前选择的模型不在新列表中，选择第一个
    if (!newModels.find(m => m.model_name === selectedModel)) {
      setSelectedModel(newModels[0].model_name)
    }
  }

  const handleSelectedModelChange = async (newDefaultModel: string) => {
    setSelectedModel(newDefaultModel)
    
    // 更新配置文件中的默认模型
    try {
      await updateConfig({
        agents: {
          defaults: {
            model: newDefaultModel
          }
        }
      })
    } catch (error) {
      console.error('Failed to update default model:', error)
    }
  }

  if (!isInitialized) {
    return (
      <ThemeProvider>
        <div className="flex items-center justify-center h-screen bg-gradient-to-br from-background via-muted/10 to-background">
          <div className="flex flex-col items-center gap-4 p-8">
            <Brain className="w-12 h-12 text-primary animate-pulse" />
            <div className="text-center">
              <p className="text-lg font-medium">正在初始化...</p>
              <p className="text-sm text-muted-foreground mt-1">PicoClaw AI Assistant</p>
            </div>
          </div>
        </div>
      </ThemeProvider>
    )
  }

  if (currentView === 'skills') {
    return (
      <ThemeProvider>
        <SkillsPage onBack={() => setCurrentView('chat')} />
      </ThemeProvider>
    )
  }

  if (currentView === 'mcp') {
    return (
      <ThemeProvider>
        <McpPage onBack={() => setCurrentView('chat')} />
      </ThemeProvider>
    )
  }

  if (currentView === 'settings') {
    return (
      <ThemeProvider>
        <SettingsPage
          onBack={() => setCurrentView('chat')}
          models={models}
          onModelsChange={handleModelsChange}
          selectedModel={selectedModel}
          onSelectedModelChange={handleSelectedModelChange}
        />
      </ThemeProvider>
    )
  }

  return (
    <div className="flex h-screen overflow-hidden">
      {/* Sidebar */}
      <Sidebar
        currentView={currentView}
        onViewChange={setCurrentView}
        isSidebarOpen={isSidebarOpen}
        onSidebarToggle={() => setIsSidebarOpen(!isSidebarOpen)}
      />

      {/* Main Content */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Header */}
        <Header
          selectedModel={selectedModel}
          models={models}
          onSidebarToggle={() => setIsSidebarOpen(!isSidebarOpen)}
        />

        {/* Messages */}
        <ScrollArea className="flex-1">
          <div className="chat-messages w-full h-full px-4">
            {messages.length === 0 ? (
              <div className="flex items-center justify-center h-64 text-muted-foreground">
                <div className="text-center">
                  <div className="relative mb-6">
                    <Brain className="w-16 h-16 mx-auto text-primary opacity-20" />
                    <div className="absolute inset-0 flex items-center justify-center">
                      <div className="w-4 h-4 bg-primary rounded-full animate-ping" />
                    </div>
                  </div>
                  <h2 className="text-2xl font-semibold mb-2">开始对话吧！</h2>
                  <p className="text-muted-foreground mb-4">我是 PicoClaw，您的智能 AI 助手</p>
                  <div className="flex flex-wrap justify-center gap-2 text-sm">
                    <span className="px-3 py-1 bg-primary/10 text-primary rounded-full">📝 内容创作</span>
                    <span className="px-3 py-1 bg-primary/10 text-primary rounded-full">💻 编程助手</span>
                    <span className="px-3 py-1 bg-primary/10 text-primary rounded-full">🔍 数据分析</span>
                    <span className="px-3 py-1 bg-primary/10 text-primary rounded-full">🎨 创意设计</span>
                  </div>
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
        <div className="border-t bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 p-4">
          <div className="w-full px-4">
            <ChatInput
              onSendMessage={handleSendMessage}
              isLoading={isLoading}
              disabled={!selectedModel || models.length === 0}
            />
          </div>
        </div>
      </div>
    </div>
  )
}

export default App