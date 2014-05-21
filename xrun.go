package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/howeyc/fsnotify"
)

var (
	runName      string
	appName      string
	buildLock    sync.Mutex
	buildProcess *os.Process
	exePath      string
	lastModified map[string]time.Time
	timeLock     sync.RWMutex
)

func isModified(f os.FileInfo) bool {
	timeLock.Lock()
	defer timeLock.Unlock()
	if t, ok := lastModified[f.Name()]; ok {
		if t.Equal(f.ModTime()) {
			return false
		}
	}
	lastModified[f.Name()] = f.ModTime()
	return true
}

type conf struct {
	excludeDirs  map[string]bool
	excludeFiles map[string]bool
	includeFiles map[string]bool
}

// a messages queue to put
var (
	webPort     string = "53126"
	messages           = make(chan string, 1000)
	curPath     string
	config      conf
	defaultConf = `{
	"excludeDirs": {
		".git":true,
		".svn":true
	},
	"excludeFiles": {
	},
	"includeFiles": {
	}
}
`
)

func loadConfig() error {
	f, err := os.Open("xrun.json")
	if err != nil {
		// Use default.
		err = json.Unmarshal([]byte(defaultConf), &config)
		if err != nil {
			return err
		}
		return nil
	}
	defer f.Close()

	Info("Loaded xrun.json")
	return json.NewDecoder(f).Decode(&config)
}

func build() error {
	buildLock.Lock()
	defer buildLock.Unlock()

	if buildProcess != nil {
		buildProcess.Kill()
		buildProcess.Wait()
	}

	Info("开始编译", appName)
	args := []string{"build"}
	args = append(args, "-o", appName)
	if len(os.Args) > 1 {
		args = append(args, os.Args[1:]...)
	}

	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	attr := &os.ProcAttr{
		Dir:   curPath,
		Env:   os.Environ(),
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	}

	Info("开始运行")
	buildProcess, err = os.StartProcess(exePath, []string{runName}, attr)
	if err != nil {
		return err
	}

	return nil
}

func scan() {
	for {
		var b string
		_, err := fmt.Scanf("%s\n", &b)
		if err != nil {
			Error(err)
			continue
		}
		//fmt.Println("===command===", string(b))
		if strings.ToLower(b) == "q" {
			buildLock.Lock()
			defer buildLock.Unlock()

			if buildProcess != nil {
				buildProcess.Kill()
				buildProcess.Wait()
			}
			os.Exit(0)
		}
	}
}

func main() {
	err := loadConfig()
	if err != nil {
		Error("load config error:", err)
		return
	}

	curPath, _ = os.Getwd()
	appName = path.Base(curPath)
	runName = appName
	if runtime.GOOS == "windows" {
		runName = runName + ".exe"
	}
	exePath = filepath.Join(curPath, runName)

	err = build()
	if err != nil {
		Error(err)
		return
	}

	os.Setenv("XRUN_DEBUG", "1")
	os.Setenv("XRUN_WEB_PORT", webPort)
	os.Setenv("XRUN_APP_PATH", exePath)
	os.Setenv("XRUN_SRC_PATH", curPath)

	Info("Start web interface")
	go web()

	//Info("Start accept commands")
	//go scan()

	Info("Start moniter")
	moniter(curPath)
}

// moniter go files
func moniter(rootDir string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	done := make(chan bool)

	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				if ev == nil {
					break
				}

				d, err := os.Stat(ev.Name)
				if err != nil {
					break
				}

				relativePath := ev.Name[len(curPath)+1:]
				if d.IsDir() {
					if _, ok := config.excludeDirs[relativePath]; ok {
						break
					}
				} else {
					if strings.HasSuffix(ev.Name, ".go") {
						if _, ok := config.excludeFiles[relativePath]; ok {
							break
						}
					} else {
						if _, ok := config.includeFiles[relativePath]; !ok {
							break
						}
					}

					if !isModified(d) {
						break
					}
				}

				fmt.Println("====", ev.Name, relativePath)

				if ev.IsCreate() {
					if d.IsDir() {
						watcher.Watch(ev.Name)
					} else {
						Infof("loaded %v", ev.Name)
						err = build()
						if err != nil {
							Errorf("load %v failed: %v", ev.Name, err)
							break
						}
					}
				} else if ev.IsDelete() {
					if d.IsDir() {
						watcher.RemoveWatch(ev.Name)
					} else {
						tmpl := ev.Name[len(rootDir)+1:]
						Info("deleted %v", tmpl)
						err = build()
						if err != nil {
							Errorf("remove %v failed: %v", ev.Name, err)
							break
						}
					}
				} else if ev.IsModify() {
					if d.IsDir() {
					} else {
						tmpl := ev.Name[len(rootDir)+1:]
						err = build()
						if err != nil {
							Errorf("reloaded %v failed: %v", tmpl, err)
							break
						}

						Info("reloaded %v", tmpl)
					}
				} else if ev.IsRename() {
					if d.IsDir() {
						watcher.RemoveWatch(ev.Name)
					} else {
						tmpl := ev.Name[len(rootDir)+1:]
						err = build()
						if err != nil {
							Errorf("reloaded %v failed: %v", tmpl, err)
							break
						}
					}
				}
			case err := <-watcher.Error:
				Errorf("watch error: %v", err)
			}
		}
	}()

	err = filepath.Walk(rootDir, func(f string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return watcher.Watch(f)
		}
		return nil
	})

	if err != nil {
		Error(err.Error())
		return err
	}

	<-done

	watcher.Close()
	return nil
}
