package fnet

import (
	"myDota/iface"
	"bytes"
	"encoding/binary"
	"myDota/utils"
	"errors"
)

var MAX_PACKET_SIZE = 1024*1024

type DataPkg struct{
	Length 	uint32
	MsgId 	uint32
	Data 	[]byte
}

func NewDataPkg(msgId uint32, data []byte) *DataPkg{
	return &DataPkg{
		Length:uint32(len(data)),
		MsgId:msgId,
		Data:data,
	}
}

func (dp *DataPkg) GetLen() uint32{
	return dp.Length
}

func (dp *DataPkg) GetMsgId () uint32{
	return dp.MsgId
}

func (dp *DataPkg) GetData() []byte{
	return dp.Data
}

func (dp *DataPkg) SetLen(len uint32){
	dp.Length =len
}

func (dp *DataPkg) SetMsgId (msgId uint32){
	dp.MsgId = msgId
}

func (dp *DataPkg) SetData(data []byte){
	dp.Data = data
}


type DataPack struct{}

func NewDataPack() *DataPack{
	return &DataPack{}
}

func (dp *DataPack) GetHeadLen() uint32{
	return 8
}

func (dp *DataPack) Pack(msg iface.IMsg) ([]byte, error){
	buff := bytes.NewBuffer([]byte{})

	if err := binary.Write(buff, binary.LittleEndian, msg.GetLen()); err != nil{
		return nil, err
	}

	if err := binary.Write(buff, binary.LittleEndian, msg.GetMsgId()); err!= nil{
		return nil, err
	}

	if err:= binary.Write(buff, binary.LittleEndian, msg.GetData()); err!= nil{
		return nil, err
	}

	return buff.Bytes(), nil
}


func (dp *DataPack) Unpack(headdata []byte)(iface.IMsg, error){
	buf := bytes.NewReader(headdata)
	headPkg := &DataPkg{}

	if err := binary.Read(buf, binary.LittleEndian, &headPkg.Length); err!= nil{
		return nil, err
	}

	if err := binary.Read(buf, binary.LittleEndian, &headPkg.MsgId); err!= nil{
		return nil, err
	}

	if utils.GlobalObject.MaxPacketSize > 0 && headPkg.Length > utils.GlobalObject.MaxPacketSize || utils.GlobalObject.MaxPacketSize == 0 && headPkg.Length > uint32(MAX_PACKET_SIZE){
		return nil, errors.New("Too many data to received!")
	}

	return headPkg, nil
}









































