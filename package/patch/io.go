// Copyright 2017 fatedier, fatedier@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package patch_io

// patch 实现暴露实时统计流量hook，实现长连接情况下，依然能够正常更新流量
// "github.com/fatedier/golib/io"

import (
	"io"
	"sync"
	"time"

	"github.com/golang/snappy"

	"github.com/fatedier/golib/crypto"
	"github.com/fatedier/golib/pool"
)

const defaultHookInterval = 5 * time.Second

// Join two io.ReadWriteCloser and do some operations.
func Join(c1 io.ReadWriteCloser, c2 io.ReadWriteCloser, inCountHook func(int64), outCountHook func(int64)) (inCount int64, outCount int64, errors []error) {
	var wait sync.WaitGroup
	recordErrs := make([]error, 2)
	pipe := func(number int, to io.ReadWriteCloser, from io.ReadWriteCloser, count *int64, pipeHook func(int64)) {
		defer wait.Done()
		defer to.Close()
		defer from.Close()

		buf := pool.GetBuf(16 * 1024)
		defer pool.PutBuf(buf)
		*count, recordErrs[number] = CopyBufferWithHook(to, from, buf, pipeHook, defaultHookInterval)
	}

	wait.Add(2)
	go pipe(0, c1, c2, &inCount, inCountHook)
	go pipe(1, c2, c1, &outCount, outCountHook)
	wait.Wait()

	for _, e := range recordErrs {
		if e != nil {
			errors = append(errors, e)
		}
	}
	return
}

func CopyBufferWithHook(dst io.Writer, src io.Reader, buf []byte, hook func(n int64), interval time.Duration) (total int64, err error) {
	var delta int64
	var lastTime time.Time

	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			written, writeErr := dst.Write(buf[:n])
			if writeErr != nil {
				return total, writeErr
			}
			total += int64(written)
			delta += int64(written)

			// 判断是否触发 hook
			if hook != nil {
				now := time.Now()
				if interval <= 0 || now.Sub(lastTime) >= interval {
					hook(delta)
					delta = 0
					lastTime = now
				}
			}
		}

		if readErr != nil {
			// 触发最后一次 hook
			if delta > 0 && hook != nil {
				hook(delta)
			}
			if readErr == io.EOF {
				return total, nil
			}
			return total, readErr
		}
	}
}

func WithEncryption(rwc io.ReadWriteCloser, key []byte) (io.ReadWriteCloser, error) {
	w, err := crypto.NewWriter(rwc, key)
	if err != nil {
		return nil, err
	}
	return WrapReadWriteCloser(crypto.NewReader(rwc, key), w, func() error {
		return rwc.Close()
	}), nil
}

func WithCompression(rwc io.ReadWriteCloser) io.ReadWriteCloser {
	sr := snappy.NewReader(rwc)
	sw := snappy.NewWriter(rwc)
	return WrapReadWriteCloser(sr, sw, func() error {
		_ = sw.Close()
		return rwc.Close()
	})
}

// WithCompressionFromPool will get snappy reader and writer from pool.
// You can recycle the snappy reader and writer by calling the returned recycle function, but it is not necessary.
func WithCompressionFromPool(rwc io.ReadWriteCloser) (out io.ReadWriteCloser, recycle func()) {
	sr := pool.GetSnappyReader(rwc)
	sw := pool.GetSnappyWriter(rwc)
	out = WrapReadWriteCloser(sr, sw, func() error {
		err := rwc.Close()
		return err
	})
	recycle = func() {
		pool.PutSnappyReader(sr)
		pool.PutSnappyWriter(sw)
	}
	return
}

type ReadWriteCloser struct {
	r       io.Reader
	w       io.Writer
	closeFn func() error

	closed bool
	mu     sync.Mutex
}

// closeFn will be called only once
func WrapReadWriteCloser(r io.Reader, w io.Writer, closeFn func() error) io.ReadWriteCloser {
	return &ReadWriteCloser{
		r:       r,
		w:       w,
		closeFn: closeFn,
		closed:  false,
	}
}

func (rwc *ReadWriteCloser) Read(p []byte) (n int, err error) {
	return rwc.r.Read(p)
}

func (rwc *ReadWriteCloser) Write(p []byte) (n int, err error) {
	return rwc.w.Write(p)
}

func (rwc *ReadWriteCloser) Close() error {
	rwc.mu.Lock()
	if rwc.closed {
		rwc.mu.Unlock()
		return nil
	}
	rwc.closed = true
	rwc.mu.Unlock()

	if rwc.closeFn != nil {
		return rwc.closeFn()
	}
	return nil
}
