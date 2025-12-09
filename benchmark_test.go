package mmap

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"time"
)

// BenchmarkSharedMemory 基准测试共享内存性能
func BenchmarkSharedMemory(b *testing.B) {
	sizes := []int{1024, 4096, 16384, 65536, 262144, 1048576} // 1KB to 1MB

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
			// 创建共享内存
			writer, err := OpenMem(MapMemKeyInvalid, size)
			if err != nil {
				b.Fatalf("Failed to create shared memory: %v", err)
			}
			defer writer.Close()

			// 准备测试数据
			data := make([]byte, size)
			for i := range data {
				data[i] = byte(i % 256)
			}

			b.ResetTimer()
			b.SetBytes(int64(size))

			for i := 0; i < b.N; i++ {
				// 写入数据
				n, err := writer.WriteAt(data, 0)
				if err != nil {
					b.Fatalf("Write failed: %v", err)
				}
				if n != size {
					b.Fatalf("Write length mismatch: expected %d, got %d", size, n)
				}

				// 读取数据验证
				reader, err := OpenMem(writer.ID(), size)
				if err != nil {
					b.Fatalf("Failed to open shared memory for reading: %v", err)
				}

				readData := make([]byte, size)
				_, err = reader.ReadAt(readData, 0)
				reader.Close()

				if err != nil {
					b.Fatalf("Read failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkTCPLocal 基准测试本地TCP通信性能
func BenchmarkTCPLocal(b *testing.B) {
	sizes := []int{1024, 4096, 16384, 65536, 262144, 1048576} // 1KB to 1MB

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
			// 设置本地TCP连接
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				b.Fatalf("Failed to listen: %v", err)
			}
			defer listener.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			var wg sync.WaitGroup
			wg.Add(1)
			serverErr := make(chan error, 1)

			// 启动服务器
			go func() {
				defer wg.Done()
				conn, err := listener.Accept()
				if err != nil {
					serverErr <- err
					return
				}
				defer conn.Close()

				// 设置读写超时
				conn.SetDeadline(time.Now().Add(5 * time.Second))

				buf := make([]byte, size)
				for {
					select {
					case <-ctx.Done():
						return
					default:
						// 确保读取完整的size字节
						totalRead := 0
						for totalRead < size {
							n, err := conn.Read(buf[totalRead:])
							if err != nil {
								return
							}
							totalRead += n
						}

						// 确保写入完整的size字节
						totalWritten := 0
						for totalWritten < size {
							n, err := conn.Write(buf[totalWritten:])
							if err != nil {
								return
							}
							totalWritten += n
						}
					}
				}
			}()

			// 客户端连接
			conn, err := net.Dial("tcp", listener.Addr().String())
			if err != nil {
				b.Fatalf("Failed to connect: %v", err)
			}
			defer conn.Close()

			// 设置读写超时
			conn.SetDeadline(time.Now().Add(5 * time.Second))

			// 准备测试数据
			data := make([]byte, size)
			for i := range data {
				data[i] = byte(i % 256)
			}

			// 等待服务器就绪
			time.Sleep(50 * time.Millisecond)
			select {
			case err := <-serverErr:
				b.Fatalf("Server error: %v", err)
			default:
			}

			b.ResetTimer()
			b.SetBytes(int64(size))

			for i := 0; i < b.N; i++ {
				// 更新超时时间
				conn.SetDeadline(time.Now().Add(1 * time.Second))

				// 确保写入完整的size字节
				totalWritten := 0
				for totalWritten < size {
					n, err := conn.Write(data[totalWritten:])
					if err != nil {
						b.Fatalf("Write failed: %v", err)
					}
					totalWritten += n
				}

				// 确保读取完整的size字节
				readData := make([]byte, size)
				totalRead := 0
				for totalRead < size {
					n, err := conn.Read(readData[totalRead:])
					if err != nil {
						b.Fatalf("Read failed: %v", err)
					}
					totalRead += n
				}
			}

			cancel()
			wg.Wait()
		})
	}
}

