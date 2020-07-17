// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/go-ocf/cloud/grpc-gateway/pb/errdetails/errorDetails.proto

package errdetails

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Content struct {
	Data                 []byte   `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	ContentType          string   `protobuf:"bytes,2,opt,name=content_type,json=contentType,proto3" json:"content_type,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Content) Reset()         { *m = Content{} }
func (m *Content) String() string { return proto.CompactTextString(m) }
func (*Content) ProtoMessage()    {}
func (*Content) Descriptor() ([]byte, []int) {
	return fileDescriptor_db67118f83db2314, []int{0}
}

func (m *Content) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Content.Unmarshal(m, b)
}
func (m *Content) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Content.Marshal(b, m, deterministic)
}
func (m *Content) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Content.Merge(m, src)
}
func (m *Content) XXX_Size() int {
	return xxx_messageInfo_Content.Size(m)
}
func (m *Content) XXX_DiscardUnknown() {
	xxx_messageInfo_Content.DiscardUnknown(m)
}

var xxx_messageInfo_Content proto.InternalMessageInfo

func (m *Content) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *Content) GetContentType() string {
	if m != nil {
		return m.ContentType
	}
	return ""
}

type DeviceError struct {
	Content              *Content `protobuf:"bytes,2,opt,name=content,proto3" json:"content,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeviceError) Reset()         { *m = DeviceError{} }
func (m *DeviceError) String() string { return proto.CompactTextString(m) }
func (*DeviceError) ProtoMessage()    {}
func (*DeviceError) Descriptor() ([]byte, []int) {
	return fileDescriptor_db67118f83db2314, []int{1}
}

func (m *DeviceError) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeviceError.Unmarshal(m, b)
}
func (m *DeviceError) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeviceError.Marshal(b, m, deterministic)
}
func (m *DeviceError) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeviceError.Merge(m, src)
}
func (m *DeviceError) XXX_Size() int {
	return xxx_messageInfo_DeviceError.Size(m)
}
func (m *DeviceError) XXX_DiscardUnknown() {
	xxx_messageInfo_DeviceError.DiscardUnknown(m)
}

var xxx_messageInfo_DeviceError proto.InternalMessageInfo

func (m *DeviceError) GetContent() *Content {
	if m != nil {
		return m.Content
	}
	return nil
}

func init() {
	proto.RegisterType((*Content)(nil), "ocf.cloud.grpcgateway.pb.errdetails.Content")
	proto.RegisterType((*DeviceError)(nil), "ocf.cloud.grpcgateway.pb.errdetails.DeviceError")
}

func init() {
	proto.RegisterFile("github.com/go-ocf/cloud/grpc-gateway/pb/errdetails/errorDetails.proto", fileDescriptor_db67118f83db2314)
}

var fileDescriptor_db67118f83db2314 = []byte{
	// 207 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0x72, 0x4d, 0xcf, 0x2c, 0xc9,
	0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x4f, 0xcf, 0xd7, 0xcd, 0x4f, 0x4e, 0xd3, 0x4f, 0xce,
	0xc9, 0x2f, 0x4d, 0xd1, 0x4f, 0x2f, 0x2a, 0x48, 0xd6, 0x4d, 0x4f, 0x2c, 0x49, 0x2d, 0x4f, 0xac,
	0xd4, 0x2f, 0x48, 0xd2, 0x4f, 0x2d, 0x2a, 0x4a, 0x49, 0x2d, 0x49, 0xcc, 0xcc, 0x29, 0x06, 0x31,
	0xf3, 0x8b, 0x5c, 0x20, 0x1c, 0xbd, 0x82, 0xa2, 0xfc, 0x92, 0x7c, 0x21, 0x65, 0xa0, 0x46, 0x3d,
	0xb0, 0x46, 0x3d, 0x90, 0x46, 0xa8, 0x3e, 0xbd, 0x82, 0x24, 0x3d, 0x84, 0x3e, 0x25, 0x07, 0x2e,
	0x76, 0xe7, 0xfc, 0xbc, 0x92, 0xd4, 0xbc, 0x12, 0x21, 0x21, 0x2e, 0x96, 0x94, 0xc4, 0x92, 0x44,
	0x09, 0x46, 0x05, 0x46, 0x0d, 0x9e, 0x20, 0x30, 0x5b, 0x48, 0x91, 0x8b, 0x27, 0x19, 0x22, 0x1d,
	0x5f, 0x52, 0x59, 0x90, 0x2a, 0xc1, 0x04, 0x94, 0xe3, 0x0c, 0xe2, 0x86, 0x8a, 0x85, 0x00, 0x85,
	0x94, 0x42, 0xb9, 0xb8, 0x5d, 0x52, 0xcb, 0x32, 0x93, 0x53, 0x5d, 0x41, 0x4e, 0x10, 0x72, 0xe3,
	0x62, 0x87, 0xca, 0x82, 0x15, 0x73, 0x1b, 0xe9, 0xe8, 0x11, 0xe1, 0x0e, 0x3d, 0xa8, 0x23, 0x82,
	0x60, 0x9a, 0x9d, 0xec, 0xa3, 0x6c, 0x49, 0x0f, 0x06, 0x6b, 0x04, 0x33, 0x89, 0x0d, 0x1c, 0x0a,
	0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0x9b, 0xa6, 0xf5, 0x6e, 0x4e, 0x01, 0x00, 0x00,
}
