// Package backup provides atomic file operations and encryption utilities
// for safe environment management.
//
// Principles:
// - Atomicity: All operations either fully complete or roll back
// - Durability: fsync() guarantees for critical operations
// - Cross-Platform: Consistent behavior across OSes
//
// Key Features:
// - Atomic file/directory copies with COW semantics
// - Age-compatible encryption
// - Checksum validation
// - Backup rotation utilities
//
// Usage:
//
//	import "github.com/shawnkhoffman/nix-foundry/pkg/backup"
//
//	err := backup.AtomicCopy(src, dest)
//	encrypted, err := backup.EncryptFile(path, key)
package backup // Provides atomic file operations and encryption
