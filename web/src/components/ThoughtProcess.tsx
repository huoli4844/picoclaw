import { useState } from 'react'
import { ChevronDown, ChevronRight, Brain, Wrench, CheckCircle, AlertCircle, Clock } from 'lucide-react'
import { Button } from './ui/button'
import { ScrollArea } from './ui/scroll-area'
import { Thought } from '@/types'

interface ThoughtProcessProps {
  thoughts: Thought[]
}

export function ThoughtProcess({ thoughts }: ThoughtProcessProps) {
  const [isExpanded, setIsExpanded] = useState(true)

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
        return 'border-l-blue-200 bg-blue-50/50'
      case 'tool_call':
        return 'border-l-orange-200 bg-orange-50/50'
      case 'tool_result':
        if (thought.result?.includes('❌') || thought.result?.includes('失败')) {
          return 'border-l-red-200 bg-red-50/50'
        }
        return 'border-l-green-200 bg-green-50/50'
      default:
        return 'border-l-gray-200 bg-gray-50/50'
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
    <div className="border rounded-lg bg-background">
      <Button
        variant="ghost"
        className="w-full justify-between p-3 h-auto"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <div className="flex items-center gap-2">
          <Brain className="w-4 h-4" />
          <span className="font-medium">AI 思考过程</span>
          <span className="text-sm text-muted-foreground">
            ({thoughts.length} 步)
          </span>
        </div>
        {isExpanded ? (
          <ChevronDown className="w-4 h-4" />
        ) : (
          <ChevronRight className="w-4 h-4" />
        )}
      </Button>

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
                    <div className="text-sm break-words">
                      {thought.content}
                    </div>
                    {thought.args && (
                      <details className="mt-2">
                        <summary className="text-xs text-muted-foreground cursor-pointer hover:text-foreground">
                          查看参数
                        </summary>
                        <pre className="text-xs bg-muted p-2 rounded mt-1 overflow-x-auto">
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