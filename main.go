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
		if len(os.Args) < 3 {
			log.Fatal("usage: mini-git file <file-path>")
		}
		HashObject(os.Args[2])
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

func HashObject(filaPath string) {
	data, err := os.ReadFile(filaPath)
	check(err)

	hasher := sha1.New()
	_, err = hasher.Write(data)
	check(err)

	hash := hasher.Sum(nil)
	hashHex := hex.EncodeToString(hash)

	objectsDir := GitDir + "/objects"
	err = os.MkdirAll(objectsDir, 0755)
	check(err)

	hashFile := fmt.Sprintf("%s/%s", objectsDir, hashHex)
	err = os.WriteFile(hashFile, data, 0644)
	check(err)

	fmt.Printf("Hashed object %s: %s\n", filaPath, hashHex)
}
