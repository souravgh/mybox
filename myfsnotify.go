package main

import (
	"flag"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"sync"

	"github.com/howeyc/fsnotify"
)

type recWatcher struct {
	//watcher []*fsnotify.Watcher
	mtx      sync.Mutex
	watchers map[string]*fsnotify.Watcher
}

func (r *recWatcher) watch(dir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	r.mtx.Lock()
	// Leave if watcher is already created.
	if _, ok := r.watchers[dir]; ok {
		r.mtx.Unlock()
		log.Println("Watcher already present for ", dir)
		return
	}
	r.watchers[dir] = watcher
	r.mtx.Unlock()
	err = watcher.Watch(dir)
	if err != nil {
		delete(r.watchers, dir)
		watcher.Close()
		//remove watcher
		return
	}
	go r.ProcessEvent(dir, watcher)

	r.scanDir(dir)
}

func (r *recWatcher) ProcessEvent(dir string, w *fsnotify.Watcher) {
Loop:
	for {
		select {
		case ev := <-w.Event:
			log.Println("event:", ev)
			if ev == nil {
				break Loop
			}
			r.Process(ev)
		case err := <-w.Error:
			log.Println("error:", err)
			break Loop
		}
	}
	log.Println("Quiting directory :", dir)
}

func (r *recWatcher) Process(e *fsnotify.FileEvent) {
	if e.IsCreate() {
		file, err := os.Open(e.Name)
		if err != nil {
			log.Fatal(err)
		}
		fi, err := file.Stat()
		if err != nil {
			log.Fatal(err)
		}
		if fi.IsDir() {
			r.watch(e.Name)
		} else if fi.Mode().IsRegular() {
			r.sync(e.Name)
		}
		return
	}

	if e.IsDelete() {
		if w, ok := r.watchers[e.Name]; ok {
			log.Println("Closing watcher =", e.Name)
			w.Close()
			delete(r.watchers, e.Name)
		}
		return
		//}
	}

}

func (r *recWatcher) scanDir(dir string) {
	_ = filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			log.Println("Found dir:", path)
			r.watch(path)
		} else if f.Mode().IsRegular() {
			r.sync(path)
		}
		return nil
	})
}

func (r *recWatcher) sync(file string) {
	log.Println("Sync file", file)
}

func main() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	directory := flag.String("dir", usr.HomeDir+"/IggyDrive", "Directory")

	r := &recWatcher{watchers: make(map[string]*fsnotify.Watcher)}

	r.watch(*directory)

	/*watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}*/

	done := make(chan bool)

	// Hang so program doesn't exit
	<-done

	/* ... do stuff ... */
	//watcher.Close()
}
