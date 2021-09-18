package qqwry

import (
	"encoding/binary"
	"log"
	"sync"

	"github.com/yinheli/mahonia"
	// "encoding/hex"
	"net"
	"os"
)

const (
	INDEX_LEN       = 7
	REDIRECT_MODE_1 = 0x01
	REDIRECT_MODE_2 = 0x02
)

var file *os.File
var lock sync.Mutex

// @author yinheli
type QQwry struct {
	Ip      string
	Country string
	City    string
}

func NewQQwry(filepath string) *QQwry {
	if filepath == "" {
		log.Fatalln("file path is null")
	}
	fileData, err := os.OpenFile(filepath, os.O_RDONLY, 0400)
	if err != nil {
		log.Fatalln(err.Error())
	}
	file = fileData
	qqwry := &QQwry{}
	return qqwry
}

func (this *QQwry) Find(ip string) *QQwry {
	lock.Lock()
	this.Ip = ip
	offset := this.searchIndex(binary.BigEndian.Uint32(net.ParseIP(ip).To4()))
	// log.Println("loc offset:", offset)
	if offset <= 0 {
		return nil
	}

	var country []byte
	var area []byte

	mode := this.readMode(offset + 4)
	// log.Println("mode", mode)
	if mode == REDIRECT_MODE_1 {
		countryOffset := this.readUInt24()
		mode = this.readMode(countryOffset)
		// log.Println("1 - mode", mode)
		if mode == REDIRECT_MODE_2 {
			c := this.readUInt24()
			country = this.readString(c)
			countryOffset += 4
		} else {
			country = this.readString(countryOffset)
			countryOffset += uint32(len(country) + 1)
		}
		area = this.readArea(countryOffset)
	} else if mode == REDIRECT_MODE_2 {
		countryOffset := this.readUInt24()
		country = this.readString(countryOffset)
		area = this.readArea(offset + 8)
	} else {
		country = this.readString(offset + 4)
		area = this.readArea(offset + uint32(5+len(country)))
	}

	enc := mahonia.NewDecoder("gbk")
	this.Country = enc.ConvertString(string(country))
	this.City = enc.ConvertString(string(area))
	lock.Unlock()
	return this
}

func (this *QQwry) readMode(offset uint32) byte {
	file.Seek(int64(offset), 0)
	mode := make([]byte, 1)
	file.Read(mode)
	return mode[0]
}

func (this *QQwry) readArea(offset uint32) []byte {
	mode := this.readMode(offset)
	if mode == REDIRECT_MODE_1 || mode == REDIRECT_MODE_2 {
		areaOffset := this.readUInt24()
		if areaOffset == 0 {
			return []byte("")
		} else {
			return this.readString(areaOffset)
		}
	} else {
		return this.readString(offset)
	}
	return []byte("")
}

func (this *QQwry) readString(offset uint32) []byte {
	file.Seek(int64(offset), 0)
	data := make([]byte, 0, 30)
	buf := make([]byte, 1)
	for {
		file.Read(buf)
		if buf[0] == 0 {
			break
		}
		data = append(data, buf[0])
	}
	return data
}

func (this *QQwry) searchIndex(ip uint32) uint32 {
	header := make([]byte, 8)
	file.Seek(0, 0)
	file.Read(header)

	start := binary.LittleEndian.Uint32(header[:4])
	end := binary.LittleEndian.Uint32(header[4:])

	// log.Printf("len info %v, %v ---- %v, %v", start, end, hex.EncodeToString(header[:4]), hex.EncodeToString(header[4:]))

	for {
		mid := this.getMiddleOffset(start, end)
		file.Seek(int64(mid), 0)
		buf := make([]byte, INDEX_LEN)
		file.Read(buf)
		_ip := binary.LittleEndian.Uint32(buf[:4])

		// log.Printf(">> %v, %v, %v -- %v", start, mid, end, hex.EncodeToString(buf[:4]))

		if end-start == INDEX_LEN {
			offset := byte3ToUInt32(buf[4:])
			file.Read(buf)
			if ip < binary.LittleEndian.Uint32(buf[:4]) {
				return offset
			} else {
				return 0
			}
		}

		// 找到的比较大，向前移
		if _ip > ip {
			end = mid
		} else if _ip < ip { // 找到的比较小，向后移
			start = mid
		} else if _ip == ip {
			return byte3ToUInt32(buf[4:])
		}

	}
	return 0
}

func (this *QQwry) readUInt24() uint32 {
	buf := make([]byte, 3)
	file.Read(buf)
	return byte3ToUInt32(buf)
}

func (this *QQwry) getMiddleOffset(start uint32, end uint32) uint32 {
	records := ((end - start) / INDEX_LEN) >> 1
	return start + records*INDEX_LEN
}

func byte3ToUInt32(data []byte) uint32 {
	i := uint32(data[0]) & 0xff
	i |= (uint32(data[1]) << 8) & 0xff00
	i |= (uint32(data[2]) << 16) & 0xff0000
	return i
}
