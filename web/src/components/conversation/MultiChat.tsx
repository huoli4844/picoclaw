
import { ConversationTabs } from './ConversationTabs'
import { ChatMessage } from '../ChatMessage'
import { ChatInput } from '../ChatInput'
import { TypingIndicator } from '../TypingIndicator'
import { ScrollArea } from '../ui/scroll-area'
import { Brain } from 'lucide-react'

import { Conversation, Message } from '../../types/conversation'

interface MultiChatProps {
  conversations: Conversation[]
  activeConversationId: string
  activeConversation: Conversation | undefined
  isLoading: boolean
  
  onConversationCreate: () => void
  onConversationSelect: (id: string) => void
  onConversationDelete: (id: string) => void
  onConversationRename: (id: string, newTitle: string) => void
  onSendMessage: (content: string) => Promise<void>
}

export function MultiChat({
  conversations,
  activeConversationId,
  activeConversation,
  isLoading,
  onConversationCreate,
  onConversationSelect,
  onConversationDelete,
  onConversationRename,
  onSendMessage
}: MultiChatProps) {

  const handleSendMessage = async (content: string) => {
    await onSendMessage(content)
  }

  return (
    <div className="flex flex-col h-screen relative">
      {/* 对话标签 */}
      <div className="flex-shrink-0">
        <ConversationTabs
          conversations={conversations}
          activeConversationId={activeConversationId}
          onConversationSelect={onConversationSelect}
          onConversationCreate={onConversationCreate}
          onConversationDelete={onConversationDelete}
          onConversationRename={onConversationRename}
        />
      </div>

      {/* 消息区域 */}
      <div className="flex-1 overflow-hidden relative pb-20">
        <ScrollArea className="h-full">
          <div className="chat-messages w-full px-4 py-4">
            {activeConversation?.messages.length === 0 ? (
              <div className="flex items-center justify-center h-64 text-muted-foreground">
                <div className="text-center">
                  <div className="relative mb-6">
                    <Brain className="w-16 h-16 mx-auto text-primary opacity-20" />
                    <div className="absolute inset-0 flex items-center justify-center">
                      <div className="w-4 h-4 bg-primary rounded-full animate-ping" />
                    </div>
                  </div>
                  <h2 className="text-2xl font-semibold mb-2">开始对话吧！</h2>
                  <p className="text-muted-foreground mb-4">我是 PicoClaw，您的智能 AI 助手</p>
                  <div className="flex flex-wrap justify-center gap-2 text-sm">
                    <span className="px-3 py-1 bg-primary/10 text-primary rounded-full">📝 内容创作</span>
                    <span className="px-3 py-1 bg-primary/10 text-primary rounded-full">💻 编程助手</span>
                    <span className="px-3 py-1 bg-primary/10 text-primary rounded-full">🔍 数据分析</span>
                    <span className="px-3 py-1 bg-primary/10 text-primary rounded-full">🎨 创意设计</span>
                  </div>
                </div>
              </div>
            ) : (
              <>
                {activeConversation?.messages.map((message: Message) => (
                  <ChatMessage key={message.id} message={message} />
                ))}
                {isLoading && <TypingIndicator />}
              </>
            )}
          </div>
        </ScrollArea>
      </div>

      {/* 输入区域 - 固定在底部 */}
      <div className="fixed bottom-0 left-0 right-0 border-t bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 p-4 z-50">
        <div className="max-w-4xl mx-auto">
          <ChatInput
            onSendMessage={handleSendMessage}
            isLoading={isLoading}
          />
        </div>
      </div>
    </div>
  )
}