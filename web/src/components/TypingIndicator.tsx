
import { Avatar, AvatarFallback } from './ui/avatar'
import { Bot } from 'lucide-react'

export function TypingIndicator() {
  return (
    <div className="chat-message assistant">
      <div className="chat-avatar">
        <Avatar className="w-8 h-8">
          <AvatarFallback className="bg-secondary text-secondary-foreground">
            <Bot className="w-4 h-4" />
          </AvatarFallback>
        </Avatar>
      </div>
      <div className="chat-content">
        <div className="typing-indicator">
          <span></span>
          <span></span>
          <span></span>
        </div>
      </div>
    </div>
  )
}