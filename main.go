package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"os"
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

func HashObject(filaPath, objType string) {
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

	fmt.Printf("Hashed object %s: %s\n", filaPath, hashHex)
}

func catFile(oid string) {
	subDir := oid[:2]
	fileName := oid[2:]

	oidFilePath := fmt.Sprintf("%s/objects/%s/%s", GitDir, subDir, fileName)
	data, err := os.ReadFile(oidFilePath)
	check(err)
	fmt.Println(string(data))
}
