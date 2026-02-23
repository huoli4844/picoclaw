import { useState, useEffect } from 'react'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Loader2, Package } from 'lucide-react'
import { useApi } from '@/hooks/useApi'
import { Skill, SkillDetail } from '@/types'

interface SkillDetailDialogProps {
  isOpen: boolean
  onClose: () => void
  skill: Skill | null
}

export function SkillDetailDialog({ isOpen, onClose, skill }: SkillDetailDialogProps) {
  const [skillDetail, setSkillDetail] = useState<SkillDetail | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const { getSkillDetail } = useApi()

  useEffect(() => {
    if (isOpen && skill) {
      loadSkillDetail()
    }
  }, [isOpen, skill])

  const loadSkillDetail = async () => {
    if (!skill) return

    setIsLoading(true)
    try {
      const result = await getSkillDetail(skill.name)
      if (result.success && result.data) {
        setSkillDetail(result.data)
      }
    } catch (error) {
      console.error('Failed to load skill detail:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const getSourceBadgeVariant = (source: string) => {
    switch (source) {
      case 'workspace':
        return 'default'
      case 'global':
        return 'secondary'
      case 'builtin':
        return 'outline'
      default:
        return 'outline'
    }
  }

  if (!skill) return null

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl max-h-[80vh]">
        <DialogHeader>
          <div className="flex items-center gap-2">
            <Package className="w-5 h-5" />
            <DialogTitle>{skill.name}</DialogTitle>
            <Badge variant={getSourceBadgeVariant(skill.source)}>
              {skill.source}
            </Badge>
          </div>
        </DialogHeader>
        
        <div className="flex flex-col h-[60vh]">
          {isLoading ? (
            <div className="flex items-center justify-center flex-1">
              <div className="flex items-center gap-2 text-muted-foreground">
                <Loader2 className="w-6 h-6 animate-spin" />
                <span>加载技能详情...</span>
              </div>
            </div>
          ) : skillDetail ? (
            <>
              <div className="mb-4">
                <h3 className="text-sm font-medium text-muted-foreground mb-2">描述</h3>
                <p className="text-sm">{skillDetail.description}</p>
              </div>

              <div className="mb-4">
                <h3 className="text-sm font-medium text-muted-foreground mb-2">路径</h3>
                <code className="text-xs bg-muted px-2 py-1 rounded">{skillDetail.path}</code>
              </div>

              {Object.keys(skillDetail.metadata).length > 0 && (
                <div className="mb-4">
                  <h3 className="text-sm font-medium text-muted-foreground mb-2">元数据</h3>
                  <div className="grid grid-cols-2 gap-2">
                    {Object.entries(skillDetail.metadata).map(([key, value]) => (
                      <div key={key} className="flex items-center gap-2">
                        <span className="text-xs font-medium">{key}:</span>
                        <span className="text-xs text-muted-foreground">{value}</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              <div className="flex-1">
                <h3 className="text-sm font-medium text-muted-foreground mb-2">技能内容</h3>
                <ScrollArea className="h-full border rounded-md p-3">
                  <pre className="text-xs whitespace-pre-wrap font-mono">
                    {skillDetail.content}
                  </pre>
                </ScrollArea>
              </div>
            </>
          ) : (
            <div className="flex items-center justify-center flex-1">
              <p className="text-muted-foreground">无法加载技能详情</p>
            </div>
          )}
        </div>

        <div className="flex justify-end gap-2 mt-4">
          <Button variant="outline" onClick={onClose}>
            关闭
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}