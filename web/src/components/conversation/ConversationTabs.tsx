import { useState } from 'react'
import { Conversation } from '../../types/conversation'
import { Button } from '../ui/button'
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
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editingTitle, setEditingTitle] = useState('')

  const handleStartEdit = (conversation: Conversation) => {
    setEditingId(conversation.id)
    setEditingTitle(conversation.title)
  }

  const handleSaveEdit = () => {
    if (editingId && editingTitle.trim()) {
      onConversationRename(editingId, editingTitle.trim())
    }
    setEditingId(null)
    setEditingTitle('')
  }

  const handleCancelEdit = () => {
    setEditingId(null)
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
          const isEditing = editingId === conversation.id

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
              {isEditing ? (
                <div className="flex items-center gap-1 w-full">
                  <Input
                    value={editingTitle}
                    onChange={(e) => setEditingTitle(e.target.value)}
                    onKeyDown={handleKeyPress}
                    onBlur={handleSaveEdit}
                    className="h-6 px-2 py-0 text-sm bg-background border-none focus:ring-1 focus:ring-primary"
                    autoFocus
                  />
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={handleSaveEdit}
                    className="h-4 w-4 p-0 hover:bg-primary/20"
                  >
                    <Check className="w-3 h-3" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={handleCancelEdit}
                    className="h-4 w-4 p-0 hover:bg-destructive/20"
                  >
                    <X className="w-3 h-3" />
                  </Button>
                </div>
              ) : (
                <>
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
                    className={`h-4 w-4 p-0 opacity-0 group-hover:opacity-100 transition-opacity
                      ${isActive ? 'opacity-100 hover:bg-primary/20' : 'hover:bg-accent/20'}
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
                      className={`h-4 w-4 p-0 opacity-0 group-hover:opacity-100 transition-opacity
                        ${isActive ? 'opacity-100 hover:bg-primary/20' : 'hover:bg-accent/20'}
                      `}
                      title="点击关闭，右键删除"
                    >
                      <X className="w-3 h-3" />
                    </Button>
                  )}
                </>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}