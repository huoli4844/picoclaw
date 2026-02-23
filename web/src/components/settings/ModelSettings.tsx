import { useState } from 'react'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '../ui/dialog'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Label } from '../ui/label'
import { Model } from '@/types'
import { Plus, Trash2 } from 'lucide-react'

interface ModelSettingsProps {
  isOpen: boolean
  onClose: () => void
  models: Model[]
  onModelsChange: (models: Model[]) => void
}

export function ModelSettings({ isOpen, onClose, models, onModelsChange }: ModelSettingsProps) {
  const [editingModels, setEditingModels] = useState<Model[]>([...models])
  const [newModel, setNewModel] = useState<Partial<Model>>({
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

  const handleSave = () => {
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
                <div key={index} className="border rounded-lg p-3 space-y-2">
                  <div className="flex items-center justify-between">
                    <div className="flex-1">
                      <div className="font-medium">{model.model_name}</div>
                      <div className="text-sm text-muted-foreground">{model.model}</div>
                    </div>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleRemoveModel(index)}
                    >
                      <Trash2 className="w-4 h-4" />
                    </Button>
                  </div>
                  <div className="text-xs text-muted-foreground">
                    提供商: {getProviderFromModel(model.model)}
                    {model.api_base && ` | API Base: ${model.api_base}`}
                  </div>
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