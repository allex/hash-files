package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"flag"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"
	"time"

	"github.com/allex/calc-hash/helper"
	"github.com/allex/calc-hash/helper/logging"
	"github.com/bmatcuk/doublestar/v4"
)

var (
	appVersion = "1.0.0-dev"
	gitCommit  = "HEAD"
	buildTime  = time.Now().Format("2006-01-02 15:04:05")
)

// hashing algorithms as constants
const (
	MD5    = "md5"
	SHA1   = "sha1"
	SHA256 = "sha256"
	SHA512 = "sha512"
)

// getHashByName maps the hashing algorithm name to the function
func getHashByName(name string) func() hash.Hash {
	switch name {
	case MD5:
		return md5.New
	case SHA1:
		return sha1.New
	case SHA256:
		return sha256.New
	case SHA512:
		return sha512.New
	default:
		return nil
	}
}

// hashFiles calculates the hash of a given file list
func hashFiles(filePaths []string, hashAlgo string) (string, error) {
	hashFunc := getHashByName(hashAlgo)
	if hashFunc == nil {
		return "", fmt.Errorf("invalid or unknown hashing algorithm: %s", hashAlgo)
	}

	overallHash := hashFunc()
	for _, file := range filePaths {
		err := func(file string) error {
			f, err := os.Open(file)
			if err != nil {
				return err
			}
			defer f.Close()

			hash := hashFunc()
			if _, err := io.Copy(hash, f); err != nil {
				return err
			}

			sum := hex.EncodeToString(hash.Sum(nil))

			logging.Debug(fmt.Sprintf("%s("+helper.ANSI_BOLD_WHITE+"%s"+helper.ANSI_RESET+")"+" => "+helper.ANSI_BOLD_MAGENTA+"%s"+helper.ANSI_RESET+"\n", hashAlgo, file, sum))

			_, err = overallHash.Write([]byte(sum))
			if err != nil {
				return err
			}

			return nil
		}(file)

		if err != nil {
			return "", err
		}
	}

	return hex.EncodeToString(overallHash.Sum(nil)), nil
}

func collectFiles(patterns []string) ([]string, error) {
	var files []string

	indexes := map[string]bool{}

	for i := range patterns {
		matches, err := doublestar.FilepathGlob(patterns[i], doublestar.WithFilesOnly())
		if err != nil {
			return nil, fmt.Errorf("error occured while pattern matching: %s", err.Error())
		}

		// append files with uniq check
		for _, file := range matches {
			if find := indexes[file]; !find {
				files = append(files, file)
				indexes[file] = true
			}
		}
	}

	return files, nil
}

// die prints an error message, optionally shows help, and terminates the program with a specified exit code.
//
// Parameters:
//   - msg: The error message to be displayed.
//   - ec: The exit code with which the program should terminate.
func die(msg string, ec int) {
	logging.Error(msg)
	os.Exit(ec)
}

func main() {
	// define the flags
	algoFlag := flag.String("a", SHA1, "Hashing algorithm [md5|sha1|sha256|sha512]")
	versionFlag := flag.Bool("v", false, "Show version information")

	// trace for debug
	logLevel := flag.String("log-level", "error", "Set the log-level [debug,info,warn,error]")

	flag.Usage = func() {
		fmt.Printf("Usage: %s [-a hashing-algorithm] glob-pattern\n", os.Args[0])
		flag.PrintDefaults()
	}

	// parse the flags
	flag.Parse()

	// handle version request
	if *versionFlag {
		die(fmt.Sprintf("calc-hash version: %s (gitCommit: %s, builtTime: %s)\n", appVersion, gitCommit, buildTime), 0)
	}

	if err := logging.SetLogLevel(*logLevel); err != nil {
		die(err.Error(), 1)
	}

	globs := flag.Args()
	if len(globs) == 0 {
		flag.Usage()
		os.Exit(0)
	}

	files, err := collectFiles(globs)
	if err != nil {
		die(err.Error(), 1)
	}

	if len(files) == 0 {
		logging.Warn(fmt.Sprintf("The glob pattern matches empty: %s\n", strings.Join(globs, ",")))
	} else {
		hash, err := hashFiles(files, *algoFlag)
		if err != nil {
			die(err.Error(), 1)
		}
		fmt.Println(hash)
	}
}
