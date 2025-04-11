package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	GitDir           = ".mini-git"
	ObjectTypeBlob   = "blob"
	ObjectTypeCommit = "commit"
	ObjectTypeTree   = "tree"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: mini-git <command> [args]")
	}

	command := os.Args[1]
	switch command {
	case "init":
		cmdInit()
	case "hash-object":
		if len(os.Args) < 3 {
			log.Fatalf("Usage: %s hash-object <file-path>", os.Args[0])
		}
		filePath := os.Args[2]
		oid, err := cmdHashObject(filePath)
		if err != nil {
			log.Fatalf("Error hashing object %s: %v", filePath, err)
		}
		fmt.Println(oid)
	case "cat-file":
		if len(os.Args) < 3 {
			log.Fatalf("Usage: %s cat-file <object-id>", os.Args[0])
		}
		oid := os.Args[2]
		content, err := cmdCatFile(oid)
		if err != nil {
			log.Fatalf("Error reading object %s: %v", oid, err)
		}
		fmt.Print(content)
	case "tree":
		if len(os.Args) < 3 {
			log.Fatalf("Usage: %s tree <path>", os.Args[0])
		}
		oid, err := cmdWriteTree(os.Args[2])
		if err != nil {
			log.Fatalf("Error writing tree: %v", err)
			return
		}
		fmt.Printf("Tree object: %s\n", oid)
	case "commit":
		oid, err := commit(os.Args[2])
		if err != nil {
			log.Fatalf("Error writing commit: %v", err)
		}
		fmt.Printf("Commit object: %s\n", oid)
	default:
		log.Fatal("unknown command")
	}
}

func check(e error) {
	if e != nil {
		log.Fatalf("Error: %v", e)
	}
}

func cmdInit() {
	err := os.Mkdir(GitDir, 0755)
	check(err)
	path, err := os.Getwd()
	check(err)
	fmt.Println("Initialized mini-git repository in: ", path)
}

func cmdHashObject(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	check(err)
	return hashBytesAsObject(ObjectTypeBlob, data)
}

func hashBytesAsObject(objectType string, data []byte) (string, error) {
	header := []byte(fmt.Sprintf("%s %d\x00", objectType, len(data)))
	fullObject := append(header, data...)

	hasher := sha1.New()
	_, err := hasher.Write(fullObject)
	check(err)

	hashBytes := hasher.Sum(nil)
	oid := hex.EncodeToString(hashBytes)

	objPath, err := getObjectPath(oid)
	check(err)

	objDir := filepath.Dir(objPath)
	err = os.MkdirAll(objDir, 0755)
	check(err)

	err = os.WriteFile(objPath, fullObject, 0644)
	check(err)
	return oid, nil
}

func cmdCatFile(oid string) (string, error) {
	objPath, err := getObjectPath(oid)
	check(err)

	fullObject, err := os.ReadFile(objPath)
	check(err)

	nullByteIndex := bytes.IndexByte(fullObject, 0)
	if nullByteIndex == -1 {
		return "", fmt.Errorf("malformed object file: missing null byte separator in %s", objPath)
	}

	content := fullObject[nullByteIndex+1:]
	return string(content), nil
}

func getObjectPath(oid string) (string, error) {
	if len(oid) < 2 {
		log.Fatalf("Invalid object ID: %s", oid)
	}

	subDir := oid[:2]
	fileName := oid[2:]

	return filepath.Join(GitDir, "objects", subDir, fileName), nil
}

type treeEntry struct {
	name string
	typ  string
	oid  string
}

func cmdWriteTree(directory string) (string, error) {
	pathDir := filepath.Clean(directory)

	dir, err := os.Open(pathDir)
	check(err)
	defer dir.Close()

	entries, err := dir.ReadDir(-1)
	check(err)

	var treeEntries []treeEntry

	for _, entry := range entries {
		fullPath := filepath.Join(pathDir, entry.Name())
		if isIgnored(fullPath) == true {
			continue
		}

		if !entry.IsDir() {
			info, err := os.Lstat(fullPath)
			check(err)
			if info.Mode()&os.ModeSymlink != 0 {
				continue
			}
			blobOID, err := cmdHashObject(fullPath)
			check(err)

			treeEntries = append(treeEntries, treeEntry{
				name: entry.Name(),
				oid:  blobOID,
				typ:  ObjectTypeBlob,
			})

		} else {
			subTreeOID, err := cmdWriteTree(fullPath)
			if err != nil {
				return "", err
			}
			if subTreeOID != "" {
				treeEntries = append(treeEntries, treeEntry{
					name: entry.Name(),
					oid:  subTreeOID,
					typ:  ObjectTypeTree,
				})
			}
		}
	}

	if len(treeEntries) == 0 {
		return "", fmt.Errorf("no objects found in %s", pathDir)
	}

	var treeContent strings.Builder
	for _, entry := range treeEntries {
		fmt.Fprintf(&treeContent, "%s %s %s\n", entry.typ, entry.oid, entry.name)
	}

	treeData := []byte(treeContent.String())
	treeOID, err := hashBytesAsObject(ObjectTypeTree, treeData)
	check(err)

	return treeOID, nil
}

func isIgnored(path string) bool {
	segments := strings.Split(path, "/")

	for _, segment := range segments {
		if segment == ".mini-git" || segment == ".git" {
			return true
		}
	}
	return false
}

func commit(message string) (string, error) {
	treeOid, err := cmdWriteTree(".")
	if err != nil {
		return "", err
	}

	commitMsg := fmt.Sprintf("tree %s\n", treeOid)
	commitMsg += "\n"
	commitMsg += message

	commitData := []byte(commitMsg)
	commitOid, ere := hashBytesAsObject(ObjectTypeCommit, commitData)
	check(ere)
	return commitOid, nil
}
