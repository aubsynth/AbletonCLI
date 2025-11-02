package cmd

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(migrate)
	migrate.Flags().String("directory", ".", "Directory to search")
	migrate.Flags().String("replace", "", "String to replace")
	migrate.Flags().String("with", "", "String to replace with")
	migrate.Flags().BoolVar(&dryRun, "dry-run", false, "Perform a trial run with no changes made")
	migrate.MarkFlagRequired("replace")
	migrate.MarkFlagRequired("with")
}

var (
	dryRun              bool
	totalFiles          int
	filesWithMatches    int
	filesWithoutMatches int
	filesModified       int
	linesModified       int
	migrate             = &cobra.Command{
		Use:   "migrate",
		Short: "Migrate files",
		Long:  `A command to migrate files`,
		Run:   migrateRun,
	}
)

func migrateRun(cmd *cobra.Command, args []string) {
	// Reset counters
	totalFiles = 0
	filesWithMatches = 0
	filesWithoutMatches = 0
	filesModified = 0
	linesModified = 0

	directoryVal, err := cmd.Flags().GetString("directory")
	if err != nil {
		logError(fmt.Sprintf("failed to read --directory flag: %v", err))
		return
	}
	replaceVal, err := cmd.Flags().GetString("replace")
	if err != nil {
		logError(fmt.Sprintf("failed to read --replace flag: %v", err))
		return
	}
	withVal, err := cmd.Flags().GetString("with")
	if err != nil {
		logError(fmt.Sprintf("failed to read --with flag: %v", err))
		return
	}

	fileExtension := ".als" // Ableton File Extension
	logInfo(fmt.Sprintf("Finding files in directory: %s with file extension: %s", directoryVal, fileExtension))

	// First, count total files to process
	err = filepath.WalkDir(directoryVal, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		// match extension case-insensitively (.als, .ALS, etc.)
		if !d.IsDir() && strings.EqualFold(filepath.Ext(d.Name()), fileExtension) {
			totalFiles++
		}
		return nil
	})
	if err != nil {
		logError(fmt.Sprintf("error walking the path %q: %v", directoryVal, err))
	}
	err = filepath.WalkDir(directoryVal, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		// match extension case-insensitively (.als, .ALS, etc.)
		if !d.IsDir() && strings.EqualFold(filepath.Ext(d.Name()), fileExtension) {
			logInfo(fmt.Sprintf("Processing %s of %s files: %s", fmt.Sprintf("%d", filesWithMatches+filesWithoutMatches+1), fmt.Sprintf("%d", totalFiles), path))
			tempGz := path + ".tmp.gz"
			tempXml := path + ".tmp.xml"

			// Ensure temp files are cleaned up when we're done with this file
			defer func() {
				os.Remove(tempGz)
				os.Remove(tempXml)
			}()

			// copy the .als (gzipped) to a temp file
			if err := copyFile(path, tempGz); err != nil {
				logError(fmt.Sprintf("Error copying file: %v", err))
				return nil
			}
			logDebug(fmt.Sprintf("File '%s' copied to '%s' successfully.", path, tempGz))

			// decompress the temp gz to a temp xml
			if err := decompressGzipFile(tempGz, tempXml); err != nil {
				logError(fmt.Sprintf("Error decompressing file: %v", err))
				return nil
			}
			logDebug(fmt.Sprintf("File '%s' decompressed to '%s' successfully.", tempGz, tempXml))

			// modify the temp xml file
			hadMatches, err := processFile(tempXml, replaceVal, withVal)
			if err != nil {
				logError(fmt.Sprintf("failed to replace file paths for %s: %v", tempXml, err))
				return nil
			}

			if !dryRun && hadMatches {
				logDebug(fmt.Sprintf("Replacement completed for %s", tempXml))
				// compress the modified XML back into a gz and replace the original .als
				if err := compressGzipFile(tempXml, tempGz); err != nil {
					logError(fmt.Sprintf("failed to compress modified xml %s -> %s: %v", tempXml, tempGz, err))
					return nil
				}
				// replace the original file with the new gz
				if err := os.Rename(tempGz, path); err != nil {
					logError(fmt.Sprintf("failed to replace original file %s with %s: %v", path, tempGz, err))
					return nil
				}
				logInfo(fmt.Sprintf("Updated original file: %s", path))
				filesModified++
			}
		}
		return nil
	})
	if err != nil {
		logError(fmt.Sprintf("error walking the path %q: %v", directoryVal, err))
	}
	logInfo(fmt.Sprintf("Total files found: %d", totalFiles))
	logInfo(fmt.Sprintf("Files with matches: %d", filesWithMatches))
	logInfo(fmt.Sprintf("Files without matches: %d", filesWithoutMatches))
	logInfo(fmt.Sprintf("Files modified: %d", filesModified))
	logDebug(fmt.Sprintf("Lines modified: %d", linesModified))
}

