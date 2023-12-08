package utils

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

func GetMyPath() (path string, err error) {
	thisExec, err := exec.LookPath(os.Args[0])
	if err != nil {
	}

	// check if that path to the executable is a symlink and follow it if so
	// to determine actual executable location
	path, err = ResolveSymlink(thisExec)
	if err != nil {
		return
	}
	return
}

// matchString is generic, simple, case insensitive regexp string matching.
//
// This function is to answer question: does string 's' contain substring 'substr' at specified match position.
//
// Parameters:
//
// - matchPosition (string): can be one of 'start', 'exact'; start means 's' must start with substr and exact means it must match fully
// - s (string): string s to match substr against
// - substr (string): substring to be matched
//
// Returns:
// - bool: true if 'substr' matched in 's'
func MatchString(matchPosition, s, substr string) bool {
	var pattern string

	s = stringTrimAndToLower(s)

	switch matchPosition {
	case "start":
		pattern = fmt.Sprintf("(?i)^%s", substr)
	case "exact":
		pattern = fmt.Sprintf("(?i)^%s$", substr)
	}

	r, err := regexp.MatchString(pattern, s)
	if err != nil || r == false {
		return false
	}
	return true
}

// kvExtractor extracts key and value fields by separating given string s by separator
//
// This funcition is splitting the arguments 's' on ':' and returns the product as a slice os strings.
//
// Parameters:
//
// - s (string): a string to be splitted
//
// Returns:
// - kvPair ([]slice): a slice of 2 strings
// - err (error): error if any or nil
func KvExtractor(s string) (kvPair []string, err error) {
	if strings.Contains(s, ":") {
		result := strings.SplitN(s, ":", 3)
		kvPair = []string{
			strings.TrimSpace(result[0]),
			strings.TrimSpace(result[1]),
		}
		return
	}
	return []string{}, errors.New("String doesn't contain separator ':'")
}

// resolveSymlink is checking if a path is symlink
//
// This function checks if a path is symlink and returns its target if it was symlink. If it wasn't as symlink then it returns back the 'path'. This is safe to use on symlinks and regular files.
//
// Parameters:
//
// - path (string): a path to be checked
//
// Returns:
// - target (string): path to target of a symlink of 'path' if not a symlink
// - err (error): error if any or nil
func ResolveSymlink(path string) (target string, err error) {
	fileInfo, err := os.Lstat(path)
	if fileInfo.Mode()&os.ModeSymlink != 0 {
		// it's a symlink, resolve it
		target, err = os.Readlink(path)
		if err != nil {
			return "", err
		}
	} else {
		target = path
	}
	return
}

// CopyFile copies a local file from source to destination
//
// This functions opens source and destination files using 'os' package and then copies it using 'io' package
//
// Parameters:
//
// - src (string): source file
// - dst (string): destination file
//
// Returns:
//
// - err (error): error if any or nil
func CopyFile(src string, dst string) (err error) {
	srcHandle, err := os.Open(src)
	if err != nil {
		return
	}

	dstHandle, err := os.Create(dst)
	if err != nil {
		return
	}

	_, err = io.Copy(dstHandle, srcHandle)

	return
}

// ExpandTilde expands '~' in path to users home dir
//
// This function checks if first character of  'path' is the '~' char
// and if so it gets replaced with current user's home dir path
//
// Parameters:
//
// - path (string): path starting with '~'
//
// Returns:
//
// - path (string): path with '~' expanded
// - err (error): error if any or nil
func ExpandTilde(path string) (pathExp string, err error) {
	if len(path) > 0 && path[0] == '~' {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		pathExp = filepath.Join(usr.HomeDir, path[1:])
		return pathExp, nil
	}
	return path, nil
}

// Local functions

// stringTrimAndToLower trims spaces and lower caps the string
func stringTrimAndToLower(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}
