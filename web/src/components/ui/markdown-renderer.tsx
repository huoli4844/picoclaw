interface MarkdownRendererProps {
  content: string
  className?: string
}

export function MarkdownRenderer({ content, className = '' }: MarkdownRendererProps) {

  // 简单的Markdown解析器
  const parseMarkdown = (text: string) => {
    // 处理代码块
    const codeBlockRegex = /```(\w+)?\n([\s\S]*?)```/g
    const parts: Array<{ type: string; content: string; language?: string }> = []
    let lastIndex = 0
    let match

    while ((match = codeBlockRegex.exec(text)) !== null) {
      // 添加代码块前的文本
      if (match.index > lastIndex) {
        const beforeText = text.substring(lastIndex, match.index)
        if (beforeText.trim()) {
          parts.push({ type: 'text', content: beforeText })
        }
      }

      // 添加代码块
      parts.push({
        type: 'code',
        content: match[2].trim(),
        language: match[1] || 'text'
      })

      lastIndex = codeBlockRegex.lastIndex
    }

    // 添加剩余的文本
    if (lastIndex < text.length) {
      const remainingText = text.substring(lastIndex)
      if (remainingText.trim()) {
        parts.push({ type: 'text', content: remainingText })
      }
    }

    // 如果没有代码块，整个内容作为文本
    if (parts.length === 0) {
      parts.push({ type: 'text', content: text })
    }

    return parts
  }

  const renderTextContent = (text: string) => {
    // 处理标题
    text = text.replace(/^### (.*$)/gim, '<h3 class="text-lg font-semibold mb-2 mt-4 text-foreground">$1</h3>')
    text = text.replace(/^## (.*$)/gim, '<h2 class="text-xl font-bold mb-3 mt-6 text-foreground">$1</h2>')
    text = text.replace(/^# (.*$)/gim, '<h1 class="text-2xl font-bold mb-4 mt-8 text-foreground">$1</h1>')

    // 处理粗体和斜体
    text = text.replace(/\*\*(.*?)\*\*/g, '<strong class="font-semibold text-foreground">$1</strong>')
    text = text.replace(/\*(.*?)\*/g, '<em class="italic text-foreground">$1</em>')

    // 处理链接
    text = text.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener noreferrer" class="text-primary hover:underline">$1</a>')

    // 处理行内代码
    text = text.replace(/`([^`]+)`/g, '<code class="bg-muted px-1.5 py-0.5 rounded text-sm font-mono">$1</code>')

    // 处理列表
    text = text.replace(/^\* (.+)$/gim, '<li class="ml-4">• $1</li>')
    text = text.replace(/(<li.*<\/li>)/s, '<ul class="list-disc space-y-1 mb-3">$1</ul>')

    // 处理数字列表
    text = text.replace(/^\d+\. (.+)$/gim, '<li class="ml-4">$1</li>')

    // 处理段落
    text = text.replace(/\n\n/g, '</p><p class="mb-4 leading-relaxed">')
    text = '<p class="mb-4 leading-relaxed">' + text + '</p>'

    // 清理多余的段落标签
    text = text.replace(/<p class="mb-4 leading-relaxed"><\/p>/g, '')
    text = text.replace(/<p class="mb-4 leading-relaxed">(.*?)<\/ul>/g, '$1</ul>')

    return text
  }

  const parts = parseMarkdown(content)

  return (
    <div className={`markdown-content space-y-4 ${className}`}>
      {parts.map((part, index) => {
        if (part.type === 'code') {
          return (
            <div key={index} className="relative w-full">
              <div className="bg-muted border rounded-lg overflow-hidden w-full">
                <div className="flex items-center justify-between px-4 py-2 bg-muted/50 border-b">
                  <span className="text-sm font-medium text-muted-foreground">
                    {part.language}
                  </span>
                  <button
                    onClick={() => {
                      navigator.clipboard.writeText(part.content)
                    }}
                    className="text-xs px-2 py-1 rounded bg-background hover:bg-muted transition-colors"
                  >
                    复制
                  </button>
                </div>
                <pre className="p-4 overflow-x-auto text-sm w-full">
                  <code className="font-mono break-words">{part.content}</code>
                </pre>
              </div>
            </div>
          )
        } else {
          return (
            <div
              key={index}
              dangerouslySetInnerHTML={{ __html: renderTextContent(part.content) }}
            />
          )
        }
      })}
    </div>
  )
}