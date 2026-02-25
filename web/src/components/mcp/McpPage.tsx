import { useState, useEffect } from 'react'
import { Search, Server, Trash2, Loader2, Globe, Terminal, Radio, AlertCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger } from '@/components/ui/alert-dialog'
import { useApi } from '@/hooks/useApi'
import { McpServer } from '@/types'
import { McpSearch } from './McpSearch'
import { McpToolTester } from './McpToolTester'

interface McpPageProps {
  onBack: () => void
}

export function McpPage({ onBack }: McpPageProps) {
  const [servers, setServers] = useState<McpServer[]>([])
  const [searchQuery, setSearchQuery] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [isSearchOpen, setIsSearchOpen] = useState(false)
  const [selectedServer, setSelectedServer] = useState<McpServer | null>(null)
  const [isDetailOpen, setIsDetailOpen] = useState(false)
  const [isUninstalling, setIsUninstalling] = useState<string | null>(null)
  const { getMcpServers, uninstallMcpServer } = useApi()

  // 为了调试，在window上暴露一个函数来打开搜索对话框
  useEffect(() => {
    (window as any).openMcpSearch = () => {
      console.log('打开MCP搜索对话框')
      setIsSearchOpen(true)
    }
    return () => {
      delete (window as any).openMcpSearch
    }
  }, [])

  const loadServers = async () => {
    setIsLoading(true)
    try {
      const result = await getMcpServers()
      console.log('MCP servers result:', result)
      if (result.success && result.data) {
        console.log('MCP servers data:', result.data)
        
        // 检查数据结构：可能是嵌套的 {data: {data: [...], success: true}}
        let serversData = result.data
        if (serversData && 'data' in serversData && Array.isArray((serversData as any).data)) {
          serversData = (serversData as any).data
        }
        
        console.log('Final servers data:', serversData)
        console.log('Is array?', Array.isArray(serversData))
        
        // 确保数据是数组
        if (Array.isArray(serversData)) {
          setServers(serversData)
        } else {
          console.warn('MCP servers data is not an array:', typeof serversData)
          setServers([])
        }
      } else {
        console.warn('MCP servers API response unsuccessful:', result)
        setServers([])
      }
    } catch (error) {
      console.error('Failed to load MCP servers:', error)
      setServers([])
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    loadServers()
  }, [])

  const filteredServers = (servers || []).filter(server =>
    server.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    server.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
    server.id.toLowerCase().includes(searchQuery.toLowerCase())
  )

  const handleServerClick = (server: McpServer) => {
    setSelectedServer(server)
    setIsDetailOpen(true)
  }

  const handleServerInstalled = () => {
    loadServers()
    setIsSearchOpen(false)
  }

  const handleUninstall = async (serverId: string) => {
    setIsUninstalling(serverId)
    try {
      const result = await uninstallMcpServer(serverId)
      if (result.success) {
        loadServers()
        if (selectedServer?.id === serverId) {
          setIsDetailOpen(false)
          setSelectedServer(null)
        }
      }
    } catch (error) {
      console.error('Failed to uninstall MCP server:', error)
    } finally {
      setIsUninstalling(null)
    }
  }

  const getTransportIcon = (transport: string) => {
    switch (transport) {
      case 'stdio':
        return <Terminal className="w-4 h-4" />
      case 'sse':
        return <Globe className="w-4 h-4" />
      case 'websocket':
        return <Radio className="w-4 h-4" />
      default:
        return <Server className="w-4 h-4" />
    }
  }

  const getStatusBadgeVariant = (status: string) => {
    switch (status) {
      case 'installed':
        return 'default'
      case 'available':
        return 'secondary'
      case 'error':
        return 'destructive'
      default:
        return 'outline'
    }
  }

  const getStatusText = (status: string) => {
    switch (status) {
      case 'installed':
        return '已安装'
      case 'available':
        return '可用'
      case 'error':
        return '错误'
      default:
        return status
    }
  }

  return (
    <div className="h-full flex flex-col">
      {/* Search Bar */}
      <div className="border-b p-4">
        <div className="max-w-6xl mx-auto">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Server className="w-6 h-6 text-primary" />
              <div>
                <h2 className="text-lg font-semibold">MCP 服务器管理</h2>
                <p className="text-sm text-muted-foreground">管理 Model Context Protocol 服务器</p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <div className="relative mr-2">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground w-4 h-4" />
                <Input
                  placeholder="搜索 MCP 服务器..."
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
                <Search className="w-4 h-4" />
                搜索服务器
              </Button>
            </div>
          </div>
        </div>
      </div>

      {/* Servers List */}
      <ScrollArea className="flex-1">
        <div className="max-w-6xl mx-auto p-4">
          {isLoading ? (
            <div className="flex items-center justify-center h-64">
              <div className="flex items-center gap-2 text-muted-foreground">
                <Loader2 className="w-6 h-6 animate-spin" />
                <span>加载服务器列表...</span>
              </div>
            </div>
          ) : !servers || !Array.isArray(servers) ? (
            <div className="flex items-center justify-center h-64">
              <div className="text-center text-muted-foreground">
                <Server className="w-12 h-12 mx-auto mb-4 opacity-50" />
                <p>无法加载 MCP 服务器列表</p>
                <p className="text-sm mt-2">请检查后端服务是否正常运行</p>
              </div>
            </div>
          ) : filteredServers.length === 0 ? (
            <div className="flex items-center justify-center h-64">
              <div className="text-center text-muted-foreground">
                <Server className="w-12 h-12 mx-auto mb-4 opacity-50" />
                <p>暂无 MCP 服务器</p>
                <p className="text-sm mt-2">点击"安装服务器"来添加新的 MCP 服务器</p>
              </div>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {filteredServers.map((server) => (
                <Card 
                  key={server.id} 
                  className="cursor-pointer hover:shadow-md transition-shadow"
                  onClick={() => handleServerClick(server)}
                >
                  <CardHeader className="pb-2">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-1">
                          <CardTitle className="text-lg">{server.name}</CardTitle>
                          {getTransportIcon(server.transport)}
                        </div>
                        <div className="flex items-center gap-2 mt-1 flex-wrap">
                          <Badge variant={getStatusBadgeVariant(server.status)}>
                            {getStatusText(server.status)}
                          </Badge>
                          <Badge variant="outline">{server.version}</Badge>
                          <Badge variant="outline">{server.transport}</Badge>
                          {server.category && (
                            <Badge variant="outline">{server.category}</Badge>
                          )}
                        </div>
                        {server.author && (
                          <p className="text-xs text-muted-foreground mt-1">
                            {server.author}
                          </p>
                        )}
                      </div>
                      {server.status === 'installed' && (
                        <AlertDialog>
                          <AlertDialogTrigger asChild>
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={(e) => {
                                e.stopPropagation()
                              }}
                            >
                              {isUninstalling === server.id ? (
                                <Loader2 className="w-4 h-4 animate-spin" />
                              ) : (
                                <Trash2 className="w-4 h-4" />
                              )}
                            </Button>
                          </AlertDialogTrigger>
                          <AlertDialogContent>
                            <AlertDialogHeader>
                              <AlertDialogTitle>确认卸载</AlertDialogTitle>
                              <AlertDialogDescription>
                                确定要卸载 MCP 服务器 "{server.name}" 吗？此操作不可撤销。
                              </AlertDialogDescription>
                            </AlertDialogHeader>
                            <AlertDialogFooter>
                              <AlertDialogCancel>取消</AlertDialogCancel>
                              <AlertDialogAction 
                                onClick={() => handleUninstall(server.id)}
                                disabled={isUninstalling === server.id}
                              >
                                {isUninstalling === server.id ? '卸载中...' : '确认卸载'}
                              </AlertDialogAction>
                            </AlertDialogFooter>
                          </AlertDialogContent>
                        </AlertDialog>
                      )}
                    </div>
                  </CardHeader>
                  <CardContent>
                    <CardDescription className="line-clamp-3">
                      {server.description}
                    </CardDescription>
                    {server.tools && server.tools.length > 0 && (
                      <div className="mt-2">
                        <p className="text-xs text-muted-foreground">
                          提供 {server.tools.length} 个工具
                        </p>
                      </div>
                    )}
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </div>
      </ScrollArea>

      {/* MCP Search Dialog */}
      <McpSearch
        isOpen={isSearchOpen}
        onClose={() => setIsSearchOpen(false)}
        onServerInstalled={handleServerInstalled}
      />

      {/* Server Detail Dialog with Tool Testing */}
      {selectedServer && (
        <AlertDialog open={isDetailOpen} onOpenChange={setIsDetailOpen}>
          <AlertDialogContent className="max-w-4xl max-h-[80vh] overflow-hidden">
            <AlertDialogHeader>
              <AlertDialogTitle className="flex items-center gap-2">
                {getTransportIcon(selectedServer.transport)}
                {selectedServer.name}
              </AlertDialogTitle>
            </AlertDialogHeader>
            
            <div className="flex flex-col h-[60vh]">
              {/* Server Info Section */}
              <div className="flex-shrink-0 space-y-2 text-sm text-left mb-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <strong>ID:</strong> {selectedServer.id}
                  </div>
                  <div>
                    <strong>版本:</strong> {selectedServer.version}
                  </div>
                  <div>
                    <strong>传输方式:</strong> {selectedServer.transport}
                  </div>
                  {selectedServer.author && (
                    <div>
                      <strong>作者:</strong> {selectedServer.author}
                    </div>
                  )}
                  {selectedServer.category && (
                    <div>
                      <strong>类别:</strong> {selectedServer.category}
                    </div>
                  )}
                  {selectedServer.license && (
                    <div>
                      <strong>许可证:</strong> {selectedServer.license}
                    </div>
                  )}
                </div>
                <div>
                  <strong>描述:</strong><br />
                  {selectedServer.description}
                </div>
                {selectedServer.keywords && selectedServer.keywords.length > 0 && (
                  <div>
                    <strong>关键词:</strong><br />
                    <div className="flex flex-wrap gap-1 mt-1">
                      {selectedServer.keywords.map((keyword, index) => (
                        <Badge key={index} variant="outline" className="text-xs">
                          {keyword}
                        </Badge>
                      ))}
                    </div>
                  </div>
                )}
                {selectedServer.homepage && (
                  <div>
                    <strong>主页:</strong><br />
                    <a 
                      href={selectedServer.homepage} 
                      target="_blank" 
                      rel="noopener noreferrer" 
                      className="text-blue-500 hover:text-blue-700 hover:underline break-all"
                    >
                      {selectedServer.homepage}
                    </a>
                  </div>
                )}
                {selectedServer.repository && (
                  <div>
                    <strong>代码仓库:</strong><br />
                    <a 
                      href={selectedServer.repository} 
                      target="_blank" 
                      rel="noopener noreferrer" 
                      className="text-blue-500 hover:text-blue-700 hover:underline break-all"
                    >
                      {selectedServer.repository}
                    </a>
                  </div>
                )}
              </div>

              {/* Tool Testing Section */}
              <div className="flex-1 min-h-0">
                <ScrollArea className="h-full">
                  {selectedServer.tools && selectedServer.tools.length > 0 ? (
                    <McpToolTester
                      serverId={selectedServer.id}
                      tools={selectedServer.tools}
                    />
                  ) : (
                    <div className="text-center py-8 text-muted-foreground">
                      <AlertCircle className="w-8 h-8 mx-auto mb-2" />
                      <p>该 MCP 服务器没有提供可用工具</p>
                    </div>
                  )}
                </ScrollArea>
              </div>
            </div>

            <AlertDialogFooter>
              <AlertDialogAction onClick={() => setIsDetailOpen(false)}>
                关闭
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      )}
    </div>
  )
}