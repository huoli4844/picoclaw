import { useState, useEffect, useRef } from 'react'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Search, Download, Loader2, Server, Globe, Terminal, Radio, ChevronDown, X, Check } from 'lucide-react'
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
  const [selectedSources, setSelectedSources] = useState<string[]>([])
  const [availableSources, setAvailableSources] = useState<string[]>([])
  const [showSourceDropdown, setShowSourceDropdown] = useState(false)
  const [searchResults, setSearchResults] = useState<McpServer[]>([])
  const [isSearching, setIsSearching] = useState(false)
  const [isInstalling, setIsInstalling] = useState<string | null>(null)
  const [isValidating, setIsValidating] = useState<string | null>(null)
  
  // 分页状态
  const [hasMore, setHasMore] = useState(true)
  const [totalCount, setTotalCount] = useState(0)
  const [isInitialLoad, setIsInitialLoad] = useState(true)
  
  const { searchMcpServers, installMcpServer, validateMcpServer } = useApi()

  // 添加调试状态
  const [debugInfo, setDebugInfo] = useState<string>('')
  
  // Intersection Observer for infinite scroll
  const loadMoreRef = useRef<HTMLDivElement>(null)

  const PAGE_SIZE = 50  // 增加每次加载的数量，减少请求次数

  // Load available MCP sources when dialog opens
  useEffect(() => {
    console.log('MCP搜索对话框状态:', isOpen, '可用源数量:', availableSources.length)
    if (isOpen && availableSources.length === 0) {
      loadAvailableSources()
    }
  }, [isOpen, availableSources.length])

  const loadAvailableSources = async () => {
    try {
      console.log('开始加载MCP源...')
      
      // 直接使用fetch API
      const response = await fetch('http://localhost:8080/api/mcp/sources')
      const data = await response.json()
      console.log('直接API调用结果:', data)
      
      if (data.success && Array.isArray(data.data)) {
        const sources = data.data
        setAvailableSources(sources)
        console.log('MCP源加载成功:', sources.length, '个源:', sources)
      } else {
        console.error('MCP源数据格式错误:', data)
      }
    } catch (error) {
      console.error('加载MCP源失败:', error)
    }
  }

  const handleSourceToggle = (source: string) => {
    setSelectedSources(prev => {
      if (prev.includes(source)) {
        return prev.filter(s => s !== source)
      } else {
        return [...prev, source]
      }
    })
  }

  const handleSourceRemove = (source: string) => {
    setSelectedSources(prev => prev.filter(s => s !== source))
  }

  // Handle clicking outside to close source dropdown
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (showSourceDropdown) {
        const target = event.target as Element
        if (!target.closest('.relative')) {
          setShowSourceDropdown(false)
        }
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [showSourceDropdown])

  // Intersection Observer for infinite scroll
  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        const target = entries[0]
        if (target.isIntersecting && hasMore && !isSearching) {
          console.log('Intersection Observer触发加载更多...')
          handleLoadMore()
        }
      },
      {
        root: null,
        rootMargin: '100px',
        threshold: 0.1
      }
    )

    if (loadMoreRef.current) {
      observer.observe(loadMoreRef.current)
    }

    return () => {
      if (loadMoreRef.current) {
        observer.unobserve(loadMoreRef.current)
      }
    }
  }, [hasMore, isSearching])

  const handleSearch = async (isLoadMore = false) => {
    console.log('开始搜索MCP服务器...', { searchQuery, category, transport, isLoadMore })
    setIsSearching(true)
    
    try {
      const request: McpSearchRequest = {
        query: searchQuery,
        limit: isLoadMore ? 1000 : PAGE_SIZE,  // 加载更多时使用大数值获取所有剩余数据
        offset: isLoadMore ? searchResults.length : 0,
        ...(category !== 'all' && { category }),
        ...(transport !== 'all' && { transport }),
        ...(selectedSources.length > 0 && { sources: selectedSources })
      }
      
      const result = await searchMcpServers(request)
      console.log('Search result:', result)
      
      if (result.success && result.data) {
        console.log('Search data:', result.data)
        
        // 处理双重嵌套：result.data 可能是 {data: {results: [...], success: true}} 或直接是 {results: [...]}
        let searchData: any = result.data
        if (searchData.data) {
          // 双重嵌套情况
          searchData = searchData.data
        }
        console.log('处理后的searchData:', searchData)
        console.log('searchData.results:', searchData.results)
        
        if (searchData && searchData.results) {
          const newResults = searchData.results
          console.log('找到新的搜索结果:', newResults.length)
          console.log('搜索结果内容:', newResults)
          
          if (isLoadMore) {
            // 加载更多：追加到现有结果
            const newTotalLength = searchResults.length + newResults.length
            setSearchResults(prev => [...prev, ...newResults])
            setTotalCount(searchData.total || 0)
            setHasMore(newTotalLength < (searchData.total || 0))
            setDebugInfo(`共找到 ${searchData.total || 0} 个结果，当前显示 ${newTotalLength} 个`)
          } else {
            // 新搜索：替换所有结果
            setSearchResults(newResults)
            setTotalCount(searchData.total || 0)
            setHasMore(newResults.length < (searchData.total || 0))
            setDebugInfo(`共找到 ${searchData.total || 0} 个结果，当前显示 ${newResults.length} 个`)
          }
        } else {
          console.log('意外的数据结构:', JSON.stringify(searchData, null, 2))
          setDebugInfo(`意外的数据结构: ${JSON.stringify(searchData, null, 2)}`)
          if (!isLoadMore) {
            setSearchResults([])
          }
        }
      } else {
        console.warn('Search API response unsuccessful:', result)
        if (!isLoadMore) {
          setSearchResults([])
        }
      }
    } catch (error) {
      console.error('Failed to search MCP servers:', error)
      if (!isLoadMore) {
        setSearchResults([])
      }
    } finally {
      setIsSearching(false)
      setIsInitialLoad(false)
    }
  }

  // 初始搜索
  const handleInitialSearch = () => {
    handleSearch(false)
  }

  // 加载更多
  const handleLoadMore = () => {
    console.log('点击加载更多...', { 
      currentResults: searchResults.length, 
      hasMore, 
      isSearching, 
      totalCount 
    })
    if (!isSearching && hasMore) {
      handleSearch(true)
    }
  }

  const handleInstall = async (server: McpServer) => {
    console.log('开始安装MCP服务器:', server.id, server)
    setIsInstalling(server.id)
    setDebugInfo(`正在安装 ${server.name}...`)
    
    try {
      const request: McpInstallRequest = {
        serverId: server.id,
        config: server.config
      }
      console.log('安装请求:', request)
      const result = await installMcpServer(request)
      console.log('Install result:', result)
      
      if (result.success) {
        setDebugInfo(`${server.name} 安装成功!`)
        // Give the server a moment to register before validation
        setTimeout(async () => {
          setIsValidating(server.id)
          try {
            const validationResult = await validateMcpServer(server.id)
            console.log('Validation result:', validationResult)
            
            if (validationResult.success) {
              setDebugInfo(`${server.name} 安装并验证成功!`)
            } else {
              setDebugInfo(`${server.name} 安装成功但验证失败: ${validationResult.error || '未知错误'}`)
              console.error('Validation failed:', validationResult)
            }
          } catch (error) {
            console.error('Validation error:', error)
            setDebugInfo(`${server.name} 安装成功但验证时出错: ${error}`)
          } finally {
            setIsValidating(null)
            onServerInstalled()
            // Remove from search results after installation
            setSearchResults(prev => (prev || []).filter(s => s.id !== server.id))
          }
        }, 1000) // Wait 1 second before validation
      } else {
        setDebugInfo(`${server.name} 安装失败: ${result.error || '未知错误'}`)
        console.error('Install failed:', result)
      }
    } catch (error) {
      setDebugInfo(`${server.name} 安装出错: ${error}`)
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
                onClick={handleInitialSearch} 
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
              <Button 
                onClick={() => {
                  console.log('手动测试MCP源API')
                  fetch('http://localhost:8080/api/mcp/sources')
                    .then(response => response.json())
                    .then(data => console.log('直接API调用结果:', data))
                    .catch(error => console.error('直接API调用错误:', error))
                }}
                variant="outline"
                className="flex items-center gap-2"
              >
                测试API
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

              {/* MCP Sources Multi-Select Dropdown */}
              <div className="relative">
                <div 
                  className="flex items-center gap-2 border rounded-md px-3 py-2 min-w-48 cursor-pointer bg-white hover:bg-gray-50"
                  onClick={() => {
                    console.log('MCP源选择器被点击，当前状态:', showSourceDropdown)
                    setShowSourceDropdown(!showSourceDropdown)
                  }}
                >
                  <div className="flex-1 text-sm text-gray-600">
                    {selectedSources.length === 0 
                      ? "选择MCP源" 
                      : `已选择 ${selectedSources.length} 个源`
                    }
                  </div>
                  <ChevronDown className="w-4 h-4 text-gray-500" />
                </div>
                
                {showSourceDropdown && (
                  <div className="absolute top-full left-0 right-0 mt-1 bg-white border rounded-md shadow-lg z-[9999] max-h-48 overflow-y-auto">
                    <div className="p-2">
                      {availableSources.length > 0 ? (
                        availableSources.map((source) => (
                        <div 
                          key={source} 
                          className="flex items-center space-x-2 p-2 hover:bg-gray-50 rounded cursor-pointer"
                          onClick={() => handleSourceToggle(source)}
                        >
                          <div className={`w-4 h-4 border rounded flex items-center justify-center ${
                            selectedSources.includes(source) 
                              ? 'bg-blue-500 border-blue-500' 
                              : 'border-gray-300'
                          }`}>
                            {selectedSources.includes(source) && (
                              <Check className="w-3 h-3 text-white" />
                            )}
                          </div>
                          <span className="text-sm flex-1">
                            {source}
                          </span>
                        </div>
                      ))
                      ) : (
                        <div className="p-2 text-sm text-gray-500">加载中...</div>
                      )}
                    </div>
                  </div>
                )}
              </div>

              {/* Selected Sources Badges */}
              {selectedSources.length > 0 && (
                <div className="flex items-center gap-1 flex-wrap">
                  {selectedSources.map((source) => (
                    <Badge 
                      key={source} 
                      variant="secondary" 
                      className="flex items-center gap-1 cursor-pointer"
                      onClick={() => handleSourceRemove(source)}
                    >
                      {source}
                      <X className="w-3 h-3" />
                    </Badge>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* Search Results */}
          <div 
            className="flex-1 overflow-y-auto"
            onScroll={(e) => {
              const { scrollTop, scrollHeight, clientHeight } = e.currentTarget
              // 当滚动到底部90%时，自动加载更多
              const threshold = (scrollHeight - clientHeight) * 0.9 + clientHeight
              console.log('滚动检测:', { 
                scrollTop, 
                scrollHeight, 
                clientHeight, 
                threshold,
                hasMore, 
                isSearching,
                shouldLoad: scrollTop + clientHeight >= threshold
              })
              
              if (scrollTop + clientHeight >= threshold && hasMore && !isSearching) {
                console.log('触发自动加载更多...')
                handleLoadMore()
              }
            }}
          >
            <div className="space-y-3">
              {(() => {
                console.log('渲染阶段 - searchResults:', searchResults)
                console.log('渲染阶段 - isSearching:', isSearching)
                return null
              })()}
              
              {isInitialLoad && !searchResults.length && !isSearching && (
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
          </div>
        </div>

          <div className="flex justify-between items-center mt-4">
            <div className="flex flex-col gap-2">
              <p className="text-sm text-muted-foreground">
                {totalCount > 0 
                  ? `显示 ${searchResults.length} / ${totalCount} 个结果`
                  : `搜索到 ${searchResults.length} 个结果`
                }
              </p>
              {debugInfo && (
                <details className="text-xs text-muted-foreground">
                  <summary>调试信息</summary>
                  <pre className="whitespace-pre-wrap break-all">{debugInfo}</pre>
                </details>
              )}
            </div>
            <div className="flex gap-2">
              {/* 加载更多按钮 */}
              {hasMore && !isSearching && searchResults.length > 0 && (
                <Button
                  variant="outline"
                  onClick={handleLoadMore}
                  disabled={isSearching}
                  className="flex items-center gap-2"
                >
                  {isSearching ? (
                    <>
                      <Loader2 className="w-4 h-4 animate-spin" />
                      加载中...
                    </>
                  ) : (
                    <>
                      <Search className="w-4 h-4" />
                      加载更多
                    </>
                  )}
                </Button>
              )}
              
              {/* 没有更多数据提示 */}
              {!hasMore && searchResults.length > 0 && (
                <p className="text-sm text-muted-foreground italic">
                  已显示所有结果
                </p>
              )}
              
              <Button variant="outline" onClick={onClose}>
                关闭
              </Button>
            </div>
          </div>
      </DialogContent>
    </Dialog>
  )
}