package files

// FileOperation defines strongly-typed constants for permitted file operations.
type FileOperation string

const (
	OperationReadFile             FileOperation = "read_file"
	OperationReadBytes            FileOperation = "read_bytes"
	OperationReadText             FileOperation = "read_text"
	OperationListDirectory        FileOperation = "list_directory"
	OperationFileExists           FileOperation = "file_exists"
	OperationFileMetadata         FileOperation = "file_metadata"
	
	OperationCreateFile           FileOperation = "create_file"
	OperationWriteFile            FileOperation = "write_file"
	OperationAppendFile           FileOperation = "append_file"
	OperationCopyFile             FileOperation = "copy_file"
	OperationMoveFile             FileOperation = "move_file"
	OperationRenameFile           FileOperation = "rename_file"
	OperationDeleteFile           FileOperation = "delete_file"
	
	OperationCreateDirectory      FileOperation = "create_directory"
	OperationDeleteDirectory      FileOperation = "delete_directory"
	OperationDirectoryMetadata    FileOperation = "directory_metadata"
	
	OperationSearchFiles          FileOperation = "search_files"
	OperationCalculateHash        FileOperation = "calculate_hash"
	
	OperationTemporaryFile        FileOperation = "temporary_file"
	OperationTemporaryDirectory   FileOperation = "temporary_directory"
)

// IsValid validates if a string matches a known FileOperation.
func (o FileOperation) IsValid() bool {
	switch o {
	case OperationReadFile, OperationReadBytes, OperationReadText, OperationListDirectory,
		OperationFileExists, OperationFileMetadata, OperationCreateFile, OperationWriteFile,
		OperationAppendFile, OperationCopyFile, OperationMoveFile, OperationRenameFile,
		OperationDeleteFile, OperationCreateDirectory, OperationDeleteDirectory,
		OperationDirectoryMetadata, OperationSearchFiles, OperationCalculateHash,
		OperationTemporaryFile, OperationTemporaryDirectory:
		return true
	}
	return false
}
