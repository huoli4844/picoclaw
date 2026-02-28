import { useState } from 'react'
import { Conversation } from '../../types/conversation'
import { Button } from '../ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '../ui/dialog'
import { Input } from '../ui/input'
import { X, Plus, Edit2, Check, MessageSquare } from 'lucide-react'

interface ConversationTabsProps {
  conversations: Conversation[]
  activeConversationId: string
  onConversationSelect: (id: string) => void
  onConversationCreate: () => void
  onConversationClose: (id: string) => void
  onConversationDelete: (id: string) => void
  onConversationRename: (id: string, newTitle: string) => void
}

export function ConversationTabs({
  conversations,
  activeConversationId,
  onConversationSelect,
  onConversationCreate,
  onConversationClose,
  onConversationDelete,
  onConversationRename
}: ConversationTabsProps) {
  const [editingConversation, setEditingConversation] = useState<Conversation | null>(null)
  const [editingTitle, setEditingTitle] = useState('')

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

  return (
    <>
      <div className="border-b border-border bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="flex items-center gap-1 p-2 overflow-x-auto">
          {/* 新建对话按钮 */}
          <Button
            variant="ghost"
            size="sm"
            onClick={onConversationCreate}
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
    </>
  )
}