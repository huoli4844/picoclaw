
import { Moon, Sun, Monitor } from "lucide-react"
import { useTheme } from "@/contexts/ThemeContext"

export function ThemeToggle() {
  const { theme, setTheme } = useTheme()

  const toggleTheme = () => {
    setTheme(theme === 'light' ? 'dark' : theme === 'dark' ? 'system' : 'light')
  }

  const getIconForTheme = () => {
    switch (theme) {
      case 'light':
        return <Sun className="h-[1.2rem] w-[1.2rem]" />
      case 'dark':
        return <Moon className="h-[1.2rem] w-[1.2rem]" />
      case 'system':
        return <Monitor className="h-[1.2rem] w-[1.2rem]" />
      default:
        return <Sun className="h-[1.2rem] w-[1.2rem]" />
    }
  }

  const getTooltip = () => {
    switch (theme) {
      case 'light':
        return '当前: 亮色主题 (点击切换到暗色)'
      case 'dark':
        return '当前: 暗色主题 (点击切换到跟随系统)'
      case 'system':
        return '当前: 跟随系统 (点击切换到亮色)'
      default:
        return '切换主题'
    }
  }

  return (
    <button
      onClick={toggleTheme}
      className="inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-9 w-9"
      title={getTooltip()}
    >
      <div className="relative">
        {getIconForTheme()}
      </div>
    </button>
  )
}