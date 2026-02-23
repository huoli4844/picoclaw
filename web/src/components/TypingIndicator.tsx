
import { Avatar, AvatarFallback } from './ui/avatar'
import { Bot } from 'lucide-react'

export function TypingIndicator() {
  return (
    <div className="chat-message assistant">
      <div className="chat-avatar assistant">
        <Avatar className="w-8 h-8">
          <AvatarFallback className="bg-secondary text-secondary-foreground">
            <Bot className="w-4 h-4" />
          </AvatarFallback>
        </Avatar>
      </div>
      <div className="chat-content assistant">
        <div className="flex items-center gap-1 text-muted-foreground">
          <span className="typing-indicator">
            <span></span>
            <span></span>
            <span></span>
          </span>
          <span className="text-xs ml-2">AI 正在思考</span>
        </div>
      </div>
    </div>
  )
}