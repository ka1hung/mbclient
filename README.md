# mbclient
simple modbus tcp client tool

# Quickly start

## install

    go get https://github.com/ka1hung/mbclient

## example

    package main

    import (
        "log"

        "github.com/ka1hung/mbclient"
    )

    func main() {
        mbc := mbclient.NewClient("127.0.0.1", 502, 1)
        err := mbc.Open()
        if err != nil {
            panic(err)
        }
        defer mbc.Close()

        data, err := mbc.ReadReg(1, 0, 4)
        log.Println(data)

        mbc.WriteReg(1, 0, 1)
        mbc.WriteRegs(1, 1, []uint16{2, 3, 4})

        data, err = mbc.ReadReg(1, 0, 4)
        log.Println(data)
    }
