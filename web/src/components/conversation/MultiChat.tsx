
import { ConversationTabs } from './ConversationTabs'
import { ChatMessage } from '../ChatMessage'
import { ChatInput } from '../ChatInput'
import { TypingIndicator } from '../TypingIndicator'
import { Button } from '../ui/button'
import { Checkbox } from '../ui/checkbox'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '../ui/dialog'
import { History, FolderOpen, Brain, Trash2 } from 'lucide-react'

import { Conversation, Message } from '../../types/conversation'
import { useState, useEffect } from 'react'

interface MultiChatProps {
  conversations: Conversation[]
  activeConversationId: string
  activeConversation: Conversation | undefined
  isLoading: boolean
  isSaving: (id: string) => boolean
  
  onConversationCreate: () => void
  onConversationSelect: (id: string) => Promise<void>
  onConversationLoad: (id: string) => Promise<void>
  onConversationClose: (id: string) => void
  onConversationDelete: (id: string) => void
  onConversationRename: (id: string, newTitle: string) => void
  onSendMessage: (content: string) => Promise<void>
}

export function MultiChat({
  conversations,
  activeConversationId,
  activeConversation,
  isLoading,
  isSaving,
  onConversationCreate,
  onConversationSelect,
  onConversationLoad,
  onConversationClose,
  onConversationDelete,
  onConversationRename,
  onSendMessage
}: MultiChatProps) {
  const [isLoadDialogOpen, setIsLoadDialogOpen] = useState(false)

  const handleSendMessage = async (content: string) => {
    await onSendMessage(content)
  }

  return (
    <div className="flex flex-col h-full min-h-0">
      {/* 对话标签 */}
      <div className="flex-shrink-0">
        <div className="flex items-center gap-2 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 p-2">
          <div className="flex-1">
            <ConversationTabs
              conversations={conversations}
              activeConversationId={activeConversationId}
              isSaving={isSaving}
              onConversationSelect={onConversationSelect}
              onConversationCreate={onConversationCreate}
              onConversationClose={onConversationClose}
              onConversationDelete={onConversationDelete}
              onConversationRename={onConversationRename}
            />
          </div>
          
          {/* 加载历史对话按钮 */}
          <Dialog open={isLoadDialogOpen} onOpenChange={setIsLoadDialogOpen}>
            <DialogTrigger asChild>
              <Button variant="outline" size="sm" className="h-8">
                <History className="w-4 h-4 mr-1" />
                加载历史
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-md">
              <DialogHeader>
                <DialogTitle className="flex items-center gap-2">
                  <FolderOpen className="w-5 h-5" />
                  加载历史对话
                </DialogTitle>
              </DialogHeader>
              <LoadHistoryDialog 
                currentConversations={conversations}
                onLoadConversation={async (id) => {
                  await onConversationLoad(id)
                  setIsLoadDialogOpen(false)
                }}
              />
            </DialogContent>
          </Dialog>
        </div>
      </div>

      {/* 消息区域 - 使用原生滚动 */}
      <div className="flex-1 overflow-y-auto px-4 py-4">
        <div className="space-y-6">
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
      </div>

      {/* 输入区域 - 使用 flexbox 固定在底部 */}
      <div className="flex-shrink-0 border-t bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 p-4">
        <ChatInput
          onSendMessage={handleSendMessage}
          isLoading={isLoading}
        />
      </div>
    </div>
  )
}

// 加载历史对话对话框组件
interface LoadHistoryDialogProps {
  currentConversations: Conversation[]
  onLoadConversation: (id: string) => Promise<void>
}

