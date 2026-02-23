
import { Model } from '@/types'

interface ModelSelectorProps {
  models: Model[]
  selectedModel: string
}

export function ModelSelector({ 
  models, 
  selectedModel
}: ModelSelectorProps) {
  const currentModel = models.find(m => m.model_name === selectedModel)
  
  return (
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
  )
}