import { useState, useEffect } from 'react'
import { FileBrowserProps } from './types'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { 
  Folder, 
  File, 
  ArrowLeft, 
  FileText, 
  Image, 
  Archive,
  Code,
  Settings,
  Database,
  Mail,
  LogIn,
  ChevronRight,
  Eye,
  Download,
  Trash2
} from 'lucide-react'

export function FileBrowser({ 
  files, 
  currentPath, 
  isLoading, 
  error, 
  onFileClick, 
  onDirectoryClick, 
  onNavigateUp,
  onFileContent,
  onDeleteFile
}: FileBrowserProps) {
  const [selectedFile, setSelectedFile] = useState<string | null>(null)

  // 获取文件图标
  const getFileIcon = (fileName: string, isDir: boolean) => {
    if (isDir) {
      return <Folder className="w-5 h-5 text-blue-500" />
    }

    const ext = fileName.toLowerCase().split('.').pop()
    
    switch (ext) {
      case 'md':
      case 'txt':
        return <FileText className="w-5 h-5 text-gray-600" />
      case 'json':
      case 'yaml':
      case 'yml':
      case 'toml':
        return <Database className="w-5 h-5 text-green-600" />
      case 'jpg':
      case 'jpeg':
      case 'png':
      case 'gif':
      case 'svg':
        return <Image className="w-5 h-5 text-purple-600" />
      case 'zip':
      case 'tar':
      case 'gz':
        return <Archive className="w-5 h-5 text-orange-600" />
      case 'go':
      case 'js':
      case 'ts':
      case 'tsx':
      case 'py':
      case 'html':
      case 'css':
        return <Code className="w-5 h-5 text-cyan-600" />
      case 'log':
        return <LogIn className="w-5 h-5 text-red-600" />
      case 'html':
        return <FileText className="w-5 h-5 text-orange-500" />
      case 'json':
        return <Database className="w-5 h-5 text-blue-600" />
      default:
        return <File className="w-5 h-5 text-gray-500" />
    }
  }

  // 格式化文件大小
  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  // 格式化修改时间
  const formatModTime = (modTime: string) => {
    return new Date(modTime).toLocaleString('zh-CN')
  }

  const handleFileClick = async (file: any) => {
    if (file.isDir) {
      onDirectoryClick(file.path)
    } else {
      setSelectedFile(file.name)
      onFileClick(file)
    }
  }

  const handleViewFile = async (file: any) => {
    if (!file.isDir) {
      await onFileContent(file.path)
    }
  }

  const handleDelete = async (file: any, event: React.MouseEvent) => {
    event.stopPropagation()
    
    const fileType = file.isDir ? '目录' : '文件'
    const fileName = file.name
    
    if (window.confirm(`确定要删除${fileType} "${fileName}" 吗？此操作不可恢复。`)) {
      const success = await onDeleteFile(file.path, file.isDir)
      if (success) {
        setSelectedFile(null)
      }
    }
  }

  // 排序文件：目录在前，然后按名称排序
  const sortedFiles = [...files].sort((a, b) => {
    if (a.isDir !== b.isDir) {
      return a.isDir ? -1 : 1
    }
    return a.name.localeCompare(b.name)
  })

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="w-8 h-8 border-2 border-primary border-t-transparent rounded-full animate-spin mx-auto mb-2"></div>
          <p className="text-sm text-muted-foreground">加载中...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="w-12 h-12 bg-red-100 dark:bg-red-900/30 rounded-full flex items-center justify-center mx-auto mb-2">
            <Folder className="w-6 h-6 text-red-600 dark:text-red-400" />
          </div>
          <p className="text-sm text-red-600 dark:text-red-400">加载失败</p>
          <p className="text-xs text-muted-foreground mt-1">{error}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col">
      {/* 路径导航 */}
      <div className="flex items-center gap-2 p-3 border-b bg-muted/30">
        <Button
          variant="ghost"
          size="sm"
          onClick={onNavigateUp}
          disabled={!currentPath || currentPath.endsWith('.picoclaw')}
          className="h-8 px-2"
        >
          <ArrowLeft className="w-4 h-4" />
        </Button>
        
        <div className="flex items-center gap-1 flex-1 min-w-0">
          <Database className="w-4 h-4 text-primary flex-shrink-0" />
          <ChevronRight className="w-4 h-4 text-muted-foreground flex-shrink-0" />
          <span className="text-sm font-mono truncate">
            {currentPath.replace(/^.*\.picoclaw\//, '') || '.picoclaw'}
          </span>
        </div>
      </div>

      {/* 文件列表 */}
      <ScrollArea className="flex-1">
        <div className="p-3">
          {files.length === 0 ? (
            <div className="text-center py-8">
              <Folder className="w-12 h-12 text-muted-foreground mx-auto mb-2" />
              <p className="text-sm text-muted-foreground">此目录为空</p>
            </div>
          ) : (
            <div className="space-y-1">
              {sortedFiles.map((file, index) => (
                <div
                  key={index}
                  className={`
                    flex items-center gap-3 p-2 rounded-md cursor-pointer transition-colors group
                    hover:bg-accent/50 ${selectedFile === file.name ? 'bg-accent' : ''}
                  `}
                  onClick={() => handleFileClick(file)}
                >
                  {/* 文件图标 */}
                  <div className="flex-shrink-0">
                    {getFileIcon(file.name, file.isDir)}
                  </div>

                  {/* 文件信息 */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium truncate">
                        {file.name}
                      </span>
                      {file.isDir && (
                        <span className="text-xs text-muted-foreground">目录</span>
                      )}
                    </div>
                    <div className="flex items-center gap-3 text-xs text-muted-foreground">
                      {!file.isDir && (
                        <>
                          <span>{formatFileSize(file.size)}</span>
                          <span>•</span>
                        </>
                      )}
                      <span>{formatModTime(file.modTime)}</span>
                    </div>
                  </div>

                  {/* 操作按钮 */}
                  <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                    {!file.isDir && (
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-6 w-6 p-0"
                        onClick={(e) => {
                          e.stopPropagation()
                          handleViewFile(file)
                        }}
                        title="查看文件"
                      >
                        <Eye className="w-3 h-3" />
                      </Button>
                    )}
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-6 w-6 p-0 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-950/20"
                      onClick={(e) => handleDelete(file, e)}
                      title={`删除${file.isDir ? '目录' : '文件'}`}
                    >
                      <Trash2 className="w-3 h-3" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </ScrollArea>
    </div>
  )
}