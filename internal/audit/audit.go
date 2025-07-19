package audit

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"time"

	"github.com/TFK70/kube-apiserver-audit-exporter/internal/logging"
	"github.com/fsnotify/fsnotify"

	"github.com/sirupsen/logrus"
)

type AuditReader struct {
	Path   string
	Events chan Event

	watcher *fsnotify.Watcher
	file    *os.File

	logger *logrus.Entry
}

type AuditReaderOptions struct {
	Path string
}

type Option func(*AuditReaderOptions)

func WithPath(p string) Option {
	return func(o *AuditReaderOptions) {
		o.Path = path.Join(p)
	}
}

func NewReader(opts ...Option) (*AuditReader, error) {
	options := &AuditReaderOptions{}

	for _, opt := range opts {
		opt(options)
	}

	logger, err := logging.GetNamedLogger("audit.go")
	if err != nil {
		return nil, err
	}

	return &AuditReader{
		Path:   options.Path,
		Events: make(chan Event, 3),

		logger: logger,
	}, nil
}

func (r *AuditReader) readNewLines(file *os.File, offset int64) (int64, error) {
	if _, err := file.Seek(offset, 0); err != nil {
		return offset, err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		var event Event
		err := json.Unmarshal([]byte(line), &event)
		if err != nil {
			return offset, err
		}

		r.Events <- event

		offset += int64(len(line)) + 1
	}

	if err := scanner.Err(); err != nil {
		return offset, err
	}

	return offset, nil
}

func (r *AuditReader) Start() error {
	var err error
	r.file, err = os.Open(r.Path)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}

	offset, err := r.file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("failed to seek to end: %v", err)
	}

	r.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %v", err)
	}

	var mu sync.Mutex

	go func() {
		mu.Lock()
		newOffset, err := r.readNewLines(r.file, offset)
		if err != nil {
			r.logger.Errorf("error reading new lines: %v", err)
		}
		offset = newOffset
		mu.Unlock()

		for {
			select {
			case event, ok := <-r.watcher.Events:
				if !ok {
					return
				}

				r.logger.Debugf("Received event: %+v", event)

				if event.Op&fsnotify.Write == fsnotify.Write && event.Name == r.Path {
					r.logger.Debugf("Write event")
					mu.Lock()
					newOffset, err := r.readNewLines(r.file, offset)
					if err != nil {
						r.logger.Errorf("error reading new lines: %v", err)
					} else {
						offset = newOffset
					}
					mu.Unlock()
				}

				if event.Op&(fsnotify.Rename|fsnotify.Remove) != 0 && event.Name == r.Path {
					r.logger.Info("file renamed or removed, reopening...")

					r.file.Close()

					for {
						f, err := os.Open(r.Path)
						if err != nil {
							r.logger.Infof("waiting to reopen file: %v", err)
							time.Sleep(time.Second)
							continue
						}

						r.file = f

						mu.Lock()
						offset, err = r.file.Seek(0, io.SeekEnd)
						if err != nil {
							r.logger.Errorf("failed to seek to end of new file: %v", err)
						}
						mu.Unlock()

						r.watcher.Add(r.Path)

						break
					}
				}

			case err, ok := <-r.watcher.Errors:
				if !ok {
					return
				}
				r.logger.Errorf("watcher error: %v", err)
			}
		}
	}()

	if err := r.watcher.Add(r.Path); err != nil {
		r.logger.Errorf("failed to watch file: %v", err)
	}

	return nil
}

func (r *AuditReader) Stop() error {
	close(r.Events)
	err := r.watcher.Close()
	if err != nil {
		return fmt.Errorf("failed to close watcher: %v", err)
	}

	err = r.file.Close()
	if err != nil {
		return fmt.Errorf("failed to close file: %v", err)
	}

	return nil
}
