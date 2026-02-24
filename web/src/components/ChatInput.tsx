import React, { useState, useRef } from 'react'
import { Button } from './ui/button'
import { Send, Loader2, Paperclip } from 'lucide-react'

interface ChatInputProps {
  onSendMessage: (message: string) => void
  isLoading: boolean
  disabled?: boolean
}

export function ChatInput({ onSendMessage, isLoading, disabled }: ChatInputProps) {
  const [input, setInput] = useState('')
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (input.trim() && !isLoading && !disabled) {
      onSendMessage(input.trim())
      setInput('')
      // 重置textarea高度
      if (textareaRef.current) {
        textareaRef.current.style.height = 'auto'
      }
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && e.altKey) {
      e.preventDefault()
      handleSubmit(e)
    }
  }

  const handleInputChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setInput(e.target.value)
    
    // 自动调整高度
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
      textareaRef.current.style.height = `${Math.min(textareaRef.current.scrollHeight, 120)}px`
    }
  }

  return (
    <div className="relative">
      <form onSubmit={handleSubmit} className="relative">
        <div className="flex items-end gap-3 bg-muted/50 rounded-xl border border-border/50 p-3 transition-all duration-200 focus-within:border-primary/50 focus-within:bg-background/80">
          <Button 
            type="button"
            variant="ghost" 
            size="sm"
            className="shrink-0 text-muted-foreground hover:text-foreground"
            disabled={isLoading || disabled}
          >
            <Paperclip className="w-4 h-4" />
          </Button>
          
          <textarea
            ref={textareaRef}
            value={input}
            onChange={handleInputChange}
            onKeyDown={handleKeyDown}
            placeholder="输入消息... (Enter 换行，Alt+Enter 发送)"
            className="flex-1 min-h-[24px] max-h-[120px] resize-none bg-transparent border-none outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50 text-sm leading-relaxed"
            disabled={isLoading || disabled}
            rows={1}
          />
          
          <Button 
            type="submit" 
            disabled={!input.trim() || isLoading || disabled}
            size="sm"
            className="shrink-0 hover-lift"
          >
            {isLoading ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <Send className="w-4 h-4" />
            )}
          </Button>
        </div>
        
        {/* 字符计数 */}
        {input.length > 100 && (
          <div className="absolute -top-2 right-3 text-xs text-muted-foreground bg-background px-2 py-1 rounded-full border">
            {input.length}/2000
          </div>
        )}
      </form>
    </div>
  )
}