function LoadHistoryDialog({ currentConversations, onLoadConversation }: LoadHistoryDialogProps) {
  const [availableConversations, setAvailableConversations] = useState<Conversation[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [refreshKey, setRefreshKey] = useState(0)
  const [selectedConversations, setSelectedConversations] = useState<string[]>([])
  const [isMultiSelectMode, setIsMultiSelectMode] = useState(false)

  // 获取所有历史对话文件
  const loadAvailableConversations = async () => {
    setIsLoading(true)
    try {
      const response = await fetch('http://localhost:8080/api/conversations')
      const data = await response.json()
      
      // 显示所有历史对话（包括当前已经在界面中的）
      const allConversations = data.map((conv: any) => ({
        ...conv,
        createdAt: new Date(conv.createdAt),
        updatedAt: new Date(conv.updatedAt)
      }))
      
      setAvailableConversations(allConversations)
    } catch (error) {
      console.error('Failed to load available conversations:', error)
    } finally {
      setIsLoading(false)
    }
  }

  // 对话框打开时获取可用对话，依赖currentConversations来重新过滤
  useEffect(() => {
    loadAvailableConversations()
    setSelectedConversations([])
  }, [currentConversations, refreshKey])

  // 批量删除对话
  const handleBatchDelete = async (conversationIds: string[]) => {
    if (conversationIds.length === 0) return
    
    const conversationNames = availableConversations
      .filter(conv => conversationIds.includes(conv.id))
      .map(conv => conv.title)
      .join('、')
    
    if (window.confirm(`确定要删除以下 ${conversationIds.length} 个历史对话吗？\n\n${conversationNames}\n\n此操作不可恢复。`)) {
      let successCount = 0
      for (const id of conversationIds) {
        try {
          const response = await fetch(`http://localhost:8080/api/conversations/${id}`, {
            method: 'DELETE'
          })
          if (response.ok) {
            successCount++
          }
        } catch (error) {
          console.error('Failed to delete conversation:', error)
        }
      }
      
      if (successCount > 0) {
        setSelectedConversations([])
        await loadAvailableConversations()
      }
    }
  }

  // 切换选择状态
  const handleConversationSelect = (conversationId: string, checked: boolean) => {
    if (checked) {
      setSelectedConversations(prev => [...prev, conversationId])
    } else {
      setSelectedConversations(prev => prev.filter(id => id !== conversationId))
    }
  }

  // 全选/取消全选
  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedConversations(availableConversations.map(conv => conv.id))
    } else {
      setSelectedConversations([])
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="text-sm text-muted-foreground">
          历史对话列表：
        </div>
        
        {/* 批量操作控制 */}
        <div className="flex items-center gap-2">
          {selectedConversations.length > 0 && (
            <>
              <span className="text-sm text-muted-foreground">
                已选择 {selectedConversations.length} 个
              </span>
              <Button
                variant="destructive"
                size="sm"
                onClick={() => handleBatchDelete(selectedConversations)}
                className="h-7 px-2"
              >
                <Trash2 className="w-3 h-3" />
              </Button>
            </>
          )}
          
          <Button
            variant={isMultiSelectMode ? "default" : "outline"}
            size="sm"
            onClick={() => {
              setIsMultiSelectMode(!isMultiSelectMode)
              setSelectedConversations([])
            }}
            className="h-7 px-2 text-xs"
          >
            {isMultiSelectMode ? '退出批量' : '批量删除'}
          </Button>
        </div>
      </div>
      
      {isLoading ? (
        <div className="text-center py-4">加载中...</div>
      ) : availableConversations.length === 0 ? (
        <div className="text-center py-4 text-muted-foreground">
          没有可加载的历史对话
        </div>
      ) : (
        <>
          {/* 全选行 */}
          {isMultiSelectMode && availableConversations.length > 0 && (
            <div className="flex items-center gap-3 p-2 bg-muted/50 rounded-md">
              <Checkbox
                checked={selectedConversations.length === availableConversations.length}
                onCheckedChange={handleSelectAll}
              />
              <span className="text-sm font-medium">
                全选 ({availableConversations.length} 个对话)
              </span>
            </div>
          )}
          
          <div className="space-y-2 max-h-60 overflow-y-auto">
            {availableConversations.map((conv) => (
              <div
                key={conv.id}
                className={`
                  p-3 border rounded-lg hover:bg-accent cursor-pointer transition-colors
                  ${selectedConversations.includes(conv.id) ? 'bg-accent/50' : ''}
                `}
                onClick={() => {
                  if (isMultiSelectMode) {
                    handleConversationSelect(conv.id, !selectedConversations.includes(conv.id))
                  } else {
                    onLoadConversation(conv.id)
                  }
                }}
                title={
                  currentConversations.some(c => c.id === conv.id) 
                    ? "此对话已在当前界面中" 
                    : "点击加载此对话"
                }
              >
                <div className="flex items-start gap-3">
                  {/* 批量选择复选框 */}
                  {isMultiSelectMode && (
                    <div className="flex-shrink-0 mt-1">
                      <Checkbox
                        checked={selectedConversations.includes(conv.id)}
                        onCheckedChange={(checked) => handleConversationSelect(conv.id, checked)}
                        onClick={(e) => e.stopPropagation()}
                      />
                    </div>
                  )}
                  
                  <div className="flex-1 min-w-0">
                    <div className="font-medium text-sm truncate">{conv.title}</div>
                    <div className="text-xs text-muted-foreground">
                      {conv.messages.length} 条消息 • {conv.updatedAt.toLocaleString()}
                    </div>
                    {isMultiSelectMode && selectedConversations.includes(conv.id) && (
                      <span className="text-xs bg-primary text-primary-foreground px-1.5 py-0.5 rounded inline-block mt-1">
                        已选择
                      </span>
                    )}
                    {!isMultiSelectMode && currentConversations.some(c => c.id === conv.id) && (
                      <span className="text-xs bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200 px-1.5 py-0.5 rounded inline-block mt-1">
                        已加载
                      </span>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </>
      )}
      
      <div className="pt-2 border-t">
        <Button 
          variant="outline" 
          size="sm" 
          onClick={() => {
            setRefreshKey(prev => prev + 1)
          }}
          disabled={isLoading}
          className="w-full"
        >
          <History className="w-4 h-4 mr-2" />
          刷新列表
        </Button>
      </div>
    </div>
  )
}