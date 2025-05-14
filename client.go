package mbclient

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// MBClient config
type MBClient struct {
	IP      string
	Port    int
	Timeout time.Duration
	Conn    net.Conn
}

// state show for error
const (
	Init        = "Init"
	ModbusError = "ModbusError"
	Ok          = "Ok"
	Disconnect  = "Disconnect"
	NoResponse  = "NoResponse"
)

// NewClient creates a new Modbus Client config.
func NewClient(IP string, port int, timeout time.Duration) *MBClient {
	m := &MBClient{}
	m.IP = IP
	m.Port = port
	m.Timeout = timeout

	return m
}

// Open modbus tcp connetion
func (m *MBClient) Open() error {
	addr := m.IP + ":" + strconv.Itoa(m.Port)
	// var err error
	conn, err := net.DialTimeout("tcp", addr, m.Timeout)
	if err != nil {
		return fmt.Errorf(Disconnect)
	}
	m.Conn = conn

	return nil
}

// Close modbus tcp connetion
func (m *MBClient) Close() {
	if m.Conn != nil {
		m.Conn.Close()
	}
}

// IsConnected for check modbus connetection
func (m *MBClient) IsConnected() bool {
	return m.Conn != nil
}

// Query make a modbus tcp query
func Query(conn net.Conn, timeout time.Duration, pdu []byte) ([]byte, error) {
	if conn == nil {
		return []byte{}, fmt.Errorf(Disconnect)
	}
	header := []byte{0, 0, 0, 0, byte(len(pdu) >> 8), byte(len(pdu))}
	wbuf := append(header, pdu...)
	//write
	_, err := conn.Write([]byte(wbuf))
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf(Disconnect)
	}

	//read
	rbuf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(timeout))
	leng, err := conn.Read(rbuf)
	if err != nil {
		fmt.Println(err)
		if strings.Contains(err.Error(), "i/o timeout") {
			return nil, fmt.Errorf(NoResponse)
		}
		return nil, fmt.Errorf(Disconnect)
	}
	if err := checkException(rbuf[:leng]); err != nil {
		return rbuf[:leng], err
	}
	if leng < 10 {
		return rbuf[:leng], fmt.Errorf(ModbusError)
	}

	return rbuf[6:leng], nil
}

// ReadCoil mdbus function 1 query and return []uint16
func (m *MBClient) ReadCoil(id uint8, addr uint16, leng uint16) ([]bool, error) {
	pdu := []byte{id, 0x01, byte(addr >> 8), byte(addr), byte(leng >> 8), byte(leng)}

	res, err := Query(m.Conn, m.Timeout, pdu)
	if err != nil {
		if err.Error() == Disconnect {
			m.Close()
			m.Conn = nil
		}
		return []bool{}, err
	}

	//check
	if int(res[2]) != (len(res) - 3) {
		fmt.Println(res)
		return []bool{}, fmt.Errorf("data length not match")
	}
	l := leng / 8
	if leng%8 != 0 {
		l += 1
	}
	if int(res[2]) != int(l) {
		fmt.Println(res)
		return []bool{}, fmt.Errorf("data length not match")
	}

	//convert
	result := []bool{}
	bc := res[2]
	for i := 0; i < int(bc); i++ {
		for j := 0; j < 8; j++ {
			if (res[3+i] & (byte(1) << byte(j))) != 0 {
				result = append(result, true)
			} else {
				result = append(result, false)
			}
		}
	}
	result = result[:leng]
	return result, nil
}

// ReadCoilIn mdbus function 2 query and return []uint16
func (m *MBClient) ReadCoilIn(id uint8, addr uint16, leng uint16) ([]bool, error) {

	pdu := []byte{id, 0x02, byte(addr >> 8), byte(addr), byte(leng >> 8), byte(leng)}

	//write
	res, err := Query(m.Conn, m.Timeout, pdu)
	if err != nil {
		if err.Error() == Disconnect {
			m.Close()
			m.Conn = nil
		}
		return []bool{}, err
	}

	//check
	if int(res[2]) != (len(res) - 3) {
		fmt.Println(res)
		return []bool{}, fmt.Errorf("data length not match")
	}
	l := leng / 8
	if leng%8 != 0 {
		l += 1
	}
	if int(res[2]) != int(l) {
		fmt.Println(res)
		return []bool{}, fmt.Errorf("data length not match")
	}

	//convert
	result := []bool{}
	bc := res[2]
	for i := 0; i < int(bc); i++ {
		for j := 0; j < 8; j++ {
			if (res[3+i] & (byte(1) << byte(j))) != 0 {
				result = append(result, true)
			} else {
				result = append(result, false)
			}
		}
	}
	result = result[:leng]
	return result, nil
}