func copyFile(src, dst string) error {
	logDebug(fmt.Sprintf("Copying file from %s to %s", src, dst))
	// Open the source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close() // Ensure the source file is closed

	// Create the destination file
	destinationFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destinationFile.Close() // Ensure the destination file is closed

	// Copy the contents from source to destination
	if _, err := io.Copy(destinationFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// preserve permissions from source
	if fi, err := os.Stat(src); err == nil {
		if err := os.Chmod(dst, fi.Mode()); err != nil {
			return fmt.Errorf("failed to set file mode: %w", err)
		}
	}

	return nil
}

func decompressGzipFile(src, dst string) error {
	logDebug(fmt.Sprintf("Decompressing file from %s to %s", src, dst))
	// Open the source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close() // Ensure the source file is closed

	// Create a gzip.Reader from the input file
	gzipReader, err := gzip.NewReader(sourceFile)
	if err != nil {
		return fmt.Errorf("error creating gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// Create the destination file
	destinationFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destinationFile.Close() // Ensure the destination file is closed

	// Copy the contents from source to destination
	if _, err := io.Copy(destinationFile, gzipReader); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// preserve permissions from source
	if fi, err := os.Stat(src); err == nil {
		if err := os.Chmod(dst, fi.Mode()); err != nil {
			return fmt.Errorf("failed to set file mode: %w", err)
		}
	}

	return nil
}

func compressGzipFile(src, dst string) error {
	logDebug(fmt.Sprintf("Compressing file from %s to %s", src, dst))
	// Open the source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file for compression: %w", err)
	}
	defer sourceFile.Close()

	// Create the destination file
	destinationFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination gzip file: %w", err)
	}
	defer destinationFile.Close()

	// Create a gzip.Writer
	gzipWriter := gzip.NewWriter(destinationFile)
	defer gzipWriter.Close()

	if _, err := io.Copy(gzipWriter, sourceFile); err != nil {
		return fmt.Errorf("failed to write compressed data: %w", err)
	}

	// preserve permissions from source
	if fi, err := os.Stat(src); err == nil {
		if err := os.Chmod(dst, fi.Mode()); err != nil {
			return fmt.Errorf("failed to set file mode: %w", err)
		}
	}

	return nil
}

func processFile(src, replaceVal, withVal string) (bool, error) {
	var mode string
	if dryRun {
		mode = "dry-run"
	} else {
		mode = "normal"
	}
	logDebug(fmt.Sprintf("Processing file in %s mode from %s", mode, src))
	// Read the source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return false, fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Declare variables in outer scope so they're accessible throughout the function
	var writer *bufio.Writer
	var tempFileName string
	var tempFile *os.File
	renamed := false
	foundMatches := false

	if !dryRun {
		// Create a temporary file in the same directory as source for atomic rename
		tempDir := filepath.Dir(src)
		tempFile, err = os.CreateTemp(tempDir, filepath.Base(src)+".tmp-")
		if err != nil {
			return false, fmt.Errorf("failed to create temporary file: %w", err)
		}
		defer tempFile.Close()
		tempFileName = tempFile.Name()
		// Clean up temp file only if we don't successfully rename it
		defer func() {
			if !renamed {
				os.Remove(tempFileName)
			}
		}()
		// create a buffered writer with 256KB buffer for better performance
		writer = bufio.NewWriterSize(tempFile, 256*1024)
	}

	// Scan through the source file line by line
	scanner := bufio.NewScanner(sourceFile)
	// increase scanner buffer to handle very long lines (Ableton XML can be long)
	buf := make([]byte, 0, 1024*1024)
	// allow tokens up to 10MB
	scanner.Buffer(buf, 10*1024*1024)
	lineNumber := 0

	// process each line
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if strings.Contains(line, replaceVal) {
			foundMatches = true
			logDebug(fmt.Sprintf("Line %d: %s", lineNumber, line))
			if !dryRun {
				// replace only the matched substring (case-sensitive)
				newLine := strings.ReplaceAll(line, replaceVal, withVal)
				if _, err := writer.WriteString(newLine + "\n"); err != nil {
					return false, fmt.Errorf("failed to write replacement line: %w", err)
				} else {
					linesModified++
				}
			}
		} else {
			if !dryRun {
				if _, err := writer.WriteString(line + "\n"); err != nil {
					return false, fmt.Errorf("failed to write original line: %w", err)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error scanning file: %w", err)
	}

	// Only do file operations if not in dry-run mode
	if !dryRun {
		// flush writer to ensure data is written to the temp file
		if err := writer.Flush(); err != nil {
			return false, fmt.Errorf("error flushing writer: %w", err)
		}

		// close temp file before operations
		if err := tempFile.Close(); err != nil {
			return false, fmt.Errorf("failed to close temp file: %w", err)
		}

		// preserve permissions from source
		if fi, err := os.Stat(src); err == nil {
			if err := os.Chmod(tempFileName, fi.Mode()); err != nil {
				return false, fmt.Errorf("failed to set file mode: %w", err)
			}
		}
		if foundMatches {
			// replace the original XML file with the modified temp file
			if err := os.Rename(tempFileName, src); err != nil {
				return false, fmt.Errorf("failed to replace original file: %w", err)
			}
			renamed = true
		}
	}

	// Update counters based on whether we found matches
	if foundMatches {
		filesWithMatches++
	} else {
		filesWithoutMatches++
	}
	return foundMatches, nil
}

// cleanUpTempFiles walks the given base directory and removes any leftover temporary files
// created during processing (files ending with .tmp.gz, .tmp.xml, or .tmp-*).
// This is a safety net - normally temp files are cleaned up immediately after use.
func cleanUpTempFiles(baseDir string) {
	logDebug("Cleaning up any leftover temporary files...")
	patterns := []string{".tmp.gz", ".tmp.xml"}

	foundAny := false
	err := filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// log and continue walking
			logWarn(fmt.Sprintf("skipping path %q due to error: %v", path, err))
			return nil
		}
		if d.IsDir() {
			return nil
		}

		// Check for standard temp patterns
		for _, suf := range patterns {
			if strings.HasSuffix(d.Name(), suf) {
				if !foundAny {
					logInfo("Found leftover temporary files, cleaning up...")
					foundAny = true
				}
				logDebug(fmt.Sprintf("Removing temporary file: %s", path))
				if remErr := os.Remove(path); remErr != nil {
					logError(fmt.Sprintf("failed to remove file %q: %v", path, remErr))
				}
				return nil
			}
		}

		// Also check for the temp files created by os.CreateTemp (contain ".tmp-")
		if strings.Contains(d.Name(), ".tmp-") {
			if !foundAny {
				logInfo("Found leftover temporary files, cleaning up...")
				foundAny = true
			}
			logDebug(fmt.Sprintf("Removing temporary file: %s", path))
			if remErr := os.Remove(path); remErr != nil {
				logError(fmt.Sprintf("failed to remove file %q: %v", path, remErr))
			}
		}

		return nil
	})
	if err != nil {
		logError(fmt.Sprintf("error during cleanup walk: %v", err))
	}

	if !foundAny {
		logDebug("No leftover temporary files found")
	}
}
