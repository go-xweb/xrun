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

	"github.com/go-xweb/log"
	"github.com/howeyc/fsnotify"
)

var (
	runName      string
	appName      string
	buildLock    sync.Mutex
	buildProcess *os.Process
	exePath      string
	version      = "0.1.0626"
)

func relativePath(p, root string) string {
	return p[len(root)+1:]
}

type cache struct {
	root  string
	lock  sync.RWMutex
	files map[string]os.FileInfo
}

func NewCache(rootDir string) *cache {
	return &cache{
		root:  rootDir,
		files: make(map[string]os.FileInfo),
	}
}

func (c *cache) get(p string) (os.FileInfo, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if d, ok := c.files[relativePath(p, c.root)]; ok {
		return d, nil
	}
	return nil, os.ErrNotExist
}

func (c *cache) add(p string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	d, err := os.Stat(p)
	if err != nil {
		return err
	}
	c.files[relativePath(p, c.root)] = d
	return nil
}

func (c *cache) isModified(p string) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	if t, ok := c.files[relativePath(p, c.root)]; ok {
		d, err := os.Stat(p)
		if err != nil {
			return false
		}

		return t.ModTime() != d.ModTime() || t.Size() != d.Size()
	}

	return true
}

func (c *cache) remove(p string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.files, relativePath(p, c.root))
}

type conf struct {
	Mode         int
	ExcludeDirs  map[string]bool
	ExcludeFiles map[string]bool
	IncludeFiles map[string]bool
	IncludeDirs  map[string]bool
}

// a messages queue to put
var (
	webPort     string = "53126"
	messages           = make(chan string, 1000)
	curPath     string
	config      conf
	defaultConf = fmt.Sprintf(`{
		"Mode":%v,
	"ExcludeDirs": {
		".git":true,
		".svn":true
	},
	"ExcludeFiles": {
	},
	"IncludeFiles": {
	},
	"IncludeDirs": {
	}
}
`, log.Linfo)
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

	log.Info("Loaded xrun.json")
	return json.NewDecoder(f).Decode(&config)
}

func build() error {
	buildLock.Lock()
	defer buildLock.Unlock()

	if buildProcess != nil {
		buildProcess.Kill()
		buildProcess.Wait()
	}

	log.Info("building", appName)
	args := []string{"build"}
	args = append(args, "-o", runName)
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

	log.Info("running...")
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
			log.Error(err)
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
	if len(os.Args) == 2 && strings.ToLower(os.Args[1]) == "--version" {
		fmt.Println("xrun version", version)
		return
	}

	log.SetPrefix("[xrun] ")

	err := loadConfig()
	if err != nil {
		log.Error("load config error:", err)
		return
	}

	log.SetOutputLevel(config.Mode)

	/*if len(config.ExcludeDirs) > 0 {
		dirs := config.ExcludeDirs
		config.ExcludeDirs = make(map[string]bool)
		for dir, v := range dirs {
			dirPath, _ := filepath.Abs(dir)
			dirPath = strings.Replace(dirPath, "\\", "/", -1)
			config.ExcludeDirs[dirPath] = v
		}
	}
	if len(config.ExcludeFiles) > 0 {
		dirs := config.ExcludeFiles
		config.ExcludeFiles = make(map[string]bool)
		for dir, v := range dirs {
			dirPath, _ := filepath.Abs(dir)
			dirPath = strings.Replace(dirPath, "\\", "/", -1)
			config.ExcludeFiles[dirPath] = v
		}
	}
	if len(config.IncludeFiles) > 0 {
		dirs := config.IncludeFiles
		config.IncludeFiles = make(map[string]bool)
		for dir, v := range dirs {
			dirPath, _ := filepath.Abs(dir)
			dirPath = strings.Replace(dirPath, "\\", "/", -1)
			config.IncludeFiles[dirPath] = v
		}
	}*/
	curPath, _ = os.Getwd()
	curPath = strings.Replace(curPath, "\\", "/", -1)
	appName = path.Base(curPath)
	runName = appName
	if runtime.GOOS == "windows" {
		runName = runName + ".exe"
	}
	exePath = filepath.Join(curPath, runName)

	err = build()
	if err != nil {
		log.Error(err)
		return
	}

	os.Setenv("XRUN_DEBUG", "1")
	os.Setenv("XRUN_WEB_PORT", webPort)
	os.Setenv("XRUN_APP_PATH", exePath)
	os.Setenv("XRUN_SRC_PATH", curPath)

	log.Info("Start web interface")
	go web()

	//Info("Start accept commands")
	//go scan()

	log.Info("Start moniter")
	moniter(curPath, config.IncludeDirs)
}

