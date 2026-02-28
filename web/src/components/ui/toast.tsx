import { useState, useEffect } from 'react'
import { X, CheckCircle, AlertCircle, Info } from 'lucide-react'
import { Button } from './button'

interface ToastProps {
  message: string
  type: 'success' | 'error' | 'info'
  onClose: () => void
}

function Toast({ message, type, onClose }: ToastProps) {
  const [isVisible, setIsVisible] = useState(false)

  useEffect(() => {
    // 触发进入动画
    const timer = setTimeout(() => setIsVisible(true), 100)
    
    // 自动关闭
    const autoCloseTimer = setTimeout(() => {
      setIsVisible(false)
      setTimeout(onClose, 300) // 等待退出动画完成
    }, 3000)

    return () => {
      clearTimeout(timer)
      clearTimeout(autoCloseTimer)
    }
  }, [onClose])

  const getIcon = () => {
    switch (type) {
      case 'success':
        return <CheckCircle className="w-5 h-5 text-green-500" />
      case 'error':
        return <AlertCircle className="w-5 h-5 text-red-500" />
      case 'info':
        return <Info className="w-5 h-5 text-blue-500" />
    }
  }

  const getBgColor = () => {
    switch (type) {
      case 'success':
        return 'bg-green-50 border-green-200'
      case 'error':
        return 'bg-red-50 border-red-200'
      case 'info':
        return 'bg-blue-50 border-blue-200'
    }
  }

  return (
    <div className={`
      fixed top-4 right-4 z-50 max-w-sm p-4 rounded-lg border shadow-lg
      flex items-start gap-3 transition-all duration-300 transform
      ${getBgColor()}
      ${isVisible 
        ? 'translate-x-0 opacity-100 scale-100' 
        : 'translate-x-full opacity-0 scale-95'
      }
    `}>
      {getIcon()}
      <div className="flex-1">
        <p className="text-sm font-medium text-gray-900">{message}</p>
      </div>
      <Button
        variant="ghost"
        size="sm"
        onClick={() => {
          setIsVisible(false)
          setTimeout(onClose, 300)
        }}
        className="h-6 w-6 p-0 text-gray-400 hover:text-gray-600"
      >
        <X className="w-4 h-4" />
      </Button>
    </div>
  )
}

interface ToastContainerProps {}

interface ToastMessage {
  id: string
  message: string
  type: 'success' | 'error' | 'info'
}

export function ToastContainer({}: ToastContainerProps) {
  const [toasts, setToasts] = useState<ToastMessage[]>([])

  useEffect(() => {
    const handleToast = (event: CustomEvent) => {
      const { message, type } = event.detail
      const id = Date.now().toString()
      setToasts(prev => [...prev, { id, message, type }])
    }

    // 监听自定义事件
    window.addEventListener('toast', handleToast as EventListener)
    
    return () => {
      window.removeEventListener('toast', handleToast as EventListener)
    }
  }, [])

  const removeToast = (id: string) => {
    setToasts(prev => prev.filter(toast => toast.id !== id))
  }

  return (
    <div className="fixed top-0 right-0 z-50 p-4 space-y-2">
      {toasts.map((toast) => (
        <Toast
          key={toast.id}
          message={toast.message}
          type={toast.type}
          onClose={() => removeToast(toast.id)}
        />
      ))}
    </div>
  )
}

// 导出便捷函数
export const showToast = (message: string, type: 'success' | 'error' | 'info' = 'info') => {
  if (window.dispatchEvent) {
    window.dispatchEvent(new CustomEvent('toast', { detail: { message, type } }))
  }
}