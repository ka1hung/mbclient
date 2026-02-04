# mbclient

A simple Modbus TCP client library for Go.

## Installation

```bash
go get -u github.com/ka1hung/mbclient
```

## Supported Modbus Functions

| Function | Code | Method |
|----------|------|--------|
| Read Coils | 0x01 | `ReadCoil(id, addr, length)` |
| Read Discrete Inputs | 0x02 | `ReadCoilIn(id, addr, length)` |
| Read Holding Registers | 0x03 | `ReadReg(id, addr, length)` |
| Read Input Registers | 0x04 | `ReadRegIn(id, addr, length)` |
| Write Single Coil | 0x05 | `WriteCoil(id, addr, value)` |
| Write Single Register | 0x06 | `WriteReg(id, addr, value)` |
| Write Multiple Coils | 0x0F | `WriteCoils(id, addr, values)` |
| Write Multiple Registers | 0x10 | `WriteRegs(id, addr, values)` |

## Quick Start

```go
package main

import (
    "log"
    "time"

    "github.com/ka1hung/mbclient"
)

func main() {
    // Create client with IP, port, and timeout
    client := mbclient.NewClient("127.0.0.1", 502, time.Second)

    // Open connection
    if err := client.Open(); err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Read holding registers (function 3)
    values, err := client.ReadReg(1, 0, 10)
    if err != nil {
        log.Fatal(err)
    }
    log.Println("Registers:", values)
}
```

## API Examples

### Reading

```go
// Read Coils (function 1)
coils, err := client.ReadCoil(1, 0, 10)

// Read Discrete Inputs (function 2)
inputs, err := client.ReadCoilIn(1, 0, 10)

// Read Holding Registers (function 3)
holdingRegs, err := client.ReadReg(1, 0, 10)

// Read Input Registers (function 4)
inputRegs, err := client.ReadRegIn(1, 0, 10)
```

### Writing

```go
// Write Single Coil (function 5)
err := client.WriteCoil(1, 0, true)

// Write Single Register (function 6)
err := client.WriteReg(1, 0, 1234)

// Write Multiple Coils (function 15)
err := client.WriteCoils(1, 0, []bool{true, false, true})

// Write Multiple Registers (function 16)
err := client.WriteRegs(1, 0, []uint16{100, 200, 300})
```

## Error Handling

The library provides sentinel errors for common error conditions:

```go
import "errors"

values, err := client.ReadReg(1, 0, 10)
if err != nil {
    switch {
    case errors.Is(err, mbclient.ErrDisconnect):
        log.Println("Connection lost")
    case errors.Is(err, mbclient.ErrNoResponse):
        log.Println("Device not responding")
    case errors.Is(err, mbclient.ErrModbusError):
        log.Println("Modbus exception:", err)
    default:
        log.Println("Error:", err)
    }
}
```

## License

MIT License
