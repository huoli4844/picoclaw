import { useState, useEffect } from 'react'
import { Settings as SettingsIcon, Palette, Database, Bot, Check, Plus, Trash2, Edit, Star, RefreshCw } from 'lucide-react'
import { useTheme } from '@/contexts/ThemeContext'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Model } from '@/types'
import { useFileBrowser } from '@/hooks/useFileBrowser'
import { FileBrowser } from '@/components/file-browser/FileBrowser'
import { FileViewer } from '@/components/file-browser/FileViewer'

interface SettingsPageProps {
  onBack: () => void
  models: any[]
  onModelsChange: (models: any[]) => void
  selectedModel: string
  onSelectedModelChange: (model: string) => void
}

export function SettingsPage({ onBack: _onBack, models, onModelsChange, selectedModel, onSelectedModelChange }: SettingsPageProps) {
  const { theme, setTheme } = useTheme()
  const [activeSection, setActiveSection] = useState('general')
  
  // 模型配置状态
  const [editingModels, setEditingModels] = useState<Model[]>([...models])
  const [newModel, setNewModel] = useState<Partial<Model>>({
    model_name: '',
    model: '',
    api_key: '',
    api_base: '',
  })
  const [editingIndex, setEditingIndex] = useState<number | null>(null)
  const [editingModel, setEditingModel] = useState<Partial<Model>>({
    model_name: '',
    model: '',
    api_key: '',
    api_base: '',
  })

  // 文件浏览器状态
  const {
    files,
    currentPath,
    isLoading,
    error,
    listFiles,
    getFileContent,
    navigateToDirectory,
    navigateUp,
    deleteFileOrDirectory
  } = useFileBrowser()
  
  const [fileContent, setFileContent] = useState<any>(null)
  const [isFileViewerOpen, setIsFileViewerOpen] = useState(false)

  // 初始化时加载文件列表
  useEffect(() => {
    if (activeSection === 'data') {
      listFiles()
    }
  }, [activeSection, listFiles])

  // 处理文件内容查看
  const handleFileContent = async (path: string) => {
    const content = await getFileContent(path)
    if (content) {
      setFileContent(content)
      setIsFileViewerOpen(true)
    }
  }

  const settingsSections = [
    { id: 'models', label: '模型配置', icon: Bot },
    { id: 'data', label: '数据', icon: Database },
    { id: 'appearance', label: '外观', icon: Palette },
    { id: 'general', label: '关于', icon: SettingsIcon },
  ]

  // 模型配置处理函数
  const handleAddModel = () => {
    if (newModel.model_name && newModel.model) {
      setEditingModels([...editingModels, newModel as Model])
      setNewModel({ model_name: '', model: '', api_key: '', api_base: '' })
    }
  }

  const handleRemoveModel = (index: number) => {
    setEditingModels(editingModels.filter((_, i) => i !== index))
  }

  const handleStartEdit = (index: number) => {
    const model = editingModels[index]
    setEditingIndex(index)
    setEditingModel({
      model_name: model.model_name,
      model: model.model,
      api_key: model.api_key,
      api_base: model.api_base || '',
    })
  }

  const handleSaveEdit = () => {
    if (editingIndex !== null && editingModel.model_name && editingModel.model) {
      const updatedModels = [...editingModels]
      updatedModels[editingIndex] = editingModel as Model
      setEditingModels(updatedModels)
      setEditingIndex(null)
      setEditingModel({ model_name: '', model: '', api_key: '', api_base: '' })
    }
  }

  const handleCancelEdit = () => {
    setEditingIndex(null)
    setEditingModel({ model_name: '', model: '', api_key: '', api_base: '' })
  }

  const handleSaveModels = () => {
    console.log('Saving models with API keys:', editingModels.map(m => ({
      name: m.model_name,
      hasApiKey: !!m.api_key,
      apiKeyLength: m.api_key?.length || 0
    })))
    onModelsChange(editingModels)
  }

  const getProviderFromModel = (model: string): string => {
    if (model.includes('openai/')) return 'openai'
    if (model.includes('anthropic/')) return 'anthropic'
    if (model.includes('zhipu/')) return 'zhipu'
    if (model.includes('deepseek/')) return 'deepseek'
    if (model.includes('gemini/')) return 'gemini'
    if (model.includes('ollama/')) return 'ollama'
    return 'other'
  }

  const renderContent = () => {
    switch (activeSection) {
      case 'models':
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
            
            {/* 模型配置 */}
            <div>
              <h3 className="text-lg font-medium mb-4">模型配置</h3>
              
              {/* 现有模型列表 */}
              <div className="space-y-3 mb-6">
                <Label>已配置的模型</Label>
                {editingModels.length === 0 ? (
                  <p className="text-sm text-muted-foreground">暂无配置的模型</p>
                ) : (
                  editingModels.map((model, index) => (
                    <div key={index} className="border rounded-lg p-3 space-y-3">
                      {editingIndex === index ? (
                        // 编辑模式
                        <div className="space-y-3">
                          <div className="grid grid-cols-2 gap-3">
                            <div>
                              <Label htmlFor={`edit-model-name-${index}`} className="text-sm">模型名称</Label>
                              <Input
                                id={`edit-model-name-${index}`}
                                value={editingModel.model_name}
                                onChange={(e) => setEditingModel({ ...editingModel, model_name: e.target.value })}
                                placeholder="例如: gpt-4"
                              />
                            </div>
                            <div>
                              <Label htmlFor={`edit-model-id-${index}`} className="text-sm">模型ID</Label>
                              <Input
                                id={`edit-model-id-${index}`}
                                value={editingModel.model}
                                onChange={(e) => setEditingModel({ ...editingModel, model: e.target.value })}
                                placeholder="例如: openai/gpt-4"
                              />
                            </div>
                            <div>
                              <Label htmlFor={`edit-api-key-${index}`} className="text-sm">API Key</Label>
                              <Input
                                id={`edit-api-key-${index}`}
                                type="password"
                                value={editingModel.api_key}
                                onChange={(e) => setEditingModel({ ...editingModel, api_key: e.target.value })}
                                placeholder="API Key (可选)"
                              />
                            </div>
                            <div>
                              <Label htmlFor={`edit-api-base-${index}`} className="text-sm">API Base</Label>
                              <Input
                                id={`edit-api-base-${index}`}
                                value={editingModel.api_base}
                                onChange={(e) => setEditingModel({ ...editingModel, api_base: e.target.value })}
                                placeholder="API Base URL (可选)"
                              />
                            </div>
                          </div>
                          <div className="flex gap-2">
                            <Button size="sm" onClick={handleSaveEdit}>
                              保存
                            </Button>
                            <Button size="sm" variant="outline" onClick={handleCancelEdit}>
                              取消
                            </Button>
                          </div>
                        </div>
                      ) : (
                        // 显示模式
                        <>
                          <div className="flex items-center justify-between">
                            <div className="flex-1 cursor-pointer hover:text-primary transition-colors" onClick={() => handleStartEdit(index)}>
                              <div className="flex items-center gap-2">
                                <div className="font-medium">{model.model_name}</div>
                                {selectedModel === model.model_name && (
                                  <div className="px-2 py-0.5 bg-primary/10 text-primary text-xs rounded-full flex items-center gap-1">
                                    <Star className="w-3 h-3 fill-current" />
                                    默认
                                  </div>
                                )}
                              </div>
                              <div className="text-sm text-muted-foreground">{model.model}</div>
                            </div>
                            <div className="flex gap-2">
                              <Button
                                variant={selectedModel === model.model_name ? "default" : "outline"}
                                size="sm"
                                onClick={() => onSelectedModelChange(model.model_name)}
                                title={selectedModel === model.model_name ? "当前默认模型" : "设为默认模型"}
                              >
                                <Star className={`w-4 h-4 ${selectedModel === model.model_name ? 'fill-current' : ''}`} />
                              </Button>
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => handleStartEdit(index)}
                                title="编辑模型"
                              >
                                <Edit className="w-4 h-4" />
                              </Button>
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => handleRemoveModel(index)}
                                title="删除模型"
                              >
                                <Trash2 className="w-4 h-4" />
                              </Button>
                            </div>
                          </div>
                          <div className="text-xs text-muted-foreground">
                            提供商: {getProviderFromModel(model.model)}
                            {model.api_base && ` | API Base: ${model.api_base}`}
                            {model.api_key && ` | API Key: 已配置`}
                          </div>
                        </>
                      )}
                    </div>
                  ))
                )}
              </div>

              {/* 添加新模型 */}
              <div className="border-t pt-4 mb-6">
                <Label>添加新模型</Label>
                <div className="grid grid-cols-2 gap-3 mt-2">
                  <div>
                    <Label htmlFor="model-name" className="text-sm">模型名称</Label>
                    <Input
                      id="model-name"
                      value={newModel.model_name}
                      onChange={(e) => setNewModel({ ...newModel, model_name: e.target.value })}
                      placeholder="例如: gpt-4"
                    />
                  </div>
                  <div>
                    <Label htmlFor="model-id" className="text-sm">模型ID</Label>
                    <Input
                      id="model-id"
                      value={newModel.model}
                      onChange={(e) => setNewModel({ ...newModel, model: e.target.value })}
                      placeholder="例如: openai/gpt-4"
                    />
                  </div>
                  <div>
                    <Label htmlFor="api-key" className="text-sm">API Key</Label>
                    <Input
                      id="api-key"
                      type="password"
                      value={newModel.api_key}
                      onChange={(e) => setNewModel({ ...newModel, api_key: e.target.value })}
                      placeholder="API Key (可选)"
                    />
                  </div>
                  <div>
                    <Label htmlFor="api-base" className="text-sm">API Base</Label>
                    <Input
                      id="api-base"
                      value={newModel.api_base}
                      onChange={(e) => setNewModel({ ...newModel, api_base: e.target.value })}
                      placeholder="API Base URL (可选)"
                    />
                  </div>
                </div>
                <Button
                  onClick={handleAddModel}
                  disabled={!newModel.model_name || !newModel.model}
                  className="mt-3"
                  variant="outline"
                >
                  <Plus className="w-4 h-4 mr-2" />
                  添加模型
                </Button>
              </div>

              {/* 常用模型模板 */}
              <div className="border-t pt-4 mb-6">
                <Label>常用模型模板</Label>
                <div className="grid grid-cols-1 gap-2 mt-2">
                  {[
                    { name: 'GPT-4', model: 'openai/gpt-4', base: 'https://api.openai.com/v1' },
                    { name: 'Claude-3.5-Sonnet', model: 'anthropic/claude-3-5-sonnet-20241022', base: 'https://api.anthropic.com/v1' },
                    { name: 'GLM-4', model: 'zhipu/glm-4', base: 'https://open.bigmodel.cn/api/paas/v4' },
                    { name: 'DeepSeek', model: 'deepseek/deepseek-chat', base: 'https://api.deepseek.com/v1' },
                    { name: 'Ollama Local', model: 'ollama/llama3', base: 'http://localhost:11434/v1' },
                  ].map((template, index) => (
                    <Button
                      key={index}
                      variant="ghost"
                      className="justify-start h-auto p-2"
                      onClick={() => setNewModel({
                        model_name: template.name,
                        model: template.model,
                        api_base: template.base,
                        api_key: ''
                      })}
                    >
                      <div className="text-left">
                        <div className="font-medium">{template.name}</div>
                        <div className="text-xs text-muted-foreground">{template.model}</div>
                      </div>
                    </Button>
                  ))}
                </div>
              </div>

              {/* 保存按钮 */}
              <div className="flex gap-2">
                <Button onClick={handleSaveModels}>
                  保存模型配置
                </Button>
              </div>
            </div>
          </div>
        )

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

      case 'data':
        return (
          <div className="space-y-6">
            <div>
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-medium">文件浏览器</h3>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => listFiles()}
                  disabled={isLoading}
                >
                  <RefreshCw className={`w-4 h-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
                  刷新
                </Button>
              </div>
              
              <div className="bg-muted/30 border rounded-lg h-[500px]">
                <FileBrowser
                  files={files}
                  currentPath={currentPath}
                  isLoading={isLoading}
                  error={error}
                  onFileClick={(file) => {
                    if (!file.isDir) {
                      handleFileContent(file.path)
                    }
                  }}
                  onDirectoryClick={navigateToDirectory}
                  onNavigateUp={navigateUp}
                  onFileContent={handleFileContent}
                  onDeleteFile={deleteFileOrDirectory}
                />
              </div>
              
              <div className="text-xs text-muted-foreground mt-2">
                <p>• 显示路径: {currentPath || '.picoclaw'}</p>
                <p>• 文件数量: {files.length} 个</p>
                <p>• 支持查看文本文件、代码文件等，最大文件大小 10MB</p>
              </div>
            </div>
          </div>
        )
      
      case 'general':
      default:
        return (
          <div className="space-y-6">
            <div>
              <h3 className="text-lg font-medium mb-4">系统信息</h3>
              <div className="bg-muted/30 border rounded-lg p-4">
                <div className="text-sm space-y-2">
                  <p>• PicoClaw AI Assistant</p>
                  <p>• 版本: 1.0.0</p>
                  <p>• 当前模型数量: {models.length}</p>
                  <p>• 当前默认模型: {selectedModel || '未设置'}</p>
                </div>
              </div>
            </div>
          </div>
        )
    }
  }

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="border-b p-4">
        <div className="w-[80%] mx-auto">
          <div className="flex items-center gap-3">
            <SettingsIcon className="w-6 h-6 text-primary" />
            <div>
              <h2 className="text-lg font-semibold">设置</h2>
              <p className="text-sm text-muted-foreground">应用程序配置和偏好设置</p>
            </div>
          </div>
        </div>
      </div>

      {/* Settings Content */}
      <div className="flex-1 overflow-y-auto">
        <div className="w-[80%] mx-auto p-4">
          {/* Settings Navigation */}
          <div className="mb-6">
            <div className="flex flex-wrap gap-2">
              {settingsSections.map((section) => {
                const Icon = section.icon
                const isActive = activeSection === section.id
                
                return (
                  <button
                    key={section.id}
                    onClick={() => setActiveSection(section.id)}
                    className={`
                      flex items-center gap-2 px-3 py-2 rounded-md text-sm font-medium transition-colors
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
            </div>
          </div>

          {/* Settings Content */}
          <div className="space-y-6">
            {renderContent()}
          </div>
        </div>
      </div>
      
      {/* 文件查看器 */}
      <FileViewer
        fileContent={fileContent}
        isOpen={isFileViewerOpen}
        onClose={() => {
          setIsFileViewerOpen(false)
          setFileContent(null)
        }}
      />
    </div>
  )
}