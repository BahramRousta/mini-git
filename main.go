package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const GitDir = ".mini-git"

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: mini-git <command> [args]")
	}

	switch os.Args[1] {
	case "init":
		Init()
	case "file":
		if len(os.Args) < 4 {
			log.Fatal("usage: mini-git file <file-path> <file-type>")
		}
		filePath := os.Args[2]
		objType := os.Args[3]
		HashObject(filePath, objType)
	case "cat":
		if len(os.Args) < 3 {
			log.Fatal("usage: mini-git cat file-path")
		}
		catFile(os.Args[2])
	case "tree":
		oid, err := writeTree(os.Args[2])
		if err != nil {
			log.Fatalf("Error writing tree: %v", err)
			return
		}
		fmt.Printf("Tree object: %s\n", oid)
	default:
		log.Fatal("unknown command")
	}
}

func check(e error) {
	if e != nil {
		log.Fatalf("Error: %v", e)
	}
}

func Init() {
	err := os.Mkdir(GitDir, 0755)
	check(err)
	path, err := os.Getwd()
	check(err)
	fmt.Println("Initialized mini-git repository in: ", path)
}

func HashObject(filaPath, objType string) string {
	header := []byte(objType + "\x00")
	data, err := os.ReadFile(filaPath)
	check(err)

	obj := append(header, data...)

	hasher := sha1.New()
	_, err = hasher.Write(obj)
	check(err)

	hash := hasher.Sum(nil)
	hashHex := hex.EncodeToString(hash)

	objectsDir := fmt.Sprintf("%s/objects/%s", GitDir, hashHex[:2])
	err = os.MkdirAll(objectsDir, 0755)
	check(err)

	hashFile := fmt.Sprintf("%s/%s", objectsDir, hashHex[2:])
	err = os.WriteFile(hashFile, obj, 0644)
	check(err)

	return hashHex
}

func catFile(oid string) {
	if len(oid) < 2 {
		log.Fatalf("Invalid object ID: %s", oid)
	}

	subDir := oid[:2]
	fileName := oid[2:]

	oidFilePath := fmt.Sprintf("%s/objects/%s/%s", GitDir, subDir, fileName)
	data, err := os.ReadFile(oidFilePath)
	check(err)
	parts := strings.SplitN(string(data), "\x00", 2)

	if len(parts) != 2 {
		fmt.Println(string(data))
	} else {
		fmt.Println(parts[1])
	}
}

func writeTree(directory string) (string, error) {
	pathDir := filepath.Clean(directory)

	dir, err := os.Open(pathDir)
	check(err)
	defer dir.Close()

	entries, err := dir.ReadDir(-1)
	check(err)

	type treeEntry struct {
		name string
		typ  string
		oid  string
	}
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
			oid := HashObject(fullPath, "blob")

			treeEntries = append(treeEntries, treeEntry{
				name: entry.Name(),
				oid:  oid,
				typ:  "blob",
			})

		} else {
			oid, err := writeTree(fullPath)
			if err != nil {
				return "", err
			}
			if oid != "" {
				treeEntries = append(treeEntries, treeEntry{
					name: entry.Name(),
					oid:  oid,
					typ:  "tree",
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

	treeBytes := []byte(treeContent.String())
	hasher := sha1.New()
	_, err = hasher.Write(append([]byte("tree\x00"), treeBytes...))
	check(err)
	hash := hasher.Sum(nil)
	oid := hex.EncodeToString(hash)

	objectsDir := filepath.Join(GitDir, "objects", oid[:2])
	err = os.MkdirAll(objectsDir, 0755)
	check(err)

	hashFile := filepath.Join(objectsDir, oid[2:])
	err = os.WriteFile(hashFile, append([]byte("tree\x00"), treeBytes...), 0644)
	check(err)

	return oid, nil
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
