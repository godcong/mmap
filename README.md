# MMAP

[![GitHub release](https://img.shields.io/github/release/godcong/mmap.svg)](https://github.com/godcong/mmap/releases)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/godcong/mmap)
[![codecov](https://codecov.io/gh/godcong/mmap/branch/main/graph/badge.svg)](https://codecov.io/gh/godcong/mmap)
[![GoDoc](https://godoc.org/github.com/godcong/mmap?status.svg)](http://godoc.org/github.com/godcong/mmap)
[![License](https://img.shields.io/github/license/godcong/mmap.svg)](https://github.com/godcong/mmap/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/godcong/mmap)](https://goreportcard.com/report/github.com/godcong/mmap)

The `MMAP` package provides a safe and efficient syscall interface for memory mapping operations, supporting both file mapping and shared memory.

## Features

- **File Memory Mapping**: Memory-mapped files with read/write access
- **Shared Memory**: Cross-process shared memory support
- **Cross-Platform**: Supports Darwin (macOS), Linux, and Windows
- **Standard Interfaces**: Implements Go's `io.Reader`, `io.Writer`, `io.ReaderAt`, `io.WriterAt`, `io.Seeker`, and `io.Closer` interfaces
- **High Performance**: Optimized for high-throughput data operations

## Installation

```bash
go get github.com/godcong/mmap@latest
```

## Requirements

- Go 1.23 or later
- Supported platforms: Windows, Linux, macOS (Darwin)

## Quick Start

### File Memory Mapping

```go
package main

import (
    "fmt"
    "log"
    "github.com/godcong/mmap"
)

func main() {
    // Read-only file mapping
    file, err := mmap.Open("example.txt")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    // Read data
    data := make([]byte, 1024)
    n, err := file.Read(data)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Read %d bytes: %s\n", n, string(data[:n]))
}
```

### Read-Write File Mapping

```go
package main

import (
    "log"
    "os"
    "github.com/godcong/mmap"
)

func main() {
    // Open file for reading and writing
    file, err := mmap.OpenFile("data.bin", os.O_RDWR|os.O_CREATE, 0644)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    // Write data at specific position
    _, err = file.WriteAt([]byte("Hello, World!"), 0)
    if err != nil {
        log.Fatal(err)
    }

    // Read data back
    buf := make([]byte, 13)
    _, err = file.ReadAt(buf, 0)
    if err != nil {
        log.Fatal(err)
    }
}
```

### Shared Memory

```go
package main

import (
    "fmt"
    "log"
    "github.com/godcong/mmap"
)

func main() {
    // Create shared memory (writer)
    writer, err := mmap.OpenMem(mmap.MapMemKeyInvalid, 1024)
    if err != nil {
        log.Fatal(err)
    }
    defer writer.Close()

    // Write data to shared memory
    _, err = writer.Write([]byte("Shared data"))
    if err != nil {
        log.Fatal(err)
    }

    // Open shared memory for reading (using the ID from writer)
    reader, err := mmap.OpenMem(writer.ID(), 1024)
    if err != nil {
        log.Fatal(err)
    }
    defer reader.Close()

    // Read data from shared memory
    buf := make([]byte, 12)
    _, err = reader.ReadAt(buf, 0)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Shared memory content: %s\n", string(buf))
}
```

## API Reference

### File Mapping

#### `Open(filename string) (*MapFile, error)`
Opens the named file for read-only memory mapping.

#### `OpenFile(filename string, flag int, mode os.FileMode) (*MapFile, error)`
Opens the named file with specified flags and permissions for memory mapping.

#### `OpenFileS(filename string, flag int, mode os.FileMode, size int) (*MapFile, error)`
Similar to `OpenFile` but with explicit size specification.

### Shared Memory

#### `OpenMem(id int, size int) (*MapMem, error)`
Opens or creates shared memory with the specified ID and size.

#### `OpenMemS(id int) (*MapMem, error)`
Opens shared memory with system-defined size.

### Constants

- `MapMemKeyInvalid` (-1): Used to create new shared memory instances
- Standard file flags: `os.O_RDONLY`, `os.O_WRONLY`, `os.O_RDWR`, `os.O_CREATE`

### Interfaces

Both `MapFile` and `MapMem` implement:
- `io.Reader`
- `io.Writer` (when writable)
- `io.ReaderAt`
- `io.WriterAt` (when writable)
- `io.Seeker`
- `io.Closer`
- `io.ByteReader`
- `io.ByteWriter` (when writable)

## Examples

See the [`examples`](examples/) folder for more detailed usage examples including:
- Basic file mapping
- Read-write operations
- Shared memory communication
- Error handling

## Performance

This package is optimized for performance with benchmark results showing:
- Shared Memory: Up to 1.8 GB/s throughput
- File Mapping: Efficient zero-copy operations
- Low latency memory access

See the [Benchmark Results](#benchmark-results) section below for detailed performance data.

## Platform Support

| Platform | File Mapping | Shared Memory | Status |
|----------|-------------|---------------|---------|
| Windows  | ✅ | ✅ | Fully Supported |
| Linux    | ✅ | ✅ | Fully Supported |
| macOS    | ✅ | ✅ | Fully Supported |

## Platform-Specific Implementation

### Windows
- Uses `CreateFileMapping` and `MapViewOfFile` for file mapping
- Uses `CreateFileMapping` with `INVALID_HANDLE_VALUE` for shared memory
- Supports both anonymous and named shared memory

### Linux
- Uses `mmap()` system call for file mapping
- Uses `shm_open()` and `mmap()` for shared memory
- Follows POSIX standards

### macOS (Darwin)
- Uses `mmap()` system call for file mapping
- Uses `shm_open()` and `mmap()` for shared memory
- Compatible with BSD-style memory mapping

## Error Handling

The package defines specific error types for common scenarios:

```go
var (
    ErrInvalid     = errors.New("invalid argument")
    ErrBadFileDesc = errors.New("bad file descriptor")
    ErrClosed      = errors.New("file closed")
    ErrShortWrite  = errors.New("short write")
    EOF            = errors.New("end of file")
)
```

## Best Practices

### Memory Management
- Always call `Close()` to properly unmap memory
- Use `defer` statements to ensure cleanup
- Be mindful of memory limits when mapping large files

### Shared Memory Usage
- Use consistent sizes when opening the same shared memory ID
- Implement proper synchronization for concurrent access
- Handle the case where shared memory might not exist yet

### Performance Tips
- Use `ReadAt` and `WriteAt` for random access patterns
- For sequential access, use `Read` and `Write` methods
- Consider buffer sizes that align with system page size

## Advanced Usage

### Large File Handling
```go
// Map a portion of a large file
file, err := mmap.OpenFile("largefile.bin", os.O_RDONLY, 0)
if err != nil {
    log.Fatal(err)
}
defer file.Close()

// Seek to position and read
_, err = file.Seek(1024*1024, io.SeekStart) // 1MB offset
if err != nil {
    log.Fatal(err)
}

buf := make([]byte, 4096)
n, err := file.Read(buf)
```

### Multi-Process Communication
```go
// Process 1: Writer
writer, _ := mmap.OpenMem(mmap.MapMemKeyInvalid, 4096)
defer writer.Close()
id := writer.ID()

// Process 2: Reader (needs to obtain the ID)
reader, _ := mmap.OpenMem(id, 4096)
defer reader.Close()
```

## Troubleshooting

### Common Issues

1. **Permission Denied**: Ensure file permissions are adequate
2. **Memory Allocation Failed**: Check available system memory
3. **Invalid File Descriptor**: Verify file exists and is accessible
4. **Mapping Size Mismatch**: Ensure consistent sizes when reopening shared memory

### Debug Mode

Enable debug logging for troubleshooting:

```go
// Enable debug logging (import debug_on.go instead of debug.go)
import _ "github.com/godcong/mmap/debug_on"
```

## Similar Packages

- **golang.org/x/exp/mmap**: Experimental mmap package from Go team
- **github.com/riobard/go-mmap**: Alternative mmap implementation
- **launchpad.net/gommap**: Go memory mapping library
- **github.com/edsrzf/mmap-go**: Cross-platform mmap implementation

### Why Choose This Package?

- **Complete Feature Set**: Both file mapping and shared memory
- **Modern Go**: Uses Go 1.23+ features and best practices
- **Production Ready**: Extensive testing and benchmarking
- **Cross-Platform**: Consistent API across Windows, Linux, and macOS
- **Performance Optimized**: Efficient zero-copy operations

## Benchmark Results

The following benchmarks were conducted on Windows 11 with an Intel Core i7-12700H CPU:

### Shared Memory Performance

| Size (Bytes) | Operations/sec | Throughput (MB/s) | Latency (ns/op) |
|-------------|---------------|------------------|-----------------|
| 1,024       | 220,854       | 37.65            | 27,198          |
| 4,096       | 142,830       | 530.60           | 7,720           |
| 16,384      | 68,470        | 1,052.54         | 15,566          |
| 65,536      | 26,972        | 1,506.76         | 43,495          |
| 262,144     | 8,418         | 1,843.62         | 142,190         |
| 1,048,576   | 1,506         | 1,206.79         | 868,899         |

### TCP Local Communication Performance

| Size (Bytes) | Operations/sec | Throughput (MB/s) | Latency (ns/op) |
|-------------|---------------|------------------|-----------------|
| 1,024       | 0.2           | ~0.00            | ~4,949,659,100  |
| 4,096       | 100           | 0.08             | 49,502,445      |
| 16,384      | 100           | 0.33             | 49,498,112      |
| 65,536      | 0.2           | ~0.01            | ~4,950,349,400  |
| 262,144     | 0.2           | ~0.05            | ~4,950,514,800  |
| 1,048,576   | 0.2           | 0.21             | ~4,949,555,300  |

### Pipe Communication Performance

| Size (Bytes) | Operations/sec | Throughput (MB/s) | Latency (ns/op) |
|-------------|---------------|------------------|-----------------|
| 1,024       | 52,519        | 47.04            | 21,769          |
| 4,096       | 53,853        | 162.11           | 25,266          |
| 16,384      | 40,072        | 526.27           | 31,132          |
| 65,536      | 22,765        | 1,322.70         | 49,547          |
| 262,144     | 9,231         | 1,610.32         | 162,790         |
| 1,048,576   | 3,385         | 2,505.86         | 418,449         |

### Memory Copy Performance (Baseline)

| Size (Bytes) | Operations/sec | Throughput (MB/s) | Latency (ns/op) |
|-------------|---------------|------------------|-----------------|
| 1,024       | 100,000,000   | 86,907.92        | 11.78           |
| 4,096       | 33,314,269    | 120,091.04       | 34.11           |
| 16,384      | 11,565,373    | 155,608.58       | 105.3           |
| 65,536      | 823,875       | 41,620.35        | 1,575           |
| 262,144     | 200,730       | 46,547.26        | 5,632           |
| 1,048,576   | 20,812        | 19,646.77        | 53,371          |

### Concurrent Shared Memory Performance (4KB data)

| Concurrency | Operations/sec | Throughput (MB/s) | Latency (ns/op) |
|-------------|---------------|------------------|-----------------|
| 1           | 182,282       | 659.32           | 6,212           |
| 2           | 157,710       | 554.99           | 7,380           |
| 4           | 151,671       | 611.26           | 6,701           |
| 8           | 123,558       | 442.67           | 9,253           |
| 16          | 138,694       | 463.51           | 8,837           |

### Performance Analysis

1. **Shared Memory** provides excellent performance for medium to large data sizes, with peak throughput around 1.8 GB/s for 256KB data.
2. **Memory Copy** serves as the theoretical maximum baseline, achieving up to 155 GB/s throughput.
3. **Pipe Communication** shows good scalability and reaches 2.5 GB/s for large data transfers.
4. **TCP Local** has the highest overhead due to network stack processing, making it suitable for network scenarios but not optimized for local communication.
5. **Concurrent Access** shows diminishing returns beyond 4 concurrent threads due to memory bandwidth limitations.

### Running Benchmarks

```bash
# Run all benchmarks
go test -run=^$ -bench=. -benchtime=1s

# Run specific benchmark
go test -run=^$ -bench=BenchmarkSharedMemory -benchtime=1s

# Run with longer duration for more stable results
go test -run=^$ -bench=. -benchtime=10s
```

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/godcong/mmap.git
cd mmap

# Run tests
go test ./...

# Run benchmarks
go test -bench=. -benchtime=1s

# Run with race detection
go test -race ./...
```

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run specific test
go test -v -run TestSharedMemory

# Run benchmarks
go test -bench=BenchmarkSharedMemory -benchtime=5s

# Run coverage
go test -cover ./...
```

## Roadmap

- [ ] TCP transmits data through shared memory between threads
- [ ] More elegant shared memory cleanup mechanisms
- [ ] Additional platform support (BSD, AIX)
- [ ] Memory-mapped I/O for device files
- [ ] Stream-based shared memory API

## Memory Map Service(TODO) Flow
See [`ServiceMemMap`](docs/ServiceMemMap.mmd)

## Version History

See [CHANGELOG.md](CHANGELOG.md) for detailed version information.

### Recent Updates
- **v1.x.x**: Added comprehensive shared memory support
- **v0.x.x**: Initial file mapping implementation

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Citation

If you use this package in your research or production code, please cite:

```bibtex
@software{godcong_mmap,
  author = {godcong},
  title = {MMAP: High-performance memory mapping library for Go},
  year = {2024},
  publisher = {GitHub},
  journal = {GitHub repository},
  howpublished = {\url{https://github.com/godcong/mmap}}
}
```

## Support

- **Issues**: [GitHub Issues](https://github.com/godcong/mmap/issues)
- **Discussions**: [GitHub Discussions](https://github.com/godcong/mmap/discussions)
- **Documentation**: [Go.dev Reference](https://pkg.go.dev/github.com/godcong/mmap)