import { useState, useEffect } from 'react'
import { Search, Package, Plus, Eye, Trash2, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from '@/components/ui/alert-dialog'
import { useApi } from '@/hooks/useApi'
import { Skill } from '@/types'
import { SkillsSearch } from './SkillsSearch'
import { SkillDetailDialog } from './SkillDetailDialog'

interface SkillsPageProps {
  onBack: () => void
}

export function SkillsPage({ onBack }: SkillsPageProps) {
  const [skills, setSkills] = useState<Skill[]>([])
  const [searchQuery, setSearchQuery] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [selectedSkill, setSelectedSkill] = useState<Skill | null>(null)
  const [isDetailOpen, setIsDetailOpen] = useState(false)
  const [isSearchOpen, setIsSearchOpen] = useState(false)
  const [skillToDelete, setSkillToDelete] = useState<Skill | null>(null)
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  const { getSkills, uninstallSkill } = useApi()

  const loadSkills = async () => {
    setIsLoading(true)
    try {
      const result = await getSkills()
      if (result.success && result.data) {
        setSkills(result.data)
      }
    } catch (error) {
      console.error('Failed to load skills:', error)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    loadSkills()
  }, [])

  const filteredSkills = skills.filter(skill =>
    skill.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    skill.description.toLowerCase().includes(searchQuery.toLowerCase())
  )

  const handleSkillClick = (skill: Skill) => {
    setSelectedSkill(skill)
    setIsDetailOpen(true)
  }

  const handleSkillInstalled = () => {
    loadSkills()
    setIsSearchOpen(false)
  }

  const handleDeleteSkill = (skill: Skill) => {
    setSkillToDelete(skill)
    setIsDeleteDialogOpen(true)
  }

  const confirmDeleteSkill = async () => {
    if (!skillToDelete) return
    
    setIsDeleting(true)
    try {
      const result = await uninstallSkill(skillToDelete.name)
      if (result.success) {
        await loadSkills() // 重新加载技能列表
        setIsDeleteDialogOpen(false)
        setSkillToDelete(null)
      } else {
        console.error('Failed to uninstall skill:', result.error)
        // 这里可以添加错误提示
      }
    } catch (error) {
      console.error('Error uninstalling skill:', error)
      // 这里可以添加错误提示
    } finally {
      setIsDeleting(false)
    }
  }

  const cancelDeleteSkill = () => {
    setIsDeleteDialogOpen(false)
    setSkillToDelete(null)
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

  return (
    <div className="h-full flex flex-col">
      {/* Search Bar */}
      <div className="border-b p-4">
        <div className="max-w-6xl mx-auto">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Package className="w-6 h-6 text-primary" />
              <div>
                <h2 className="text-lg font-semibold">技能管理</h2>
                <p className="text-sm text-muted-foreground">管理 PicoClaw 技能</p>
              </div>
            </div>
            </div>
            <div className="flex items-center gap-2">
              <div className="relative mr-2">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground w-4 h-4" />
                <Input
                  placeholder="搜索技能..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10 w-64"
                />
              </div>
              <Button
                variant="outline"
                onClick={() => setIsSearchOpen(true)}
                className="flex items-center gap-2"
              >
                <Plus className="w-4 h-4" />
                安装技能
              </Button>
            </div>
          </div>
        </div>
      </div>

      {/* Skills List */}
      <ScrollArea className="flex-1">
        <div className="max-w-6xl mx-auto p-4">
          {isLoading ? (
            <div className="flex items-center justify-center h-64">
              <div className="flex items-center gap-2 text-muted-foreground">
                <Loader2 className="w-6 h-6 animate-spin" />
                <span>加载技能列表...</span>
              </div>
            </div>
          ) : filteredSkills.length === 0 ? (
            <div className="flex items-center justify-center h-64">
              <div className="text-center text-muted-foreground">
                <Package className="w-12 h-12 mx-auto mb-4 opacity-50" />
                <p>暂无技能</p>
                <p className="text-sm mt-2">点击"安装技能"来添加新技能</p>
              </div>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {filteredSkills.map((skill) => (
                <Card 
                  key={skill.name} 
                  className="cursor-pointer hover:shadow-md transition-shadow"
                  onClick={() => handleSkillClick(skill)}
                >
                  <CardHeader className="pb-2">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <CardTitle className="text-lg">{skill.name}</CardTitle>
                        <Badge variant={getSourceBadgeVariant(skill.source)} className="mt-1">
                          {skill.source}
                        </Badge>
                      </div>
                      <div className="flex gap-1">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={(e) => {
                            e.stopPropagation()
                            handleSkillClick(skill)
                          }}
                        >
                          <Eye className="w-4 h-4" />
                        </Button>
                        {skill.source !== 'builtin' && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={(e) => {
                              e.stopPropagation()
                              handleDeleteSkill(skill)
                            }}
                            className="text-destructive hover:text-destructive hover:bg-destructive/10"
                          >
                            <Trash2 className="w-4 h-4" />
                          </Button>
                        )}
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent>
                    <CardDescription className="line-clamp-3">
                      {skill.description}
                    </CardDescription>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </div>
      </ScrollArea>

      {/* Skill Detail Dialog */}
      <SkillDetailDialog
        isOpen={isDetailOpen}
        onClose={() => {
          setIsDetailOpen(false)
          setSelectedSkill(null)
        }}
        skill={selectedSkill}
      />

      {/* Skills Search Dialog */}
      <SkillsSearch
        isOpen={isSearchOpen}
        onClose={() => setIsSearchOpen(false)}
        onSkillInstalled={handleSkillInstalled}
      />

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={isDeleteDialogOpen} onOpenChange={setIsDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>删除技能</AlertDialogTitle>
            <AlertDialogDescription>
              确定要删除技能 "<span className="font-semibold">{skillToDelete?.name}</span>" 吗？
              <br />
              <span className="text-sm text-muted-foreground">
                此操作将删除技能文件且不可恢复。
              </span>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={cancelDeleteSkill} disabled={isDeleting}>
              取消
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmDeleteSkill}
              disabled={isDeleting}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {isDeleting ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  删除中...
                </>
              ) : (
                '删除'
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}