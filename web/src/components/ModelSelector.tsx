
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select'
import { Button } from './ui/button'
import { Settings, Package } from 'lucide-react'
import { Model } from '@/types'

interface ModelSelectorProps {
  models: Model[]
  selectedModel: string
  onModelChange: (model: string) => void
  onOpenSettings: () => void
  onOpenSkills: () => void
}

export function ModelSelector({ 
  models, 
  selectedModel, 
  onModelChange, 
  onOpenSettings,
  onOpenSkills
}: ModelSelectorProps) {
  return (
    <div className="border-b bg-background p-4">
      <div className="flex items-center justify-between max-w-4xl mx-auto">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium">模型:</span>
          <Select value={selectedModel} onValueChange={onModelChange}>
            <SelectTrigger className="w-[200px]">
              <SelectValue placeholder="选择模型" />
            </SelectTrigger>
            <SelectContent>
              {models.map((model) => (
                <SelectItem key={model.model_name} value={model.model_name}>
                  {model.model_name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={onOpenSkills}>
            <Package className="w-4 h-4 mr-2" />
            技能
          </Button>
          <Button variant="outline" size="sm" onClick={onOpenSettings}>
            <Settings className="w-4 h-4 mr-2" />
            设置
          </Button>
        </div>
      </div>
    </div>
  )
}