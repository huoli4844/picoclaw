import { useState, useCallback } from 'react'

interface FileInfo {
  name: string
  path: string
  isDir: boolean
  size: number
  modTime: string
}

interface FileContent {
  success: boolean
  path: string
  content: string
  size: number
  contentType: string
}

interface UseFileBrowserReturn {
  files: FileInfo[]
  currentPath: string
  isLoading: boolean
  error: string | null
  
  listFiles: (path?: string) => Promise<void>
  getFileContent: (path: string) => Promise<FileContent | null>
  navigateToDirectory: (path: string) => Promise<void>
  navigateUp: () => void
  deleteFileOrDirectory: (path: string, isDir: boolean) => Promise<boolean>
}

export function useFileBrowser(): UseFileBrowserReturn {
  const [files, setFiles] = useState<FileInfo[]>([])
  const [currentPath, setCurrentPath] = useState<string>('')
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const listFiles = useCallback(async (path?: string) => {
    setIsLoading(true)
    setError(null)
    
    try {
      const url = path 
        ? `http://localhost:8080/api/files?path=${encodeURIComponent(path)}`
        : 'http://localhost:8080/api/files'
      
      const response = await fetch(url)
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      
      const result = await response.json()
      if (result.success) {
        setFiles(result.files || [])
        setCurrentPath(result.path || '')
      } else {
        throw new Error(result.error || 'Failed to list files')
      }
    } catch (error) {
      console.error('Failed to list files:', error)
      setError(error instanceof Error ? error.message : 'Unknown error')
      setFiles([])
    } finally {
      setIsLoading(false)
    }
  }, [])

  const getFileContent = useCallback(async (path: string): Promise<FileContent | null> => {
    try {
      const url = `http://localhost:8080/api/file-content?path=${encodeURIComponent(path)}`
      const response = await fetch(url)
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      
      const result = await response.json()
      if (result.success) {
        return result
      } else {
        throw new Error(result.error || 'Failed to read file')
      }
    } catch (error) {
      console.error('Failed to read file:', error)
      setError(error instanceof Error ? error.message : 'Unknown error')
      return null
    }
  }, [])

  const navigateToDirectory = useCallback(async (path: string) => {
    await listFiles(path)
  }, [listFiles])

  const navigateUp = useCallback(() => {
    if (currentPath) {
      const parentPath = currentPath.split('/').slice(0, -1).join('/')
      if (parentPath) {
        listFiles(parentPath)
      } else {
        listFiles() // 回到默认目录
      }
    }
  }, [currentPath, listFiles])

  const deleteFileOrDirectory = useCallback(async (path: string, _isDir: boolean): Promise<boolean> => {
    try {
      const response = await fetch('http://localhost:8080/api/file-delete', {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ path })
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}))
        throw new Error(errorData.error || `HTTP error! status: ${response.status}`)
      }

      const result = await response.json()
      if (result.success) {
        // 刷新当前目录的文件列表
        await listFiles(currentPath)
        return true
      } else {
        throw new Error(result.error || 'Failed to delete')
      }
    } catch (error) {
      console.error('Failed to delete file/directory:', error)
      setError(error instanceof Error ? error.message : 'Unknown error')
      return false
    }
  }, [listFiles, currentPath])

  return {
    files,
    currentPath,
    isLoading,
    error,
    listFiles,
    getFileContent,
    navigateToDirectory,
    navigateUp,
    deleteFileOrDirectory
  }
}