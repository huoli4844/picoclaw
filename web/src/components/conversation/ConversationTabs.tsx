import { useState, useEffect } from 'react'
import { Conversation } from '../../types/conversation'
import { Button } from '../ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '../ui/dialog'
import { Input } from '../ui/input'
import { X, Plus, Edit2, Check, MessageSquare, History, Clock } from 'lucide-react'

interface ConversationTabsProps {
  conversations: Conversation[]
  activeConversationId: string
  onConversationSelect: (id: string) => void
  onConversationCreate: (title?: string) => void
  onConversationClose: (id: string) => void
  onConversationDelete: (id: string) => void
  onConversationRename: (id: string, newTitle: string) => void
  onLoadConversation?: (id: string) => Promise<void>
}

export function ConversationTabs({
  conversations,
  activeConversationId,
  onConversationSelect,
  onConversationCreate,
  onConversationClose,
  onConversationDelete,
  onConversationRename,
  onLoadConversation
}: ConversationTabsProps) {
  const [editingConversation, setEditingConversation] = useState<Conversation | null>(null)
  const [editingTitle, setEditingTitle] = useState('')
  const [isNewConversationDialogOpen, setIsNewConversationDialogOpen] = useState(false)
  const [newConversationTitle, setNewConversationTitle] = useState('')
  const [historyConversations, setHistoryConversations] = useState<Conversation[]>([])
  const [isLoadingHistory, setIsLoadingHistory] = useState(false)
  const [refreshKey, setRefreshKey] = useState(0)

  const handleStartEdit = (conversation: Conversation) => {
    setEditingConversation(conversation)
    setEditingTitle(conversation.title)
  }

  const handleSaveEdit = () => {
    if (editingConversation && editingTitle.trim() && editingTitle.trim() !== editingConversation.title) {
      onConversationRename(editingConversation.id, editingTitle.trim())
    }
    setEditingConversation(null)
    setEditingTitle('')
  }

  const handleCancelEdit = () => {
    setEditingConversation(null)
    setEditingTitle('')
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSaveEdit()
    } else if (e.key === 'Escape') {
      handleCancelEdit()
    }
  }

  // 加载历史对话的函数
  const loadAvailableConversations = async () => {
    setIsLoadingHistory(true)
    
    try {
      // 获取所有历史对话文件，使用与LoadHistoryDialog相同的API
      const response = await fetch('http://localhost:8080/api/conversations')
      const data = await response.json()
      
      // 过滤出当前不在界面中的对话
      const currentIds = new Set(conversations.map(conv => conv.id))
      const availableConversations = data.filter((conv: Conversation) => !currentIds.has(conv.id))
      
      setHistoryConversations(availableConversations.map((conv: any) => ({
        ...conv,
        createdAt: new Date(conv.createdAt),
        updatedAt: new Date(conv.updatedAt)
      })))
    } catch (error) {
      console.error('Failed to load available conversations:', error)
    } finally {
      setIsLoadingHistory(false)
    }
  }

  const handleNewConversation = async () => {
    setIsNewConversationDialogOpen(true)
    setNewConversationTitle('')
    await loadAvailableConversations()
  }

  // 当对话框打开且依赖变化时重新加载
  useEffect(() => {
    if (isNewConversationDialogOpen) {
      loadAvailableConversations()
    }
  }, [conversations, refreshKey, isNewConversationDialogOpen])

  const handleSelectHistoryConversation = async (conversationId: string) => {
    try {
      if (onLoadConversation) {
        await onLoadConversation(conversationId)
      } else {
        await onConversationSelect(conversationId)
      }
      setIsNewConversationDialogOpen(false)
      setNewConversationTitle('')
      setHistoryConversations([]) // 清空历史对话缓存
    } catch (error) {
      console.error('Failed to load conversation:', error)
    }
  }

  const handleCreateNewConversation = () => {
    if (newConversationTitle.trim()) {
      // 创建新对话并使用自定义标题
      onConversationCreate(newConversationTitle.trim())
      setIsNewConversationDialogOpen(false)
      setNewConversationTitle('')
    }
  }

  const handleCancelNewConversation = () => {
    setIsNewConversationDialogOpen(false)
    setNewConversationTitle('')
  }

  const handleNewConversationKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleCreateNewConversation()
    } else if (e.key === 'Escape') {
      handleCancelNewConversation()
    }
  }

  return (
    <>
      <div className="border-b border-border bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="flex items-center gap-1 p-2 overflow-x-auto">
          {/* 新建对话按钮 */}
          <Button
            variant="ghost"
            size="sm"
            onClick={handleNewConversation}
            className="flex items-center gap-2 px-3 py-1.5 min-w-0"
          >
            <Plus className="w-4 h-4 flex-shrink-0" />
            <span className="hidden sm:inline">新建对话</span>
          </Button>

          {/* 对话标签 */}
          {conversations.map((conversation) => {
            const isActive = conversation.id === activeConversationId

            return (
              <div
                key={conversation.id}
                className={`
                  group relative flex items-center gap-2 px-3 py-1.5 rounded-md border min-w-0 max-w-xs
                  transition-colors cursor-pointer
                  ${isActive 
                    ? 'bg-primary text-primary-foreground border-primary' 
                    : 'border-border hover:bg-accent hover:text-accent-foreground'
                  }
                `}
              >
                <MessageSquare className="w-4 h-4 flex-shrink-0" />
                <button
                  onClick={async () => {
                    await onConversationSelect(conversation.id)
                  }}
                  className="flex-1 text-left truncate text-sm font-medium"
                >
                  {conversation.title}
                </button>
                
                {/* 编辑按钮 */}
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={(e) => {
                    e.stopPropagation()
                    handleStartEdit(conversation)
                  }}
                  className={`h-4 w-4 p-0 transition-all duration-300
                    ${isActive 
                      ? 'opacity-100 text-primary-foreground hover:bg-white/20' 
                      : 'opacity-0 group-hover:opacity-100 text-muted-foreground hover:bg-accent/20'}
                  `}
                >
                  <Edit2 className="w-3 h-3" />
                </Button>
                
                {/* 关闭按钮 - 点击只是从界面移除，不删除文件 */}
                {conversations.length > 1 && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={(e) => {
                      e.stopPropagation()
                      onConversationClose(conversation.id)
                    }}
                    onContextMenu={(e) => {
                      e.preventDefault()
                      e.stopPropagation()
                      if (window.confirm('确定要永久删除这个对话吗？删除后无法恢复。')) {
                        onConversationDelete(conversation.id)
                      }
                    }}
                    className={`h-4 w-4 p-0 transition-all duration-300
                      ${isActive 
                        ? 'opacity-100 text-primary-foreground hover:bg-white/20' 
                        : 'opacity-0 group-hover:opacity-100 text-muted-foreground hover:bg-accent/20'}
                    `}
                    title="点击关闭，右键删除"
                  >
                    <X className="w-3 h-3" />
                  </Button>
                )}
              </div>
            )
          })}
        </div>
      </div>

      {/* 编辑对话名称对话框 */}
      <Dialog open={!!editingConversation} onOpenChange={(open) => !open && handleCancelEdit()}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>重命名对话</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <label htmlFor="title" className="text-sm font-medium">
                对话名称
              </label>
              <Input
                id="title"
                value={editingTitle}
                onChange={(e) => setEditingTitle(e.target.value)}
                onKeyDown={handleKeyPress}
                placeholder="请输入新的对话名称"
                className="mt-1"
                autoFocus
              />
            </div>
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={handleCancelEdit}>
                取消
              </Button>
              <Button 
                onClick={handleSaveEdit}
                disabled={!editingTitle.trim() || editingTitle.trim() === editingConversation?.title}
              >
                保存
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* 新建对话对话框 */}
      <Dialog open={isNewConversationDialogOpen} onOpenChange={(open) => !open && handleCancelNewConversation()}>
        <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Plus className="w-5 h-5" />
              新建对话
            </DialogTitle>
          </DialogHeader>
          <div className="space-y-6">
            {/* 创建新对话部分 */}
            <div className="space-y-3">
              <h3 className="text-sm font-medium text-muted-foreground">创建新对话</h3>
              <div>
                <label htmlFor="new-title" className="text-sm font-medium">
                  对话名称
                </label>
                <Input
                  id="new-title"
                  value={newConversationTitle}
                  onChange={(e) => setNewConversationTitle(e.target.value)}
                  onKeyDown={handleNewConversationKeyPress}
                  placeholder="请输入对话名称"
                  className="mt-1"
                  autoFocus
                />
              </div>
              <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={handleCancelNewConversation}>
                  取消
                </Button>
                <Button 
                  onClick={handleCreateNewConversation}
                  disabled={!newConversationTitle.trim()}
                >
                  创建
                </Button>
              </div>
            </div>

            {/* 分隔线 */}
            <div className="border-t"></div>

            {/* 历史对话部分 */}
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <h3 className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                  <History className="w-4 h-4" />
                  选择历史对话
                </h3>
                <Button 
                  variant="ghost" 
                  size="sm" 
                  onClick={() => setRefreshKey(prev => prev + 1)}
                  disabled={isLoadingHistory}
                  className="h-7 px-2"
                >
                  <Clock className="w-3 h-3 mr-1" />
                  刷新
                </Button>
              </div>
              <div className="space-y-2 max-h-60 overflow-y-auto">
                {isLoadingHistory ? (
                  <div className="text-center py-4 text-muted-foreground">
                    <Clock className="w-8 h-8 mx-auto mb-2 opacity-50 animate-pulse" />
                    <p className="text-sm">加载历史对话中...</p>
                  </div>
                ) : historyConversations.length === 0 ? (
                  <div className="text-center py-8 text-muted-foreground">
                    <Clock className="w-8 h-8 mx-auto mb-2 opacity-50" />
                    <p className="text-sm">暂无历史对话</p>
                    <p className="text-xs mt-1">所有历史对话已在当前会话中</p>
                  </div>
                ) : (
                  historyConversations.map((conversation) => (
                    <div
                      key={conversation.id}
                      className="flex items-center gap-3 p-3 rounded-lg border border-border hover:bg-accent hover:text-accent-foreground hover:border-accent cursor-pointer transition-all duration-200"
                      onClick={() => handleSelectHistoryConversation(conversation.id)}
                    >
                      <MessageSquare className="w-4 h-4 flex-shrink-0" />
                      <div className="flex-1 min-w-0">
                        <p className="font-medium truncate">{conversation.title}</p>
                        <p className="text-xs opacity-70">
                          {conversation.messages.length} 条消息 • {conversation.updatedAt.toLocaleDateString()}
                        </p>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </>
  )
}