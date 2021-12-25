package swift

import (
	"container/list"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"sort"
	"sync"

	"github.com/traPtitech/trap-collection-server/pkg/common"
)

const cacheMaxSize = 1024 * 1024 * 1024 * 5

type Cache struct {
	cacheLocker    sync.RWMutex
	cacheDirectory string
	cacheList      *list.List
	cacheItemMap   map[string]*list.Element
	cacheSize      int64
}

type cacheItem struct {
	name string
	size int64
}

func NewCache(cacheDirectory common.FilePath) (*Cache, error) {
	_, err := os.Stat(string(cacheDirectory))
	if errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(string(cacheDirectory), 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create cache directory: %w", err)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat cache directory: %w", err)
	}

	dirEntry, err := os.ReadDir(string(cacheDirectory))
	if err != nil {
		return nil, fmt.Errorf("failed to read cache directory: %w", err)
	}

	cacheList := list.New()
	cacheItemMap := make(map[string]*list.Element, len(dirEntry))
	cacheSize := int64(0)
	sort.Slice(dirEntry, func(i, j int) bool {
		iInfo, err := dirEntry[i].Info()
		if err != nil {
			return false
		}

		jInfo, err := dirEntry[j].Info()
		if err != nil {
			return false
		}

		return iInfo.ModTime().After(jInfo.ModTime())
	})
	for _, entry := range dirEntry {
		if cacheSize > cacheMaxSize {
			err := os.Remove(path.Join(string(cacheDirectory), entry.Name()))
			if err != nil {
				return nil, fmt.Errorf("failed to remove cache file: %w", err)
			}
			continue
		}

		fileInfo, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("failed to get file info: %w", err)
		}

		cacheSize += fileInfo.Size()
		cacheList.PushBack(cacheItem{
			name: entry.Name(),
			size: fileInfo.Size(),
		})
		cacheItemMap[entry.Name()] = cacheList.Back()
	}

	cache := &Cache{
		cacheLocker:    sync.RWMutex{},
		cacheDirectory: string(cacheDirectory),
		cacheList:      cacheList,
		cacheItemMap:   cacheItemMap,
	}

	return cache, nil
}

func (c *Cache) save(name string, r io.Reader) error {
	c.cacheLocker.Lock()
	defer c.cacheLocker.Unlock()

	f, err := os.Create(c.cacheFilePath(name))
	if err != nil {
		return fmt.Errorf("failed to create cache file: %w", err)
	}
	defer f.Close()

	n, err := io.Copy(f, r)
	if err != nil {
		return fmt.Errorf("failed to copy content: %w", err)
	}

	c.cacheSize += n

	c.cacheItemMap[name] = c.cacheList.PushFront(name)

	if c.cacheSize > cacheMaxSize {
		go func() {
			c.cacheLocker.Lock()
			defer c.cacheLocker.Unlock()
			for c.cacheSize > cacheMaxSize {
				item := c.cacheList.Back()
				c.cacheList.Remove(item)

				val, ok := item.Value.(cacheItem)
				if !ok {
					log.Printf("error: failed to cast cache item")
					break
				}

				delete(c.cacheItemMap, val.name)

				err := os.Remove(c.cacheFilePath(val.name))
				if err != nil {
					log.Printf("error: failed to remove cache file: %s", err)
					break
				}

				c.cacheSize -= val.size
			}
		}()
	}

	return nil
}

func (c *Cache) load(name string, w io.Writer) (bool, error) {
	c.cacheLocker.RLock()
	defer c.cacheLocker.RUnlock()
	if item, ok := c.cacheItemMap[name]; ok {
		f, err := os.Open(c.cacheFilePath(name))
		if err != nil {
			return false, fmt.Errorf("failed to get cache: %w", err)
		}
		defer f.Close()

		_, err = io.Copy(w, f)
		if err != nil {
			return false, fmt.Errorf("failed to copy cache: %w", err)
		}

		c.cacheList.MoveToFront(item)

		return true, nil
	}

	return false, nil
}

func (c *Cache) clean() error {
	c.cacheLocker.Lock()
	defer c.cacheLocker.Unlock()

	for name := range c.cacheItemMap {
		err := os.Remove(c.cacheFilePath(name))
		if err != nil {
			return fmt.Errorf("failed to remove cache file: %w", err)
		}
	}

	c.cacheList = list.New()
	c.cacheItemMap = make(map[string]*list.Element)
	c.cacheSize = 0

	return nil
}

func (c *Cache) cacheFilePath(name string) string {
	return path.Join(c.cacheDirectory, name)
}