// ReadReg mdbus function 3 query and return []uint16
func (m *MBClient) ReadReg(id uint8, addr uint16, leng uint16) ([]uint16, error) {

	pdu := []byte{id, 0x03, byte(addr >> 8), byte(addr), byte(leng >> 8), byte(leng)}

	//write
	res, err := Query(m.Conn, m.Timeout, pdu)
	if err != nil {
		if err.Error() == Disconnect {
			m.Close()
			m.Conn = nil
		}
		return []uint16{}, err
	}

	//check
	if (int(leng*2) != (len(res) - 3)) || int(leng*2) != int(res[2]) {
		fmt.Println(res)
		return []uint16{}, fmt.Errorf("data length not match")
	}

	//convert
	result := []uint16{}
	for i := 0; i < int(leng); i++ {
		var b uint16
		b = uint16(res[i*2+3]) << 8
		b |= uint16(res[i*2+4])
		result = append(result, b)
	}

	return result, nil
}

// ReadRegIn mdbus function 4 query and return []uint16
func (m *MBClient) ReadRegIn(id uint8, addr uint16, leng uint16) ([]uint16, error) {

	pdu := []byte{id, 0x04, byte(addr >> 8), byte(addr), byte(leng >> 8), byte(leng)}

	//write
	res, err := Query(m.Conn, m.Timeout, pdu)
	if err != nil {
		if err.Error() == Disconnect {
			m.Close()
			m.Conn = nil
		}
		return []uint16{}, err
	}

	//check
	if (int(leng*2) != (len(res) - 3)) || int(leng*2) != int(res[2]) {
		fmt.Println(res)
		return []uint16{}, fmt.Errorf("data length not match")
	}

	//convert
	result := []uint16{}
	for i := 0; i < int(leng); i++ {
		var b uint16
		b = uint16(res[i*2+3]) << 8
		b |= uint16(res[i*2+4])
		result = append(result, b)
	}

	return result, nil
}

// WriteCoil mdbus function 5 query and return []uint16
func (m *MBClient) WriteCoil(id uint8, addr uint16, data bool) error {

	var pdu = []byte{}
	if data {
		pdu = []byte{id, 0x5, byte(addr >> 8), byte(addr), 0xff, 0x00}
	} else {
		pdu = []byte{id, 0x5, byte(addr >> 8), byte(addr), 0x00, 0x00}
	}

	//write
	_, err := Query(m.Conn, m.Timeout, pdu)
	if err != nil {
		if err.Error() == Disconnect {
			m.Close()
			m.Conn = nil
		}
		return err
	}

	return nil
}

// WriteReg mdbus function 6 query and return []uint16
func (m *MBClient) WriteReg(id uint8, addr uint16, data uint16) error {

	pdu := []byte{id, 0x06, byte(addr >> 8), byte(addr), byte(data >> 8), byte(data)}

	//write
	_, err := Query(m.Conn, m.Timeout, pdu)
	if err != nil {
		if err.Error() == Disconnect {
			m.Close()
			m.Conn = nil
		}
		return err
	}

	return nil
}

// WriteCoils mdbus function 15(0x0f) query and return []uint16
func (m *MBClient) WriteCoils(id uint8, addr uint16, data []bool) error {
	var pdu []byte
	if len(data)%8 == 0 {
		pdu = []byte{id, 0x0f, byte(addr >> 8), byte(addr), byte(len(data) >> 8), byte(len(data)), byte(len(data) / 8)}
	} else {
		pdu = []byte{id, 0x0f, byte(addr >> 8), byte(addr), byte(len(data) >> 8), byte(len(data)), byte(len(data)/8) + 1}
	}
	var tbuf byte
	for i := 0; i < len(data); i++ {
		if data[i] {
			tbuf |= byte(1 << uint(i%8))
		}

		if (i+1)%8 == 0 || i == len(data)-1 {
			pdu = append(pdu, tbuf)
			tbuf = 0
		}
	}

	//write
	_, err := Query(m.Conn, m.Timeout, pdu)
	if err != nil {
		if err.Error() == Disconnect {
			m.Close()
			m.Conn = nil
		}
		return err
	}

	return nil
}

// WriteRegs mdbus function 16(0x10) query and return []uint16
func (m *MBClient) WriteRegs(id uint8, addr uint16, data []uint16) error {

	pdu := []byte{id, 0x10, byte(addr >> 8), byte(addr), byte(len(data) >> 8), byte(len(data)), byte(len(data)) * 2}

	for i := 0; i < len(data); i++ {
		pdu = append(pdu, byte(data[i]>>8))
		pdu = append(pdu, byte(data[i]))
	}

	//write
	_, err := Query(m.Conn, m.Timeout, pdu)
	if err != nil {
		if err.Error() == Disconnect {
			m.Close()
			m.Conn = nil
		}
		return err
	}

	return nil
}
