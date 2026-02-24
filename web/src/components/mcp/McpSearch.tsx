import { useState } from 'react'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Search, Download, Loader2, Server, Globe, Terminal, Radio } from 'lucide-react'
import { useApi } from '@/hooks/useApi'
import { McpSearchRequest, McpServer, McpInstallRequest } from '@/types'

interface McpSearchProps {
  isOpen: boolean
  onClose: () => void
  onServerInstalled: () => void
}

export function McpSearch({ isOpen, onClose, onServerInstalled }: McpSearchProps) {
  const [searchQuery, setSearchQuery] = useState('')
  const [category, setCategory] = useState<string>('all')
  const [transport, setTransport] = useState<string>('all')
  const [searchResults, setSearchResults] = useState<McpServer[]>([])
  const [isSearching, setIsSearching] = useState(false)
  const [isInstalling, setIsInstalling] = useState<string | null>(null)
  const [isValidating, setIsValidating] = useState<string | null>(null)
  const { searchMcpServers, installMcpServer, validateMcpServer } = useApi()

  // 添加调试状态
  const [debugInfo, setDebugInfo] = useState<string>('')

  const handleSearch = async () => {
    // 允许空搜索，这样可以显示所有可用的MCP服务器
    console.log('开始搜索MCP服务器...', { searchQuery, category, transport })
    setIsSearching(true)
    try {
      const request: McpSearchRequest = {
        query: searchQuery,
        limit: 20,
        ...(category !== 'all' && { category }),
        ...(transport !== 'all' && { transport })
      }
      const result = await searchMcpServers(request)
      console.log('Search result:', result)
      if (result.success && result.data) {
        console.log('Search data:', result.data)
        
        // API返回的数据结构: {success: true, data: {data: {query, results: [...]}}}
        let searchData = result.data as any
        
        // 检查是否有嵌套的data结构
        if (searchData && searchData.data && searchData.data.results) {
          const results = searchData.data.results
          console.log('找到嵌套数据结构! 即将设置searchResults:', results)
          setSearchResults(results)
          setDebugInfo(`找到 ${results.length} 个结果 (使用嵌套data结构)`)
          console.log('searchResults设置完成')
        } else if (searchData && searchData.results) {
          // 兼容直接的数据结构
          const results = searchData.results
          console.log('找到直接数据结构! 即将设置searchResults:', results)
          setSearchResults(results)
          setDebugInfo(`找到 ${results.length} 个结果 (使用直接data结构)`)
          console.log('searchResults设置完成')
        } else {
          console.log('条件失败! 进入else分支')
          console.log('完整数据结构:', JSON.stringify(searchData, null, 2))
          setDebugInfo(`意外的数据结构: ${JSON.stringify(searchData, null, 2)}`)
          setSearchResults([])
        }
      } else {
        console.warn('Search API response unsuccessful:', result)
        setSearchResults([])
      }
    } catch (error) {
      console.error('Failed to search MCP servers:', error)
      setSearchResults([])
    } finally {
      // 添加延迟确保状态更新完成
      setTimeout(() => {
        setIsSearching(false)
      }, 100)
    }
  }

  const handleInstall = async (server: McpServer) => {
    setIsInstalling(server.id)
    try {
      const request: McpInstallRequest = {
        serverId: server.id,
        config: server.config
      }
      const result = await installMcpServer(request)
      console.log('Install result:', result)
      if (result.success) {
        // Validate installation after successful install
        setIsValidating(server.id)
        const validationResult = await validateMcpServer(server.id)
        console.log('Validation result:', validationResult)
        
        if (validationResult.success) {
          onServerInstalled()
          // Remove from search results after successful installation and validation
          setSearchResults(prev => (prev || []).filter(s => s.id !== server.id))
        } else {
          console.error('Validation failed:', validationResult)
          // Still remove from search results, but the validation error will be shown
          onServerInstalled()
        }
      } else {
        console.error('Install failed:', result)
      }
    } catch (error) {
      console.error('Failed to install MCP server:', error)
    } finally {
      setIsInstalling(null)
      setIsValidating(null)
    }
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSearch()
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

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-5xl max-h-[85vh]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Search className="w-5 h-5" />
            搜索和安装 MCP 服务器
          </DialogTitle>
          <p className="text-sm text-muted-foreground">
            在官方MCP仓库中搜索可用的服务器，支持按类别和传输方式筛选
          </p>
        </DialogHeader>
        
        <div className="flex flex-col h-[65vh]">
          {/* Search and Filter Controls */}
          <div className="space-y-4 mb-4">
            <div className="flex gap-2">
              <Input
                placeholder="搜索 MCP 服务器 (例如: file system, database, git)..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onKeyPress={handleKeyPress}
                className="flex-1"
              />
              <Button 
                onClick={handleSearch} 
                disabled={isSearching}
                className="flex items-center gap-2"
              >
                {isSearching ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  <Search className="w-4 h-4" />
                )}
                搜索
              </Button>
              <Button 
                onClick={() => {
                  console.log('测试数据设置 - 开始')
                  const testData = [{
                    id: "filesystem",
                    name: "Filesystem MCP Server",
                    description: "提供文件系统操作工具，包括文件读写、目录管理、文件搜索等功能",
                    version: "1.0.0",
                    author: "Model Context Protocol",
                    homepage: "https://github.com/modelcontextprotocol/servers",
                    repository: "https://github.com/modelcontextprotocol/servers",
                    license: "MIT",
                    keywords: ["filesystem", "files", "directory", "io"],
                    category: "filesystem",
                    transport: "stdio" as const,
                    command: "mcp-server-filesystem",
                    args: ["/Users/huoli4844/Documents/ai_project/picoclaw"],
                    status: "available" as const,
                    tools: [
                      {
                        name: "read_file",
                        description: "读取文件内容",
                        inputSchema: {
                          type: "object",
                          properties: {
                            path: {
                              type: "string",
                              description: "文件路径"
                            }
                          },
                          required: ["path"]
                        },
                        serverId: "filesystem"
                      }
                    ]
                  }]
                  console.log('设置前 - searchResults:', searchResults)
                  setSearchResults(testData)
                  setDebugInfo(`测试数据设置成功 - ${testData.length} 个结果`)
                  console.log('设置后 - 应该有数据了')
                  // 强制重新渲染
                  setTimeout(() => {
                    console.log('延迟检查 - searchResults:', searchResults)
                  }, 100)
                }}
                variant="outline"
                className="flex items-center gap-2"
              >
                测试数据
              </Button>
            </div>
            
            {/* Filters */}
            <div className="flex gap-2">
              <Select value={category} onValueChange={setCategory}>
                <SelectTrigger className="w-40">
                  <SelectValue placeholder="选择类别" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">所有类别</SelectItem>
                  <SelectItem value="filesystem">文件系统</SelectItem>
                  <SelectItem value="database">数据库</SelectItem>
                  <SelectItem value="development">开发工具</SelectItem>
                  <SelectItem value="communication">通信</SelectItem>
                  <SelectItem value="productivity">生产力</SelectItem>
                  <SelectItem value="ai">AI 工具</SelectItem>
                </SelectContent>
              </Select>
              
              <Select value={transport} onValueChange={setTransport}>
                <SelectTrigger className="w-40">
                  <SelectValue placeholder="传输方式" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">所有传输</SelectItem>
                  <SelectItem value="stdio">STDIO</SelectItem>
                  <SelectItem value="sse">SSE</SelectItem>
                  <SelectItem value="websocket">WebSocket</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* Search Results */}
          <ScrollArea className="flex-1">
            <div className="space-y-3">
              {(() => {
                console.log('渲染阶段 - searchResults:', searchResults)
                console.log('渲染阶段 - isSearching:', isSearching)
                return null
              })()}
              
              {(!searchResults || searchResults.length === 0) && !isSearching && (
                <div className="flex items-center justify-center h-32 text-muted-foreground">
                  <div className="text-center">
                    <Server className="w-12 h-12 mx-auto mb-2 opacity-50" />
                    <p>输入关键词或选择筛选条件搜索 MCP 服务器</p>
                  </div>
                </div>
              )}
              
              {searchResults.map((server) => (
                <Card key={server.id} className="hover:shadow-md transition-shadow">
                  <CardHeader className="pb-2">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-1">
                          <CardTitle className="text-lg">{server.name}</CardTitle>
                          {getTransportIcon(server.transport)}
                        </div>
                        <div className="flex items-center gap-2 mt-1 flex-wrap">
                          <Badge variant="outline">{server.id}</Badge>
                          <Badge variant="secondary">{server.version}</Badge>
                          <Badge variant={getStatusBadgeVariant(server.status)}>
                            {server.status}
                          </Badge>
                          <Badge variant="outline">{server.transport}</Badge>
                          {server.category && (
                            <Badge variant="outline">{server.category}</Badge>
                          )}
                          {server.tools && (
                            <Badge variant="outline">工具: {server.tools.length}</Badge>
                          )}
                        </div>
                        {server.author && (
                          <p className="text-sm text-muted-foreground mt-1">
                            作者: {server.author}
                          </p>
                        )}
                      </div>
                      <Button
                        size="sm"
                        onClick={() => handleInstall(server)}
                        disabled={isInstalling === server.id || isValidating === server.id || server.status === 'installed'}
                        className="flex items-center gap-1"
                      >
                        {(isInstalling === server.id || isValidating === server.id) ? (
                          <Loader2 className="w-4 h-4 animate-spin" />
                        ) : (
                          <Download className="w-4 h-4" />
                        )}
                        {isInstalling === server.id ? '安装中...' : 
                         isValidating === server.id ? '验证中...' :
                         server.status === 'installed' ? '已安装' : '安装'}
                      </Button>
                    </div>
                  </CardHeader>
                  <CardContent>
                    <CardDescription className="mb-2">
                      {server.description}
                    </CardDescription>
                    {server.keywords && server.keywords.length > 0 && (
                      <div className="flex flex-wrap gap-1 mt-2">
                        {server.keywords.map((keyword, index) => (
                          <Badge key={index} variant="outline" className="text-xs">
                            {keyword}
                          </Badge>
                        ))}
                      </div>
                    )}
                    {server.repository && (
                      <p className="text-xs text-muted-foreground mt-2">
                        仓库: <a href={server.repository} target="_blank" rel="noopener noreferrer" className="text-blue-500 hover:underline">{server.repository}</a>
                      </p>
                    )}
                  </CardContent>
                </Card>
              ))}
            </div>
          </ScrollArea>
        </div>

          <div className="flex justify-between items-center mt-4">
            <div className="flex flex-col gap-2">
              <p className="text-sm text-muted-foreground">
                搜索到 {(searchResults || []).length} 个结果
              </p>
              {debugInfo && (
                <details className="text-xs text-muted-foreground">
                  <summary>调试信息</summary>
                  <pre className="whitespace-pre-wrap break-all">{debugInfo}</pre>
                </details>
              )}
            </div>
          <div className="flex gap-2">
            <Button variant="outline" onClick={onClose}>
              关闭
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}