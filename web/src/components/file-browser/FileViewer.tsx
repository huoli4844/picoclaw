import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { X, Copy, Download } from 'lucide-react'
import { FileContent } from './types'

interface FileViewerProps {
  fileContent: FileContent | null
  isOpen: boolean
  onClose: () => void
}

export function FileViewer({ fileContent, isOpen, onClose }: FileViewerProps) {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    if (fileContent?.content) {
      try {
        await navigator.clipboard.writeText(fileContent.content)
        setCopied(true)
        setTimeout(() => setCopied(false), 2000)
      } catch (error) {
        console.error('Failed to copy to clipboard:', error)
      }
    }
  }

  const handleDownload = () => {
    if (fileContent) {
      const blob = new Blob([fileContent.content], { type: fileContent.contentType })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = fileContent.path.split('/').pop() || 'file'
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
    }
  }

  if (!isOpen || !fileContent) {
    return null
  }

  // 检测是否为代码文件，用于语法高亮
  const isCodeFile = (path: string) => {
    const codeExtensions = ['.go', '.js', '.ts', '.tsx', '.py', '.html', '.css', '.json', '.yaml', '.yml', '.sh', '.bat']
    return codeExtensions.some(ext => path.toLowerCase().endsWith(ext))
  }

  const fileName = fileContent.path.split('/').pop() || 'unknown'
  const isCode = isCodeFile(fileContent.path)

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-background border rounded-lg shadow-lg w-[90vw] h-[80vh] flex flex-col">
        {/* 头部 */}
        <div className="flex items-center justify-between p-4 border-b">
          <div className="flex items-center gap-3">
            <div className="font-mono text-sm bg-muted px-2 py-1 rounded">
              {fileName}
            </div>
            <div className="text-xs text-muted-foreground">
              {fileContent.contentType} • {(fileContent.size / 1024).toFixed(1)} KB
            </div>
          </div>
          
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={handleCopy}
              className="h-8 px-2"
            >
              <Copy className="w-4 h-4 mr-1" />
              {copied ? '已复制' : '复制'}
            </Button>
            
            <Button
              variant="ghost"
              size="sm"
              onClick={handleDownload}
              className="h-8 px-2"
            >
              <Download className="w-4 h-4 mr-1" />
              下载
            </Button>
            
            <Button
              variant="ghost"
              size="sm"
              onClick={onClose}
              className="h-8 px-2"
            >
              <X className="w-4 h-4" />
            </Button>
          </div>
        </div>

        {/* 内容区域 */}
        <ScrollArea className="flex-1">
          <div className="p-4">
            {isCode ? (
              <pre className="bg-muted rounded-md p-4 overflow-x-auto text-sm font-mono">
                <code>{fileContent.content}</code>
              </pre>
            ) : (
              <div className="prose prose-sm max-w-none dark:prose-invert">
                {fileContent.content.startsWith('<!DOCTYPE') || 
                 fileContent.content.startsWith('<html') ? (
                  // HTML 文件预览
                  <div 
                    dangerouslySetInnerHTML={{ __html: fileContent.content }}
                    className="border rounded-md p-4"
                  />
                ) : (
                  // Markdown 或纯文本
                  <pre className="whitespace-pre-wrap text-sm font-mono bg-muted rounded-md p-4">
                    {fileContent.content}
                  </pre>
                )}
              </div>
            )}
          </div>
        </ScrollArea>
      </div>
    </div>
  )
}