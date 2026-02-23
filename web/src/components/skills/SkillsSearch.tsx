import { useState } from 'react'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Badge } from '@/components/ui/badge'
import { Search, Download, Loader2, Package } from 'lucide-react'
import { useApi } from '@/hooks/useApi'
import { SearchSkillsRequest, InstallSkillRequest } from '@/types'

interface SkillsSearchProps {
  isOpen: boolean
  onClose: () => void
  onSkillInstalled: () => void
}

interface SearchResult {
  slug: string
  display_name: string
  summary: string
  version: string
  registry_name: string
  score: number
}

export function SkillsSearch({ isOpen, onClose, onSkillInstalled }: SkillsSearchProps) {
  const [searchQuery, setSearchQuery] = useState('')
  const [searchResults, setSearchResults] = useState<SearchResult[]>([])
  const [isSearching, setIsSearching] = useState(false)
  const [isInstalling, setIsInstalling] = useState<string | null>(null)
  const { searchSkills, installSkill } = useApi()

  const handleSearch = async () => {
    if (!searchQuery.trim()) return

    setIsSearching(true)
    try {
      const request: SearchSkillsRequest = {
        query: searchQuery,
        limit: 20
      }
      const result = await searchSkills(request)
      if (result.success && result.data) {
        // Parse the results - the API returns results as a string that needs parsing
        const results = parseSearchResults(result.data)
        setSearchResults(results)
      }
    } catch (error) {
      console.error('Failed to search skills:', error)
    } finally {
      setIsSearching(false)
    }
  }

  const parseSearchResults = (data: any): SearchResult[] => {
    // The search results might be in different formats, try to extract the array
    if (Array.isArray(data.results)) {
      return data.results.map((item: any) => ({
        slug: item.slug || '',
        display_name: item.display_name || item.name || '',
        summary: item.summary || item.description || '',
        version: item.version || 'latest',
        registry_name: item.registry_name || 'clawhub',
        score: item.score || 0
      }))
    }
    return []
  }

  const handleInstall = async (skill: SearchResult) => {
    setIsInstalling(skill.slug)
    try {
      const request: InstallSkillRequest = {
        slug: skill.slug,
        registry: skill.registry_name,
        version: skill.version
      }
      const result = await installSkill(request)
      if (result.success) {
        onSkillInstalled()
        // Remove from search results after successful installation
        setSearchResults(prev => prev.filter(s => s.slug !== skill.slug))
      }
    } catch (error) {
      console.error('Failed to install skill:', error)
    } finally {
      setIsInstalling(null)
    }
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSearch()
    }
  }

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl max-h-[80vh]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Search className="w-5 h-5" />
            搜索和安装技能
          </DialogTitle>
        </DialogHeader>
        
        <div className="flex flex-col h-[60vh]">
          {/* Search Input */}
          <div className="flex gap-2 mb-4">
            <Input
              placeholder="搜索技能 (例如: github, weather, database)..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyPress={handleKeyPress}
              className="flex-1"
            />
            <Button 
              onClick={handleSearch} 
              disabled={!searchQuery.trim() || isSearching}
              className="flex items-center gap-2"
            >
              {isSearching ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : (
                <Search className="w-4 h-4" />
              )}
              搜索
            </Button>
          </div>

          {/* Search Results */}
          <ScrollArea className="flex-1">
            <div className="space-y-3">
              {searchResults.length === 0 && !isSearching && (
                <div className="flex items-center justify-center h-32 text-muted-foreground">
                  <div className="text-center">
                    <Package className="w-12 h-12 mx-auto mb-2 opacity-50" />
                    <p>输入关键词搜索技能</p>
                  </div>
                </div>
              )}
              
              {searchResults.map((skill) => (
                <Card key={`${skill.registry_name}-${skill.slug}`} className="hover:shadow-md transition-shadow">
                  <CardHeader className="pb-2">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <CardTitle className="text-lg">{skill.display_name}</CardTitle>
                        <div className="flex items-center gap-2 mt-1">
                          <Badge variant="outline">{skill.slug}</Badge>
                          <Badge variant="secondary">{skill.version}</Badge>
                          <Badge variant="outline">{skill.registry_name}</Badge>
                          {skill.score > 0 && (
                            <Badge variant="outline">评分: {skill.score.toFixed(2)}</Badge>
                          )}
                        </div>
                      </div>
                      <Button
                        size="sm"
                        onClick={() => handleInstall(skill)}
                        disabled={isInstalling === skill.slug}
                        className="flex items-center gap-1"
                      >
                        {isInstalling === skill.slug ? (
                          <Loader2 className="w-4 h-4 animate-spin" />
                        ) : (
                          <Download className="w-4 h-4" />
                        )}
                        {isInstalling === skill.slug ? '安装中...' : '安装'}
                      </Button>
                    </div>
                  </CardHeader>
                  <CardContent>
                    <CardDescription>
                      {skill.summary}
                    </CardDescription>
                  </CardContent>
                </Card>
              ))}
            </div>
          </ScrollArea>
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