import { useState } from 'react'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '../ui/dialog'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Label } from '../ui/label'
import { Model } from '@/types'
import { Plus, Trash2, Edit, Star } from 'lucide-react'

interface ModelSettingsProps {
  isOpen: boolean
  onClose: () => void
  models: Model[]
  onModelsChange: (models: Model[]) => void
  selectedModel: string
  onSelectedModelChange: (model: string) => void
}

export function ModelSettings({ isOpen, onClose, models, onModelsChange, selectedModel, onSelectedModelChange }: ModelSettingsProps) {
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

  const handleSave = () => {
    console.log('Saving models with API keys:', editingModels.map(m => ({
      name: m.model_name,
      hasApiKey: !!m.api_key,
      apiKeyLength: m.api_key?.length || 0
    })))
    onModelsChange(editingModels)
    onClose()
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

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>模型设置</DialogTitle>
        </DialogHeader>
        
        <div className="space-y-6">
          {/* 现有模型列表 */}
          <div className="space-y-3">
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
                      </div>
                    </>
                  )}
                </div>
              ))
            )}
          </div>

          {/* 添加新模型 */}
          <div className="border-t pt-4">
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
          <div className="border-t pt-4">
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
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            取消
          </Button>
          <Button onClick={handleSave}>
            保存
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}