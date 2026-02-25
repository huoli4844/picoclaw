export interface FileInfo {
  name: string
  path: string
  isDir: boolean
  size: number
  modTime: string
}

export interface FileContent {
  success: boolean
  path: string
  content: string
  size: number
  contentType: string
}

export interface FileBrowserProps {
  files: FileInfo[]
  currentPath: string
  isLoading: boolean
  error: string | null
  
  onFileClick: (file: FileInfo) => void
  onDirectoryClick: (path: string) => void
  onNavigateUp: () => void
  onFileContent: (path: string) => Promise<void>
  onDeleteFile: (path: string, isDir: boolean) => Promise<boolean>
}