package mbclient

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// MBClient is a Modbus TCP client configuration and connection holder.
type MBClient struct {
	IP      string
	Port    int
	Timeout time.Duration
	Conn    net.Conn
}

// Sentinel errors for Modbus operations.
var (
	ErrDisconnect  = errors.New("Disconnect")
	ErrNoResponse  = errors.New("NoResponse")
	ErrModbusError = errors.New("ModbusError")
)

// NewClient creates a new Modbus Client config.
func NewClient(IP string, port int, timeout time.Duration) *MBClient {
	m := &MBClient{}
	m.IP = IP
	m.Port = port
	m.Timeout = timeout

	return m
}

// Open establishes a Modbus TCP connection.
func (m *MBClient) Open() error {
	addr := m.IP + ":" + strconv.Itoa(m.Port)
	conn, err := net.DialTimeout("tcp", addr, m.Timeout)
	if err != nil {
		return Disconnect
	}
	m.Conn = conn

	return nil
}

// Close closes the Modbus TCP connection.
func (m *MBClient) Close() {
	if m.Conn != nil {
		m.Conn.Close()
	}
}

// IsConnected returns true if the connection object exists.
func (m *MBClient) IsConnected() bool {
	return m.Conn != nil
}

// handleDisconnect handles disconnect error and cleans up connection
func (m *MBClient) handleDisconnect(err error) {
	if errors.Is(err, Disconnect) {
		m.Close()
		m.Conn = nil
	}
}

// Query sends a Modbus TCP request and returns the response.
func Query(conn net.Conn, timeout time.Duration, pdu []byte, byteLen int) ([]byte, error) {
	if conn == nil {
		return []byte{}, Disconnect
	}
	header := []byte{0, 0, 0, 0, byte(len(pdu) >> 8), byte(len(pdu))}
	wbuf := append(header, pdu...)
	// write
	_, err := conn.Write(wbuf)
	if err != nil {
		return nil, Disconnect
	}

	// read
	rbs := []byte{}
	for i := range 10 {
		rbuf := make([]byte, 1024)
		conn.SetReadDeadline(time.Now().Add(timeout))
		leng, err := conn.Read(rbuf)
		if err != nil {
			if strings.Contains(err.Error(), "i/o timeout") {
				return nil, ErrNoResponse
			}
			return nil, Disconnect
		}

		if i == 0 {
			if err := checkException(rbuf[:leng]); err != nil {
				return rbuf[:leng], err
			}
		}
		rbs = append(rbs, rbuf[:leng]...)

		if len(rbs) >= byteLen {
			break
		}

	}

	if len(rbs) < 10 {
		return rbs, ErrModbusError
	}

	return rbs[6:], nil
}

// readCoilInternal is shared logic for Modbus function 1 (Read Coils) and 2 (Read Discrete Inputs).
func (m *MBClient) readCoilInternal(id uint8, addr uint16, leng uint16, funcCode byte) ([]bool, error) {
	pdu := []byte{id, funcCode, byte(addr >> 8), byte(addr), byte(leng >> 8), byte(leng)}

	byteLen := 9 + (int(leng)+7)/8
	res, err := Query(m.Conn, m.Timeout, pdu, byteLen)
	if err != nil {
		m.handleDisconnect(err)
		return []bool{}, err
	}

	// check
	expectedBytes := (leng + 7) / 8
	if int(res[2]) != (len(res)-3) || int(res[2]) != int(expectedBytes) {
		return []bool{}, fmt.Errorf("data length not match")
	}

	// convert
	result := []bool{}
	bc := res[2]
	for i := 0; i < int(bc); i++ {
		for j := 0; j < 8; j++ {
			result = append(result, (res[3+i]&(byte(1)<<byte(j))) != 0)
		}
	}
	return result[:leng], nil
}

// ReadCoil executes Modbus function 1 (Read Coils) and returns coil states.
func (m *MBClient) ReadCoil(id uint8, addr uint16, leng uint16) ([]bool, error) {
	return m.readCoilInternal(id, addr, leng, 0x01)
}

// ReadCoilIn executes Modbus function 2 (Read Discrete Inputs) and returns input states.
func (m *MBClient) ReadCoilIn(id uint8, addr uint16, leng uint16) ([]bool, error) {
	return m.readCoilInternal(id, addr, leng, 0x02)
}

