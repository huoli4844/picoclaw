
import { Button } from './ui/button'
import { Settings, Package } from 'lucide-react'
import { Model } from '@/types'

interface ModelSelectorProps {
  models: Model[]
  selectedModel: string
  onOpenSettings: () => void
  onOpenSkills: () => void
}

export function ModelSelector({ 
  models, 
  selectedModel, 
  onOpenSettings,
  onOpenSkills
}: ModelSelectorProps) {
  const currentModel = models.find(m => m.model_name === selectedModel)
  
  return (
    <div className="border-b bg-background p-4">
      <div className="flex items-center justify-between max-w-4xl mx-auto">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium">当前模型:</span>
          <div className="flex items-center gap-2 px-3 py-1 bg-muted rounded-md">
            <div className="flex flex-col">
              <div className="font-medium">
                {currentModel?.model_name || selectedModel || '未配置'}
              </div>
              {currentModel?.model && (
                <div className="text-xs text-muted-foreground">
                  {currentModel.model}
                </div>
              )}
            </div>
          </div>
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