package dm

import (
	"strconv"
	"sync/atomic"
	"time"

	"github.com/patrickmn/go-cache"
)

// MsgCacheEntry 消息缓存条目
type MsgCacheEntry struct {
	msgType MsgType    // 消息类型
	devID   int        // 设备id
	err     chan error // 用于wait通道
	done    uint32     // 表示消息接收到应答
}

// cacheInit 缓存初始化
func (sf *Client) cacheInit() {
	if sf.workOnWho == WorkOnHTTP {
		return
	}
	sf.msgCache = cache.New(sf.cacheExpiration, sf.cacheCleanupInterval)
	sf.pool = newPool()
	sf.msgCache.OnEvicted(func(id string, v interface{}) { // 超时处理
		entry := v.(*MsgCacheEntry)
		if atomic.LoadUint32(&entry.done) == 0 {
			if err := sf.eventProc.EvtRequestWaitResponseTimeout(sf, entry.msgType, entry.devID); err != nil {
				sf.warnf("ipc send message cache timeout failed, %+v", err)
			}
		}
		sf.debugf("cache evicted - @%s", id)
		sf.pool.Put(entry)
	})
	// TODO: 删除时放回Pool
	// sf.msgCache.OnDeleted(func(s string, v interface{}) {
	//	sf.pool.Put(v.(*MsgCacheEntry))
	// })
}

// CacheInsert 缓存插入指定ID
func (sf *Client) CacheInsert(id uint, devID int, msgType MsgType) {
	if sf.workOnWho == WorkOnHTTP {
		return
	}
	entry := sf.pool.Get()
	entry.devID = devID
	entry.msgType = msgType
	entry.done = 0
	sf.msgCache.SetDefault(strconv.FormatUint(uint64(id), 10), entry)
	sf.debugf("cache insert - @%d", id)
}

// CacheGet 获取缓存消息设备ID
func (sf *Client) CacheGet(id uint) (int, bool) {
	if sf.workOnWho == WorkOnHTTP {
		return 0, false
	}
	v, ok := sf.msgCache.Get(strconv.FormatUint(uint64(id), 10))
	if ok {
		return v.(*MsgCacheEntry).devID, true
	}
	return 0, false
}

// CacheWait 等待缓存ID的消息收到回复
func (sf *Client) CacheWait(id uint, t ...time.Duration) error {
	if sf.workOnWho == WorkOnHTTP {
		return ErrNotSupportWork
	}
	v, ok := sf.msgCache.Get(strconv.FormatUint(uint64(id), 10))
	if !ok {
		return ErrNotFound
	}

	entry := v.(*MsgCacheEntry)

	tm := 10 * time.Second
	if len(t) > 0 {
		tm = t[0]
	}

	tk := time.NewTicker(tm)
	defer tk.Stop()
	sf.debugf("cache wait - @%d", id)
	select {
	case v := <-entry.err:
		return v
	case <-tk.C:
	}
	return ErrWaitMessageTimeout
}

// CacheDone 指定缓存id收到回复,并发出同步通知
func (sf *Client) CacheDone(id uint, err error) {
	if sf.workOnWho == WorkOnHTTP {
		return
	}

	v, ok := sf.msgCache.Get(strconv.FormatUint(uint64(id), 10))
	if !ok {
		return
	}

	sf.debugf("cache done - @%d", id)
	entry := v.(*MsgCacheEntry)
	atomic.StoreUint32(&entry.done, 1)
	select {
	case entry.err <- err:
	default:
	}
}
