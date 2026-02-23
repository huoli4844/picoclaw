
import { Menu } from 'lucide-react'
import { ModelSelector } from '@/components/ModelSelector'
import { ThemeToggle } from '@/components/ui/theme-toggle'

interface HeaderProps {
  selectedModel: string
  models: any[]
  onModelChange: (model: string) => void
  onOpenSettings: () => void
  onSidebarToggle: () => void
}

export function Header({ selectedModel, models, onModelChange, onOpenSettings, onSidebarToggle }: HeaderProps) {
  return (
    <header className="sticky top-0 z-30 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="flex h-16 items-center px-4 lg:px-6">
        <button
          onClick={onSidebarToggle}
          className="lg:hidden mr-4 p-2 rounded-md hover:bg-accent"
        >
          <Menu className="w-5 h-5" />
          <span className="sr-only">打开侧边栏</span>
        </button>

        <div className="flex-1 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-semibold hidden lg:block">PicoClaw Web</h1>
            <span className="text-sm text-muted-foreground lg:hidden">PicoClaw</span>
          </div>

          <div className="flex items-center gap-3">
            <ThemeToggle />
            <ModelSelector
              models={models}
              selectedModel={selectedModel}
              onModelChange={onModelChange}
              onOpenSettings={onOpenSettings}
              onOpenSkills={() => {}} // 这将在主App中处理
            />
          </div>
        </div>
      </div>
    </header>
  )
}