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

			_, err = overallHash.Write([]byte(hex.EncodeToString(hash.Sum(nil))))
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

func main() {
	// define the flags
	algoFlag = flag.String("a", SHA1, "Hashing algorithm [md5|sha1|sha256|sha512]")
	versionFlag := flag.Bool("v", false, "Show version information")

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

	globPattern := flag.Arg(0)

	if globPattern == "" {
		flag.Usage()
		os.Exit(1)
	}

	hashAlgo := getHashByName(*algoFlag)
	if hashAlgo == nil {
		fmt.Printf("Invalid or unknown hashing algorithm: %s\n", *algoFlag)
		flag.Usage()
		os.Exit(1)
	}

	matches, err := doublestar.FilepathGlob(globPattern, doublestar.WithFilesOnly())
	if err != nil {
		fmt.Printf("Error occured while pattern matching: %s\n", err.Error())
		os.Exit(1)
	}

	sum := ""
	if len(matches) > 0 {
		hash, err := HashFiles(matches, hashAlgo)
		if err != nil {
			fmt.Printf("Error computing hash for files %v: %s\n", matches, err.Error())
			os.Exit(1)
		}
		sum = hash
	}

	fmt.Println(sum)
}
