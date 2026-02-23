
import { Avatar, AvatarFallback } from './ui/avatar'
import { formatMessageTime } from '@/lib/utils'
import { Message } from '@/types'
import { Bot, User } from 'lucide-react'
import { ThoughtProcess } from './ThoughtProcess'

interface ChatMessageProps {
  message: Message
}

export function ChatMessage({ message }: ChatMessageProps) {
  const isUser = message.role === 'user'

  return (
    <div className={`chat-message ${message.role}`}>
      <div className="chat-avatar">
        <Avatar className="w-8 h-8">
          <AvatarFallback className={isUser ? 'bg-primary text-primary-foreground' : 'bg-secondary text-secondary-foreground'}>
            {isUser ? <User className="w-4 h-4" /> : <Bot className="w-4 h-4" />}
          </AvatarFallback>
        </Avatar>
      </div>
      <div className="chat-content">
        <div className="flex flex-col space-y-3">
          <p className="text-sm whitespace-pre-wrap break-words">{message.content}</p>
          
          {/* 显示思考过程 */}
          {!isUser && message.thoughts && message.thoughts.length > 0 && (
            <ThoughtProcess thoughts={message.thoughts} />
          )}
          
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <span>{formatMessageTime(message.timestamp)}</span>
            {message.model && <span>• {message.model}</span>}
          </div>
        </div>
      </div>
    </div>
  )
}