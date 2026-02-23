
import { Avatar, AvatarFallback } from './ui/avatar'
import { formatMessageTime } from '@/lib/utils'
import { Message } from '@/types'
import { Bot, User } from 'lucide-react'
import { ThoughtProcess } from './ThoughtProcess'
import { MarkdownRenderer } from './ui/markdown-renderer'

interface ChatMessageProps {
  message: Message
}

export function ChatMessage({ message }: ChatMessageProps) {
  const isUser = message.role === 'user'

  if (isUser) {
    return (
      <div className={`chat-message user`}>
        <div className="chat-avatar user">
          <Avatar className="w-8 h-8">
            <AvatarFallback className="bg-primary text-primary-foreground">
              <User className="w-4 h-4" />
            </AvatarFallback>
          </Avatar>
        </div>
        <div className="chat-content user">
          <div className="flex flex-col">
            <div className="text-sm whitespace-pre-wrap break-words leading-relaxed">{message.content}</div>
            <div className="flex items-center gap-2 text-xs mt-2 text-primary-foreground/70">
              <span>{formatMessageTime(message.timestamp)}</span>
              {message.model && <span>• {message.model}</span>}
            </div>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className={`chat-message assistant`}>
      <div className="chat-avatar assistant">
        <Avatar className="w-8 h-8">
          <AvatarFallback className="bg-secondary text-secondary-foreground">
            <Bot className="w-4 h-4" />
          </AvatarFallback>
        </Avatar>
      </div>
      <div className="chat-content assistant">
        <div className="flex flex-col space-y-4">
          <MarkdownRenderer 
            content={message.content} 
            className="text-sm leading-relaxed"
          />
          
          {/* 显示思考过程 */}
          {message.thoughts && message.thoughts.length > 0 && (
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