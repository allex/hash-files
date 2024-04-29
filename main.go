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
	"time"

	"github.com/allex/calc-hash/helper"
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

var algoFlag *string

var trace bool

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

// printTrace writes formatted output to standard error.
func printTrace(format string, a ...any) {
	if trace {
		fmt.Fprintf(os.Stderr, format, a...)
	}
}

func printError(msg string) {
	fmt.Fprint(os.Stderr, msg)
}

// HashFiles calculates the hash of a given file list
func HashFiles(filePaths []string, hashAlgo func() hash.Hash) (string, error) {
	overallHash := hashAlgo()

	for _, file := range filePaths {
		err := func(file string) error {
			f, err := os.Open(file)
			if err != nil {
				return err
			}
			defer f.Close()

			hash := hashAlgo()
			if _, err := io.Copy(hash, f); err != nil {
				return err
			}

			sum := hex.EncodeToString(hash.Sum(nil))
			printTrace("-> calc "+helper.ANSI_BOLD_WHITE+"%s"+helper.ANSI_RESET+" - "+helper.ANSI_BOLD_MAGENTA+"%s"+helper.ANSI_RESET+"\n", file, sum)

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

func main() {
	// define the flags
	algoFlag = flag.String("a", SHA1, "Hashing algorithm [md5|sha1|sha256|sha512]")
	versionFlag := flag.Bool("v", false, "Show version information")

	// trace for debug
	flag.BoolVar(&trace, "trace", false, "Trace the hash files")

	flag.Usage = func() {
		fmt.Printf("Usage: %s [-a hashing-algorithm] glob-pattern\n", os.Args[0])
		flag.PrintDefaults()
	}

	// parse the flags
	flag.Parse()

	// handle version request
	if *versionFlag {
		fmt.Printf("calc-hash version: %s (gitCommit: %s, builtTime: %s)\n", appVersion, gitCommit, buildTime)
		os.Exit(0) // Terminate the program after printing version
	}

	hashAlgo := getHashByName(*algoFlag)
	if hashAlgo == nil {
		printError(fmt.Sprintf("Invalid or unknown hashing algorithm: %s\n", *algoFlag))
		flag.Usage()
		os.Exit(1)
	}

	files, err := collectFiles(flag.Args())
	if err != nil {
		printError(err.Error())
		flag.Usage()
		os.Exit(1)
	}

	if len(files) == 0 {
		printTrace("The glob pattern matches empty")
	} else {
		hash, err := HashFiles(files, hashAlgo)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Println(hash)
	}
}
