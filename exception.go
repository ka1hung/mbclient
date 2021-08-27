package mbclient

import "fmt"

// ref:https://www.simplymodbus.ca/exceptions.htm
func checkException(data []byte) error {
	if len(data) < 9 {
		return fmt.Errorf(ModbusError + "data length too short")
	}
	//ModbusError
	if (data[7] & 0x80) != 0 {
		return fmt.Errorf(ModbusError + exception(data[8]))
	}
	return nil
}

func exception(code byte) string {
	switch code {
	case 1:
		return fmt.Sprintf("(%02X illegal function)", code)
	case 2:
		return fmt.Sprintf("(%02X illegal data address)", code)
	case 3:
		return fmt.Sprintf("(%02X illegal data value)", code)
	case 4:
		return fmt.Sprintf("(%02X slave device failure)", code)
	case 5:
		return fmt.Sprintf("(%02X acknowledge)", code)
	case 6:
		return fmt.Sprintf("(%02X slave device busy)", code)
	case 7:
		return fmt.Sprintf("(%02X negative acknowledge)", code)
	case 8:
		return fmt.Sprintf("(%02X memory parity error)", code)
	case 10:
		return fmt.Sprintf("(%02X gateway path unavailable)", code)
	case 11:
		return fmt.Sprintf("(%02X gateway target device failed to respond)", code)

	}
	return fmt.Sprintf("(%02X error code not in list)", code)
}
