import { useState } from 'react'
import { Play, Copy, CheckCircle, AlertCircle, Loader2, Code } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useApi } from '@/hooks/useApi'
import { McpTool } from '@/types'

interface McpToolTesterProps {
  serverId: string
  tools: McpTool[]
}

interface ToolCallResult {
  success: boolean
  serverID: string
  toolName: string
  arguments: any
  result: string
  timestamp: string
  isSimulation: boolean
}

export function McpToolTester({ serverId, tools }: McpToolTesterProps) {
  const [selectedTool, setSelectedTool] = useState<McpTool | null>(null)
  const [toolArguments, setToolArguments] = useState<Record<string, any>>({})
  const [result, setResult] = useState<ToolCallResult | null>(null)
  const [isCalling, setIsCalling] = useState(false)
  const [copied, setCopied] = useState(false)
  const { callMcpTool } = useApi()

  const handleToolSelect = (tool: McpTool) => {
    setSelectedTool(tool)
    setResult(null)
    
    // 根据工具的输入schema初始化参数
    const initialArgs: Record<string, any> = {}
    if (tool.inputSchema && tool.inputSchema.properties) {
      Object.keys(tool.inputSchema.properties).forEach(key => {
        const prop = tool.inputSchema!.properties![key]
        initialArgs[key] = prop.default || ''
      })
    }
    setToolArguments(initialArgs)
  }

  const handleArgumentChange = (key: string, value: any) => {
    setToolArguments(prev => ({
      ...prev,
      [key]: value
    }))
  }

  const handleCallTool = async () => {
    if (!selectedTool) return

    setIsCalling(true)
    try {
      console.log('调用MCP工具:', serverId, selectedTool.name, toolArguments)
      const response = await callMcpTool(serverId, selectedTool.name, toolArguments)
      console.log('MCP工具调用响应:', response)
      
      if (response.success && response.data) {
        setResult(response.data)
      } else {
        console.error('MCP工具调用失败:', response)
        setResult({
          success: false,
          serverID: serverId,
          toolName: selectedTool.name,
          arguments: toolArguments,
          result: response.error || '调用失败',
          timestamp: new Date().toISOString(),
          isSimulation: false
        })
      }
    } catch (error) {
      console.error('MCP工具调用异常:', error)
      setResult({
        success: false,
        serverID: serverId,
        toolName: selectedTool.name,
        arguments: toolArguments,
        result: error instanceof Error ? error.message : '未知错误',
        timestamp: new Date().toISOString(),
        isSimulation: false
      })
    } finally {
      setIsCalling(false)
    }
  }

  const handleCopyResult = () => {
    if (result) {
      navigator.clipboard.writeText(JSON.stringify(result, null, 2))
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  const renderArgumentInput = (key: string, schema: any) => {
    const value = toolArguments[key] || ''
    
    if (schema.type === 'string') {
      if (schema.description?.toLowerCase().includes('path') || 
          key.toLowerCase().includes('path') ||
          key.toLowerCase().includes('file')) {
        return (
          <Input
            value={value}
            onChange={(e) => handleArgumentChange(key, e.target.value)}
            placeholder={`输入${schema.description || key}`}
            className="font-mono"
          />
        )
      } else if (schema.description?.toLowerCase().includes('content') ||
                 key.toLowerCase().includes('content')) {
        return (
          <Textarea
            value={value}
            onChange={(e) => handleArgumentChange(key, e.target.value)}
            placeholder={`输入${schema.description || key}`}
            rows={4}
            className="font-mono"
          />
        )
      } else {
        return (
          <Input
            value={value}
            onChange={(e) => handleArgumentChange(key, e.target.value)}
            placeholder={`输入${schema.description || key}`}
          />
        )
      }
    } else if (schema.type === 'number' || schema.type === 'integer') {
      return (
        <Input
          type="number"
          value={value}
          onChange={(e) => handleArgumentChange(key, Number(e.target.value))}
          placeholder={`输入${schema.description || key}`}
        />
      )
    } else if (schema.type === 'boolean') {
      return (
        <select
          value={value.toString()}
          onChange={(e) => handleArgumentChange(key, e.target.value === 'true')}
          className="w-full p-2 border rounded"
        >
          <option value="false">false</option>
          <option value="true">true</option>
        </select>
      )
    } else {
      return (
        <Textarea
          value={typeof value === 'string' ? value : JSON.stringify(value, null, 2)}
          onChange={(e) => {
            try {
              const parsed = JSON.parse(e.target.value)
              handleArgumentChange(key, parsed)
            } catch {
              handleArgumentChange(key, e.target.value)
            }
          }}
          placeholder={`输入JSON格式的${schema.description || key}`}
          rows={3}
          className="font-mono"
        />
      )
    }
  }

  if (tools.length === 0) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        <AlertCircle className="w-8 h-8 mx-auto mb-2" />
        <p>该 MCP 服务器没有提供可用工具</p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* 工具选择 */}
      <div>
        <h3 className="text-lg font-semibold mb-3">选择工具</h3>
        <div className="grid grid-cols-1 gap-2">
          {tools.map((tool) => (
            <Card
              key={tool.name}
              className={`cursor-pointer transition-colors ${
                selectedTool?.name === tool.name 
                  ? 'ring-2 ring-primary bg-primary/5' 
                  : 'hover:bg-muted/50'
              }`}
              onClick={() => handleToolSelect(tool)}
            >
              <CardHeader className="pb-2">
                <CardTitle className="text-base flex items-center gap-2">
                  <Code className="w-4 h-4" />
                  {tool.name}
                  {tool.category && (
                    <Badge variant="outline" className="text-xs">
                      {tool.category}
                    </Badge>
                  )}
                </CardTitle>
                <CardDescription className="text-sm">
                  {tool.description}
                </CardDescription>
              </CardHeader>
            </Card>
          ))}
        </div>
      </div>

      {selectedTool && (
        <>
          {/* 参数输入 */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Code className="w-5 h-5" />
                {selectedTool.name} 参数配置
              </CardTitle>
              <CardDescription>
                配置工具调用所需的参数
              </CardDescription>
            </CardHeader>
            <CardContent>
              {selectedTool.inputSchema?.properties ? (
                <div className="space-y-4">
                  {Object.entries(selectedTool.inputSchema.properties).map(([key, schema]: [string, any]) => (
                    <div key={key} className="space-y-2">
                      <Label htmlFor={key} className="flex items-center gap-2">
                        {key}
                        {selectedTool.inputSchema?.required?.includes(key) && (
                          <Badge variant="destructive" className="text-xs px-1 py-0">
                            必需
                          </Badge>
                        )}
                        {schema.type && (
                          <Badge variant="outline" className="text-xs">
                            {schema.type}
                          </Badge>
                        )}
                      </Label>
                      {renderArgumentInput(key, schema)}
                      {schema.description && (
                        <p className="text-xs text-muted-foreground">
                          {schema.description}
                        </p>
                      )}
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-muted-foreground">该工具不需要参数</p>
              )}

              <div className="flex items-center gap-2 pt-4">
                <Button
                  onClick={handleCallTool}
                  disabled={isCalling}
                  className="flex items-center gap-2"
                >
                  {isCalling ? (
                    <>
                      <Loader2 className="w-4 h-4 animate-spin" />
                      调用中...
                    </>
                  ) : (
                    <>
                      <Play className="w-4 h-4" />
                      调用工具
                    </>
                  )}
                </Button>
                <Button variant="outline" onClick={() => setResult(null)}>
                  清除结果
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* 调用结果 */}
          {result && (
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <CardTitle className="flex items-center gap-2">
                      {result.success ? (
                        <CheckCircle className="w-5 h-5 text-green-500" />
                      ) : (
                        <AlertCircle className="w-5 h-5 text-red-500" />
                      )}
                      调用结果
                    </CardTitle>
                    {result.isSimulation && (
                      <Badge variant="secondary" className="text-xs">
                        模拟结果
                      </Badge>
                    )}
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={handleCopyResult}
                    className="flex items-center gap-2"
                  >
                    {copied ? (
                      <>
                        <CheckCircle className="w-4 h-4" />
                        已复制
                      </>
                    ) : (
                      <>
                        <Copy className="w-4 h-4" />
                        复制
                      </>
                    )}
                  </Button>
                </div>
                <CardDescription>
                  {new Date(result.timestamp).toLocaleString()}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <Tabs defaultValue="result" className="w-full">
                  <TabsList className="grid w-full grid-cols-3">
                    <TabsTrigger value="result">结果</TabsTrigger>
                    <TabsTrigger value="arguments">参数</TabsTrigger>
                    <TabsTrigger value="json">JSON</TabsTrigger>
                  </TabsList>
                  
                  <TabsContent value="result" className="mt-4">
                    <ScrollArea className="h-[200px] w-full rounded border p-4">
                      <pre className="text-sm whitespace-pre-wrap">
                        {result.result}
                      </pre>
                    </ScrollArea>
                  </TabsContent>
                  
                  <TabsContent value="arguments" className="mt-4">
                    <ScrollArea className="h-[200px] w-full rounded border p-4">
                      <pre className="text-sm">
                        {JSON.stringify(result.arguments, null, 2)}
                      </pre>
                    </ScrollArea>
                  </TabsContent>
                  
                  <TabsContent value="json" className="mt-4">
                    <ScrollArea className="h-[200px] w-full rounded border p-4">
                      <pre className="text-sm">
                        {JSON.stringify(result, null, 2)}
                      </pre>
                    </ScrollArea>
                  </TabsContent>
                </Tabs>
              </CardContent>
            </Card>
          )}
        </>
      )}
    </div>
  )
}