package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage())
		os.Exit(2)
	}
	cmd := os.Args[1]
	switch cmd {
	case "login":
		fs := flag.NewFlagSet("login", flag.ExitOnError)
		base := fs.String("base", "", "API base URL")
		_ = fs.Parse(os.Args[2:])
		if err := runLogin(*base); err != nil {
			fmt.Fprintln(os.Stderr, "错误:", err)
			os.Exit(1)
		}
	case "logout":
		if err := runLogout(); err != nil {
			fmt.Fprintln(os.Stderr, "错误:", err)
			os.Exit(1)
		}
	case "upload":
		fs := flag.NewFlagSet("upload", flag.ExitOnError)
		base := fs.String("base", "", "API base URL")
		token := fs.String("token", "", "session token")
		dir := fs.String("dir", "", "directory to scan")
		glob := fs.String("glob", "**/*.md", "glob under --dir")
		dryRun := fs.Bool("dry-run", false, "parse only")
		onExists := fs.String("on-exists", "ask", "ask|overwrite|skip|create")
		_ = fs.Parse(os.Args[2:])
		opt := uploadOptions{
			Base:     *base,
			Token:    *token,
			Dir:      *dir,
			Glob:     *glob,
			DryRun:   *dryRun,
			OnExists: *onExists,
			Files:    fs.Args(),
		}
		if err := runUpload(opt); err != nil {
			fmt.Fprintln(os.Stderr, "错误:", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		fmt.Println(usage())
	default:
		fmt.Fprintf(os.Stderr, "未知命令: %s\n\n%s\n", cmd, usage())
		os.Exit(2)
	}
}
