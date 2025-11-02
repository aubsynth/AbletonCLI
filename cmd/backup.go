package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(backup)
	backup.Flags().String("directory", ".", "Directory to backup")
	backup.Flags().String("destination", "../backup/", "Backup destination directory")
}

var (
	total_files_backed_up int
	total_files_to_backup int
	backup                = &cobra.Command{
		Use:   "backup",
		Short: "Backup files",
		Long:  `A command to backup files`,
		Run:   backupRun,
	}
)

func backupRun(cmd *cobra.Command, args []string) {
	// Reset counters
	total_files_backed_up = 0
	total_files_to_backup = 0

	directoryVal, err := cmd.Flags().GetString("directory")
	if err != nil {
		logError(fmt.Sprintf("failed to read --directory flag: %v", err))
		return
	}
	destinationPath, err := cmd.Flags().GetString("destination")
	if err != nil || destinationPath == "" {
		logError("Backup destination not specified or invalid.")
		return
	}
	fileExtension := ".als" // Ableton File Extension
	logInfo(fmt.Sprintf("Finding files in directory: %s with file extension: %s", directoryVal, fileExtension))

	err = filepath.WalkDir(directoryVal, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		// match extension case-insensitively (.als, .ALS, etc.)
		if !d.IsDir() && strings.EqualFold(filepath.Ext(d.Name()), fileExtension) {
			total_files_to_backup++
			logDebug(fmt.Sprintf("Found file: %s", path))
			// copy folder structure to backup location
			relativePath, err := filepath.Rel(directoryVal, path)
			if err != nil {
				logError(fmt.Sprintf("Error determining relative path: %v", err))
				return nil
			}
			destinationVal := filepath.Join(destinationPath, relativePath)
			// Get source file info
			sourceFileInfo, err := os.Stat(path)
			if err != nil {
				return err
			}
			// Copy file to destination
			if err := os.MkdirAll(filepath.Dir(destinationVal), sourceFileInfo.Mode()); err != nil {
				logError(fmt.Sprintf("Error creating directories: %v", err))
				return nil
			}
			if err := copyFile(path, destinationVal); err != nil {
				logError(fmt.Sprintf("Error copying file: %v", err))
				return nil
			}
			logDebug(fmt.Sprintf("File '%s' copied to '%s' successfully.", path, destinationVal))
			total_files_backed_up++
		}
		return nil
	})
	if err != nil {
		logError(fmt.Sprintf("error walking the path %q: %v", directoryVal, err))
		return
	}

	logInfo(fmt.Sprintf("Total files to backup: %d", total_files_to_backup))
	logInfo(fmt.Sprintf("Total files backed up: %d", total_files_backed_up))
}
