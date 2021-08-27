# mbclient
simple modbus tcp client tool

# Quickly start

## install

    go get -u github.com/ka1hung/mbclient

## example

    package main

    import (
        "log"
        "time"

        "github.com/ka1hung/mbclient"
    )

    func main() {
        mbc := mbclient.NewClient("127.0.0.1", 502, time.Second)
        err := mbc.Open()
        if err != nil {
            panic(err)
        }
        defer mbc.Close()

        // read
        log.Println(mbc.ReadCoil(1, 0, 10))   //func1
        log.Println(mbc.ReadCoilIn(1, 0, 10)) //func2
        log.Println(mbc.ReadReg(1, 0, 10))    //func3
        log.Println(mbc.ReadRegIn(1, 0, 10))  //func4

        // write coil
        mbc.WriteCoil(1, 0, true)                      //func5
        mbc.WriteCoils(1, 1, []bool{true, true, true}) //func15
        log.Println(mbc.ReadReg(1, 0, 4))

        // write reg
        mbc.WriteReg(1, 0, 1)                  //func6
        mbc.WriteRegs(1, 1, []uint16{2, 3, 4}) //func16
        log.Println(mbc.ReadReg(1, 0, 4))
    }
