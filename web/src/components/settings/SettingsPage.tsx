import { useState } from 'react'
import { ArrowLeft, Settings as SettingsIcon, Palette, Bell, Database, Shield, Bot, Check } from 'lucide-react'
import { useTheme } from '@/contexts/ThemeContext'
import { ModelSettings } from './ModelSettings'

interface SettingsPageProps {
  onBack: () => void
  models: any[]
  onModelsChange: (models: any[]) => void
  selectedModel: string
  onSelectedModelChange: (model: string) => void
}

export function SettingsPage({ onBack, models, onModelsChange, selectedModel, onSelectedModelChange }: SettingsPageProps) {
  const { theme, setTheme } = useTheme()
  const [activeSection, setActiveSection] = useState('general')

  const settingsSections = [
    { id: 'general', label: '通用', icon: SettingsIcon },
    { id: 'appearance', label: '外观', icon: Palette },
    { id: 'notifications', label: '通知', icon: Bell },
    { id: 'data', label: '数据', icon: Database },
    { id: 'privacy', label: '隐私', icon: Shield },
  ]

  const renderContent = () => {
    switch (activeSection) {
      case 'appearance':
        return (
          <div className="space-y-6">
            <div>
              <h3 className="text-lg font-medium mb-4">主题设置</h3>
              <div className="space-y-3">
                {[
                  { value: 'light', label: '亮色主题', description: '使用明亮界面，适合日间使用' },
                  { value: 'dark', label: '暗色主题', description: '使用深色界面，适合夜间使用' },
                  { value: 'system', label: '跟随系统', description: '自动跟随系统主题设置' }
                ].map((option) => (
                  <label
                    key={option.value}
                    className="flex items-center p-4 border rounded-lg cursor-pointer hover:bg-accent/50 transition-colors"
                  >
                    <input
                      type="radio"
                      name="theme"
                      value={option.value}
                      checked={theme === option.value}
                      onChange={(e) => setTheme(e.target.value as any)}
                      className="mr-3"
                    />
                    <div>
                      <div className="font-medium">{option.label}</div>
                      <div className="text-sm text-muted-foreground">{option.description}</div>
                    </div>
                  </label>
                ))}
              </div>
            </div>
          </div>
        )
      
      case 'general':
      default:
        return (
          <div className="space-y-6">
            {/* 当前使用模型显示 */}
            <div>
              <h3 className="text-lg font-medium mb-4">当前模型</h3>
              <div className="bg-muted/50 border rounded-lg p-4">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 bg-primary/10 rounded-full flex items-center justify-center">
                    <Bot className="w-5 h-5 text-primary" />
                  </div>
                  <div className="flex-1">
                    <div className="font-medium text-lg">
                      {selectedModel || '未选择模型'}
                    </div>
                    <div className="text-sm text-muted-foreground">
                      {selectedModel ? '当前正在使用此模型进行对话' : '请选择一个模型开始对话'}
                    </div>
                  </div>
                  {selectedModel && (
                    <div className="w-6 h-6 bg-green-100 dark:bg-green-900/30 rounded-full flex items-center justify-center">
                      <Check className="w-4 h-4 text-green-600 dark:text-green-400" />
                    </div>
                  )}
                </div>
                
                {/* 模型统计信息 */}
                {selectedModel && (
                  <div className="mt-4 grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div className="bg-background rounded p-3">
                      <div className="text-sm text-muted-foreground">可用模型</div>
                      <div className="text-xl font-semibold">{models.length}</div>
                    </div>
                    <div className="bg-background rounded p-3">
                      <div className="text-sm text-muted-foreground">模型类型</div>
                      <div className="text-sm font-medium">
                        {models.find(m => m.model_name === selectedModel)?.model?.split('/')[0] || '未知'}
                      </div>
                    </div>
                    <div className="bg-background rounded p-3">
                      <div className="text-sm text-muted-foreground">状态</div>
                      <div className="text-sm font-medium text-green-600 dark:text-green-400">
                        活跃
                      </div>
                    </div>
                  </div>
                )}
              </div>
            </div>
            
            <div>
              <h3 className="text-lg font-medium mb-4">模型配置</h3>
              <ModelSettings
                isOpen={true}
                onClose={onBack}
                models={models}
                onModelsChange={onModelsChange}
                selectedModel={selectedModel}
                onSelectedModelChange={onSelectedModelChange}
              />
            </div>
          </div>
        )
    }
  }

  return (
    <div className="flex h-full">
      {/* 侧边栏 */}
      <div className="w-64 border-r bg-card">
        <div className="p-4">
          <button
            onClick={onBack}
            className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground mb-4"
          >
            <ArrowLeft className="w-4 h-4" />
            返回
          </button>
          
          <h2 className="text-lg font-semibold mb-4">设置</h2>
          
          <nav className="space-y-2">
            {settingsSections.map((section) => {
              const Icon = section.icon
              const isActive = activeSection === section.id
              
              return (
                <button
                  key={section.id}
                  onClick={() => setActiveSection(section.id)}
                  className={`
                    w-full flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors
                    ${isActive 
                      ? 'bg-primary text-primary-foreground' 
                      : 'text-muted-foreground hover:text-foreground hover:bg-accent'
                    }
                  `}
                >
                  <Icon className="w-4 h-4" />
                  <span>{section.label}</span>
                </button>
              )
            })}
          </nav>
        </div>
      </div>

      {/* 主内容区 */}
      <div className="flex-1 overflow-y-auto">
        <div className="max-w-2xl mx-auto p-6">
          {renderContent()}
        </div>
      </div>
    </div>
  )
}