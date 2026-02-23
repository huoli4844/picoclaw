import { useState } from 'react'
import { ChevronDown, ChevronRight, Brain, Wrench, CheckCircle, AlertCircle, Clock, Download } from 'lucide-react'
import { Button } from './ui/button'
import { ScrollArea } from './ui/scroll-area'
import { MarkdownRenderer } from './ui/markdown-renderer'
import { Thought } from '@/types'

interface ThoughtProcessProps {
  thoughts: Thought[]
}

export function ThoughtProcess({ thoughts }: ThoughtProcessProps) {
  const [isExpanded, setIsExpanded] = useState(true)

  const exportToJSON = () => {
    // 准备导出的数据，包含思考过程的完整信息
    const exportData = {
      exportInfo: {
        timestamp: new Date().toISOString(),
        totalThoughts: thoughts.length,
        exportType: 'AI思考过程'
      },
      thoughts: thoughts.map((thought, index) => ({
        index: index + 1,
        timestamp: thought.timestamp,
        type: thought.type,
        content: thought.content,
        toolName: thought.tool_name || null,
        args: thought.args ? JSON.parse(thought.args) : null,
        result: thought.result || null,
        duration: thought.duration || null
      }))
    }

    // 创建下载链接
    const dataStr = JSON.stringify(exportData, null, 2)
    const dataBlob = new Blob([dataStr], { type: 'application/json' })
    const url = URL.createObjectURL(dataBlob)
    
    // 创建临时下载链接并触发下载
    const link = document.createElement('a')
    link.href = url
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19)
    link.download = `ai-thoughts-${timestamp}.json`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
  }

  const getThoughtIcon = (thought: Thought) => {
    switch (thought.type) {
      case 'thinking':
        return <Brain className="w-4 h-4 text-blue-500" />
      case 'tool_call':
        return <Wrench className="w-4 h-4 text-orange-500" />
      case 'tool_result':
        if (thought.result?.includes('❌') || thought.result?.includes('失败')) {
          return <AlertCircle className="w-4 h-4 text-red-500" />
        }
        return <CheckCircle className="w-4 h-4 text-green-500" />
      default:
        return <Brain className="w-4 h-4 text-gray-500" />
    }
  }

  const getThoughtStyle = (thought: Thought) => {
    switch (thought.type) {
      case 'thinking':
        return 'border-l-blue-500/20 bg-blue-500/5 dark:bg-blue-500/10'
      case 'tool_call':
        return 'border-l-orange-500/20 bg-orange-500/5 dark:bg-orange-500/10'
      case 'tool_result':
        if (thought.result?.includes('❌') || thought.result?.includes('失败')) {
          return 'border-l-red-500/20 bg-red-500/5 dark:bg-red-500/10'
        }
        return 'border-l-green-500/20 bg-green-500/5 dark:bg-green-500/10'
      default:
        return 'border-l-gray-500/20 bg-gray-500/5 dark:bg-gray-500/10'
    }
  }

  const formatTimestamp = (timestamp: Date | string) => {
    const date = new Date(timestamp)
    return date.toLocaleTimeString('zh-CN', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    }) + '.' + date.getMilliseconds().toString().padStart(3, '0')
  }

  if (!thoughts || thoughts.length === 0) {
    return null
  }

  return (
    <div className="thought-process-container border rounded-lg bg-background">
      <div className="flex items-center justify-between p-3 border-b bg-muted/30">
        <div className="flex items-center gap-2">
          <Brain className="w-4 h-4" />
          <span className="font-medium">AI 思考过程</span>
          <span className="text-sm text-muted-foreground">
            ({thoughts.length} 步)
          </span>
        </div>
        
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={exportToJSON}
            title="导出为JSON文件"
          >
            <Download className="w-4 h-4 mr-1" />
            导出
          </Button>
          
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setIsExpanded(!isExpanded)}
            title={isExpanded ? "收起" : "展开"}
          >
            {isExpanded ? (
              <ChevronDown className="w-4 h-4" />
            ) : (
              <ChevronRight className="w-4 h-4" />
            )}
          </Button>
        </div>
      </div>

      {isExpanded && (
        <ScrollArea className="h-64 border-t">
          <div className="p-4 space-y-2">
            {thoughts.map((thought, index) => (
              <div
                key={index}
                className={`border-l-4 pl-4 py-2 pr-3 rounded-r-md ${getThoughtStyle(
                  thought
                )}`}
              >
                <div className="flex items-start gap-2">
                  {getThoughtIcon(thought)}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="text-xs text-muted-foreground">
                        {formatTimestamp(thought.timestamp)}
                      </span>
                      {thought.tool_name && (
                        <span className="text-xs font-mono bg-muted px-1.5 py-0.5 rounded">
                          {thought.tool_name}
                        </span>
                      )}
                      {thought.duration && (
                        <span className="text-xs text-muted-foreground flex items-center gap-1">
                          <Clock className="w-3 h-3" />
                          {thought.duration}ms
                        </span>
                      )}
                    </div>
                    <div className="thought-item-content text-sm break-words">
                      <MarkdownRenderer content={thought.content} />
                    </div>
                    {thought.args && (
                      <details className="mt-2">
                        <summary className="text-xs text-muted-foreground cursor-pointer hover:text-foreground">
                          查看参数
                        </summary>
                        <pre className="thought-code-block text-xs bg-muted p-2 rounded mt-1 overflow-x-auto">
                          {JSON.stringify(JSON.parse(thought.args), null, 2)}
                        </pre>
                      </details>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </ScrollArea>
      )}
    </div>
  )
}