// readRegInternal is shared logic for Modbus function 3 (Read Holding Registers) and 4 (Read Input Registers).
func (m *MBClient) readRegInternal(id uint8, addr uint16, leng uint16, funcCode byte) ([]uint16, error) {
	pdu := []byte{id, funcCode, byte(addr >> 8), byte(addr), byte(leng >> 8), byte(leng)}

	byteLen := 9 + int(leng*2)
	res, err := Query(m.Conn, m.Timeout, pdu, byteLen)
	if err != nil {
		m.handleDisconnect(err)
		return []uint16{}, err
	}

	// check
	if (int(leng*2) != (len(res) - 3)) || int(leng*2) != int(res[2]) {
		return []uint16{}, fmt.Errorf("data length not match")
	}

	// convert
	result := make([]uint16, leng)
	for i := 0; i < int(leng); i++ {
		result[i] = uint16(res[i*2+3])<<8 | uint16(res[i*2+4])
	}
	return result, nil
}

// ReadReg executes Modbus function 3 (Read Holding Registers) and returns register values.
func (m *MBClient) ReadReg(id uint8, addr uint16, leng uint16) ([]uint16, error) {
	return m.readRegInternal(id, addr, leng, 0x03)
}

// ReadRegIn executes Modbus function 4 (Read Input Registers) and returns register values.
func (m *MBClient) ReadRegIn(id uint8, addr uint16, leng uint16) ([]uint16, error) {
	return m.readRegInternal(id, addr, leng, 0x04)
}

// WriteCoil executes Modbus function 5 (Write Single Coil).
func (m *MBClient) WriteCoil(id uint8, addr uint16, data bool) error {
	pdu := []byte{id, 0x5, byte(addr >> 8), byte(addr), 0x00, 0x00}
	if data {
		pdu[4] = 0xff
	}

	// write

	byteLen := 10
	_, err := Query(m.Conn, m.Timeout, pdu, byteLen)
	if err != nil {
		m.handleDisconnect(err)
		return err
	}

	return nil
}

// WriteReg executes Modbus function 6 (Write Single Register).
func (m *MBClient) WriteReg(id uint8, addr uint16, data uint16) error {

	pdu := []byte{id, 0x06, byte(addr >> 8), byte(addr), byte(data >> 8), byte(data)}

	// write
	byteLen := 10
	_, err := Query(m.Conn, m.Timeout, pdu, byteLen)
	if err != nil {
		m.handleDisconnect(err)
		return err
	}

	return nil
}

// WriteCoils executes Modbus function 15 (Write Multiple Coils).
func (m *MBClient) WriteCoils(id uint8, addr uint16, data []bool) error {
	byteCount := (len(data) + 7) / 8
	pdu := []byte{id, 0x0f, byte(addr >> 8), byte(addr), byte(len(data) >> 8), byte(len(data)), byte(byteCount)}
	coilBytes := make([]byte, byteCount)
	for i, v := range data {
		if v {
			coilBytes[i/8] |= 1 << uint(i%8)
		}
	}
	pdu = append(pdu, coilBytes...)

	// write
	byteLen := 12
	_, err := Query(m.Conn, m.Timeout, pdu, byteLen)
	if err != nil {
		m.handleDisconnect(err)
		return err
	}

	return nil
}

// WriteRegs executes Modbus function 16 (Write Multiple Registers).
func (m *MBClient) WriteRegs(id uint8, addr uint16, data []uint16) error {
	byteCount := len(data) * 2
	pdu := []byte{id, 0x10, byte(addr >> 8), byte(addr), byte(len(data) >> 8), byte(len(data)), byte(byteCount)}
	regBytes := make([]byte, byteCount)
	for i, v := range data {
		regBytes[i*2] = byte(v >> 8)
		regBytes[i*2+1] = byte(v)
	}
	pdu = append(pdu, regBytes...)

	// write
	byteLen := 12
	_, err := Query(m.Conn, m.Timeout, pdu, byteLen)
	if err != nil {
		m.handleDisconnect(err)
		return err
	}

	return nil
}
