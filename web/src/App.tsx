import { useState, useEffect } from 'react'
import { Sidebar } from './components/layout/Sidebar'
import { Header } from './components/layout/Header'
import { SettingsPage } from './components/settings/SettingsPage'
import { SkillsPage } from './components/skills/SkillsPage'
import { McpPage } from './components/mcp/McpPage'
import { MultiChat } from './components/conversation/MultiChat'
import { ThemeProvider } from './contexts/ThemeContext'
import { useApi } from './hooks/useApi'
import { useConversations } from './hooks/useConversations'
import { Model } from './types'
import { Brain } from 'lucide-react'

function App() {
  const [models, setModels] = useState<Model[]>([])
  const [selectedModel, setSelectedModel] = useState('')
  const [currentView, setCurrentView] = useState<'chat' | 'skills' | 'settings' | 'mcp'>('chat')
  const [isSidebarOpen, setIsSidebarOpen] = useState(false)
  const [isInitialized, setIsInitialized] = useState(false)
  
  const { getConfig, updateConfig } = useApi()
  const {
    conversations,
    activeConversationId,
    activeConversation,
    isLoading,
    createConversation,
    selectConversation,
    deleteConversation,
    renameConversation,
    sendMessage
  } = useConversations()

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

  const handleChatClick = () => {
    setCurrentView('chat')
    // 如果当前在聊天视图，创建新的对话
    if (currentView === 'chat') {
      createConversation()
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
        onViewChange={(view) => {
          if (view === 'chat') {
            handleChatClick()
          } else {
            setCurrentView(view)
          }
        }}
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

        {/* MultiChat Component */}
        <MultiChat
          conversations={conversations}
          activeConversationId={activeConversationId}
          activeConversation={activeConversation}
          isLoading={isLoading}
          onConversationCreate={createConversation}
          onConversationSelect={selectConversation}
          onConversationDelete={deleteConversation}
          onConversationRename={renameConversation}
          onSendMessage={sendMessage}
        />
      </div>
    </div>
  )
}

export default App