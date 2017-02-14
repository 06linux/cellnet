package socket

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"

	"github.com/davyxu/cellnet"
)

const (
	PackageHeaderSize = 8 // MsgID(uint32) + Ser(uint16) + Size(uint16)
)

type PacketStream struct {
	recvser      uint16
	sendser      uint16
	conn         net.Conn
	sendtagGuard sync.RWMutex

	outputWriter     *bufio.Writer
	outputHeadBuffer *bytes.Buffer

	inputHeadBuffer []byte
	headReader      *bytes.Reader

	maxPacketSize int
}

var (
	packageTagNotMatch     = errors.New("ReadPacket: package tag not match")
	packageDataSizeInvalid = errors.New("ReadPacket: package crack, invalid size")
	packageTooBig          = errors.New("ReadPacket: package too big")
)

// 从socket读取1个封包,并返回
func (self *PacketStream) Read(ev *cellnet.SessionEvent) (err error) {

	if _, err = self.headReader.Seek(0, 0); err != nil {
		return err
	}

	if _, err = io.ReadFull(self.conn, self.inputHeadBuffer); err != nil {
		return err
	}

	// 读取ID
	if err = binary.Read(self.headReader, binary.LittleEndian, &ev.MsgID); err != nil {
		return err
	}

	// 读取序号
	var ser uint16
	if err = binary.Read(self.headReader, binary.LittleEndian, &ser); err != nil {
		return err
	}

	// 读取整包大小
	var fullsize uint16
	if err = binary.Read(self.headReader, binary.LittleEndian, &fullsize); err != nil {
		return err
	}

	// 封包太大
	if self.maxPacketSize > 0 && int(fullsize) > self.maxPacketSize {
		return packageTooBig
	}

	// 序列号不匹配
	if self.recvser != ser {
		return packageTagNotMatch
	}

	dataSize := fullsize - PackageHeaderSize
	if dataSize < 0 {
		return packageDataSizeInvalid
	}

	// 读取数据
	msgBytes := make([]byte, dataSize)
	if _, err = io.ReadFull(self.conn, msgBytes); err != nil {
		return err
	}

	ev.Data = msgBytes

	// 增加序列号值
	self.recvser++

	return
}

// 将一个封包发送到socket
func (self *PacketStream) Write(ev *cellnet.SessionEvent) (err error) {

	// 防止将Send放在go内造成的多线程冲突问题
	self.sendtagGuard.Lock()
	defer self.sendtagGuard.Unlock()

	self.outputHeadBuffer.Reset()

	// 写ID
	if err = binary.Write(self.outputHeadBuffer, binary.LittleEndian, ev.MsgID); err != nil {
		return err
	}

	// 写序号
	if err = binary.Write(self.outputHeadBuffer, binary.LittleEndian, self.sendser); err != nil {
		return err
	}

	// 写包大小
	if err = binary.Write(self.outputHeadBuffer, binary.LittleEndian, uint16(len(ev.Data)+PackageHeaderSize)); err != nil {
		return err
	}

	// 发包头
	if err = self.writeFull(self.outputHeadBuffer.Bytes()); err != nil {
		return err
	}

	// 发包内容
	if err = self.writeFull(ev.Data); err != nil {
		return err
	}

	// 增加序号值
	self.sendser++

	return
}

// 完整发送所有封包
func (self *PacketStream) writeFull(p []byte) error {

	sizeToWrite := len(p)

	for {

		n, err := self.outputWriter.Write(p)

		if err != nil {
			return err
		}

		if n >= sizeToWrite {
			break
		}

		copy(p[0:sizeToWrite-n], p[n:sizeToWrite])
		sizeToWrite -= n
	}

	return nil

}

const sendTotalTryCount = 100

func (self *PacketStream) Flush() error {

	var err error
	for tryTimes := 0; tryTimes < sendTotalTryCount; tryTimes++ {

		err = self.outputWriter.Flush()

		// 如果没写完, flush底层会将没发完的buff准备好, 我们只需要重新调一次flush
		if err != io.ErrShortWrite {
			break
		}
	}

	return err
}

// 关闭
func (self *PacketStream) Close() error {
	return self.conn.Close()
}

// 裸socket操作
func (self *PacketStream) Raw() net.Conn {
	return self.conn
}

// 封包流 relay模式: 在封包头有clientid信息
func NewPacketStream(conn net.Conn) *PacketStream {

	s := &PacketStream{
		conn:             conn,
		recvser:          1,
		sendser:          1,
		outputWriter:     bufio.NewWriter(conn),
		outputHeadBuffer: bytes.NewBuffer([]byte{}),
		inputHeadBuffer:  make([]byte, PackageHeaderSize),
	}
	s.headReader = bytes.NewReader(s.inputHeadBuffer)

	return s
}
