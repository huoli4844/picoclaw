
import { Brain, Settings, Package, MessageSquare, Github } from 'lucide-react'
import { ThemeToggle } from '@/components/ui/theme-toggle'

interface SidebarProps {
  currentView: 'chat' | 'skills' | 'settings'
  onViewChange: (view: 'chat' | 'skills' | 'settings') => void
  isSidebarOpen: boolean
  onSidebarToggle: () => void
}

export function Sidebar({ currentView, onViewChange, isSidebarOpen, onSidebarToggle }: SidebarProps) {

  const menuItems = [
    { id: 'chat', label: '对话', icon: MessageSquare },
    { id: 'skills', label: '技能', icon: Package },
    { id: 'settings', label: '设置', icon: Settings },
  ]

  return (
    <>
      {/* 移动端遮罩 */}
      {isSidebarOpen && (
        <div
          className="fixed inset-0 bg-black/50 z-40 lg:hidden"
          onClick={onSidebarToggle}
        />
      )}
      
      {/* 侧边栏 */}
      <aside className={`
        fixed top-0 left-0 z-50 h-screen w-64 transform transition-transform duration-300 ease-in-out
        bg-card border-r border-border
        lg:relative lg:translate-x-0 lg:z-0
        ${isSidebarOpen ? 'translate-x-0' : '-translate-x-full'}
      `}>
        <div className="flex flex-col h-full">
          {/* Header */}
          <div className="flex items-center justify-between p-6 border-b border-border">
            <div className="flex items-center gap-3">
              <div className="relative">
                <Brain className="w-8 h-8 text-primary" />
                <div className="absolute -top-1 -right-1 w-3 h-3 bg-green-500 rounded-full animate-pulse" />
              </div>
              <div>
                <h1 className="text-lg font-bold">PicoClaw</h1>
                <p className="text-xs text-muted-foreground">AI Assistant</p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <ThemeToggle />
              <button
                onClick={onSidebarToggle}
                className="lg:hidden p-2 rounded-md hover:bg-accent"
              >
                <span className="sr-only">关闭侧边栏</span>
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>

          {/* Navigation */}
          <nav className="flex-1 p-4">
            <ul className="space-y-2">
              {menuItems.map((item) => {
                const Icon = item.icon
                const isActive = currentView === item.id
                
                return (
                  <li key={item.id}>
                    <button
                      onClick={() => {
                        onViewChange(item.id as any)
                        onSidebarToggle() // 移动端选择后关闭侧边栏
                      }}
                      className={`
                        w-full flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors
                        ${isActive 
                          ? 'bg-primary text-primary-foreground' 
                          : 'text-muted-foreground hover:text-foreground hover:bg-accent'
                        }
                      `}
                    >
                      <Icon className="w-4 h-4" />
                      <span>{item.label}</span>
                    </button>
                  </li>
                )
              })}
            </ul>
          </nav>

          {/* Footer */}
          <div className="p-4 border-t border-border">
            <div className="flex items-center justify-between">
              <a
                href="https://github.com/sipeed/picoclaw"
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors"
              >
                <Github className="w-4 h-4" />
                <span className="hidden sm:inline">GitHub</span>
              </a>
              <ThemeToggle />
            </div>
            
            <div className="mt-3 text-xs text-muted-foreground">
              <p>Version 0.1.0</p>
              <p className="flex items-center gap-1">
                Status: 
                <span className="flex items-center gap-1">
                  <span className="w-2 h-2 bg-green-500 rounded-full"></span>
                  Online
                </span>
              </p>
            </div>
          </div>
        </div>
      </aside>
    </>
  )
}