// BenchmarkPipe 基准测试管道通信性能
func BenchmarkPipe(b *testing.B) {
	sizes := []int{1024, 4096, 16384, 65536, 262144, 1048576} // 1KB to 1MB

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
			r, w, err := os.Pipe()
			if err != nil {
				b.Fatalf("Failed to create pipe: %v", err)
			}
			defer r.Close()
			defer w.Close()

			// 准备测试数据
			data := make([]byte, size)
			for i := range data {
				data[i] = byte(i % 256)
			}

			b.ResetTimer()
			b.SetBytes(int64(size))

			for i := 0; i < b.N; i++ {
				// 使用并发读写避免死锁
				var wg sync.WaitGroup
				var writeErr, readErr error
				var writeBytes, readBytes int

				wg.Add(2)

				// 写入goroutine
				go func() {
					defer wg.Done()
					writeBytes, writeErr = w.Write(data)
				}()

				// 读取goroutine
				go func() {
					defer wg.Done()
					readData := make([]byte, size)
					totalRead := 0
					for totalRead < size {
						n, err := r.Read(readData[totalRead:])
						if err != nil {
							readErr = err
							return
						}
						totalRead += n
					}
					readBytes = totalRead
				}()

				wg.Wait()

				if writeErr != nil {
					b.Fatalf("Write failed: %v", writeErr)
				}
				if writeBytes != size {
					b.Fatalf("Write length mismatch: expected %d, got %d", size, writeBytes)
				}
				if readErr != nil {
					b.Fatalf("Read failed: %v", readErr)
				}
				if readBytes != size {
					b.Fatalf("Read length mismatch: expected %d, got %d", size, readBytes)
				}
			}
		})
	}
}

// BenchmarkMemoryCopy 基准测试普通内存拷贝性能
func BenchmarkMemoryCopy(b *testing.B) {
	sizes := []int{1024, 4096, 16384, 65536, 262144, 1048576} // 1KB to 1MB

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
			// 准备测试数据
			src := make([]byte, size)
			dst := make([]byte, size)
			for i := range src {
				src[i] = byte(i % 256)
			}

			b.ResetTimer()
			b.SetBytes(int64(size))

			for i := 0; i < b.N; i++ {
				copy(dst, src)
			}
		})
	}
}

// BenchmarkConcurrentSharedMemory 并发共享内存性能测试
func BenchmarkConcurrentSharedMemory(b *testing.B) {
	concurrencyLevels := []int{1, 2, 4, 8, 16}
	size := 4096 // 4KB

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("concurrency-%d", concurrency), func(b *testing.B) {
			// 创建共享内存
			writer, err := OpenMem(MapMemKeyInvalid, size)
			if err != nil {
				b.Fatalf("Failed to create shared memory: %v", err)
			}
			defer writer.Close()

			// 准备测试数据
			data := make([]byte, size)
			for i := range data {
				data[i] = byte(i % 256)
			}

			b.ResetTimer()
			b.SetBytes(int64(size))

			var wg sync.WaitGroup
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < b.N/concurrency; j++ {
						// 写入数据
						_, err := writer.WriteAt(data, 0)
						if err != nil {
							b.Errorf("Write failed: %v", err)
							return
						}

						// 读取数据验证
						reader, err := OpenMem(writer.ID(), size)
						if err != nil {
							b.Errorf("Failed to open shared memory for reading: %v", err)
							return
						}

						readData := make([]byte, size)
						_, err = reader.ReadAt(readData, 0)
						reader.Close()

						if err != nil {
							b.Errorf("Read failed: %v", err)
							return
						}
					}
				}()
			}
			wg.Wait()
		})
	}
}

// TestPerformanceComparison 性能对比测试
func TestPerformanceComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	size := 4096 // 4KB
	iterations := 10000

	t.Run("SharedMemory", func(t *testing.T) {
		writer, err := OpenMem(MapMemKeyInvalid, size)
		if err != nil {
			t.Fatalf("Failed to create shared memory: %v", err)
		}
		defer writer.Close()

		data := make([]byte, size)
		for i := range data {
			data[i] = byte(i % 256)
		}

		start := time.Now()
		for i := 0; i < iterations; i++ {
			writer.WriteAt(data, 0)
			reader, _ := OpenMem(writer.ID(), size)
			readData := make([]byte, size)
			reader.ReadAt(readData, 0)
			reader.Close()
		}
		duration := time.Since(start)

		t.Logf("SharedMemory: %d operations in %v (%.2f ops/sec, %.2f ns/op)",
			iterations, duration,
			float64(iterations)/duration.Seconds(),
			float64(duration.Nanoseconds())/float64(iterations))
	})

	t.Run("MemoryCopy", func(t *testing.T) {
		src := make([]byte, size)
		dst := make([]byte, size)
		for i := range src {
			src[i] = byte(i % 256)
		}

		start := time.Now()
		for i := 0; i < iterations; i++ {
			copy(dst, src)
		}
		duration := time.Since(start)

		t.Logf("MemoryCopy: %d operations in %v (%.2f ops/sec, %.2f ns/op)",
			iterations, duration,
			float64(iterations)/duration.Seconds(),
			float64(duration.Nanoseconds())/float64(iterations))
	})
}