// moniter go files
func moniter(rootDir string, otherDirs map[string]bool) error {
	cache := NewCache(rootDir)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	done := make(chan bool)

	go func() {
		for {
		start:
			select {
			case ev := <-watcher.Event:
				if ev == nil {
					break
				}

				rPath := relativePath(ev.Name, rootDir)
				log.Debug("relativePath is", rPath, ev)
				for p, _ := range config.ExcludeDirs {
					log.Debug("cmp", rPath, p)
					if strings.HasPrefix(rPath, p) {
						goto start
					}
				}
				if rPath == runName {
					break
				}

				if ev.IsDelete() || ev.IsRename() {
					d, err := cache.get(ev.Name)
					if os.IsNotExist(err) {
						err = build()
						if err != nil {
							log.Errorf("remove %v failed: %v", ev.Name, err)
						}
						break
					}
					if err != nil {
						log.Error(err)
						break
					}

					if d.IsDir() {
						watcher.RemoveWatch(ev.Name)
						cache.remove(ev.Name)
					} else {
						cache.remove(ev.Name)
						log.Infof("deleted %v", ev.Name)
						err = build()
						if err != nil {
							log.Errorf("remove %v failed: %v", ev.Name, err)
							break
						}
					}
					break
				}

				var d os.FileInfo
				d, err = os.Stat(ev.Name)
				if err != nil {
					log.Errorf("file stat error:", err)
					break
				}

				if !d.IsDir() {
					log.Debug("ext file name is", filepath.Ext(ev.Name))
					if filepath.Ext(ev.Name) == ".go" {
						if _, ok := config.ExcludeFiles[rPath]; ok {
							break
						}
					} else {
						if _, ok := config.IncludeFiles[rPath]; !ok {
							break
						}
					}
				}

				log.Info("File or Dir is changed:", ev.Name)

				if ev.IsCreate() {
					if d.IsDir() {
						watcher.Watch(ev.Name)
					} else {
						if !cache.isModified(ev.Name) {
							log.Debug("File", ev.Name, "is not modified.")
							break
						}

						log.Infof("loaded %v", ev.Name)
						err = build()
						if err != nil {
							log.Errorf("after %v changed build failed: %v", ev.Name, err)
							break
						}
					}
				} else if ev.IsModify() {
					if d.IsDir() {
					} else {
						if !cache.isModified(ev.Name) {
							log.Debug("File", ev.Name, "is not modified.")
							break
						}

						err = build()
						if err != nil {
							log.Errorf("reloaded %v failed: %v", ev.Name, err)
							break
						}

						log.Infof("reloaded %v", ev.Name)
					}
				} else {
					log.Errorf("unknown event: %v", ev)
					break
				}
			case err := <-watcher.Error:
				log.Errorf("watch error: %v", err)
			}
		}
	}()

	fn := func(f string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return watcher.Watch(f)
		}
		return nil
	}

	err = filepath.Walk(rootDir, fn)

	if len(otherDirs) > 0 {
		for otherDir, v := range otherDirs {
			if !v {
				continue
			}
			absPath, _ := filepath.Abs(otherDir)
			absPath = strings.Replace(absPath, "\\", "/", -1)
			if strings.HasPrefix(absPath+"/", rootDir+"/") {
				continue
			}
			err = filepath.Walk(absPath, fn)
		}
	}

	if err != nil {
		log.Error(err.Error())
		return err
	}

	<-done

	watcher.Close()
	return nil
}
