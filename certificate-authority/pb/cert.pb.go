// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/plgd-dev/cloud/certificate-authority/pb/cert.proto

package pb

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type SignCertificateRequest struct {
	CertificateSigningRequest []byte `protobuf:"bytes,1,opt,name=certificate_signing_request,json=certificateSigningRequest,proto3" json:"certificate_signing_request,omitempty"`
}

func (m *SignCertificateRequest) Reset()         { *m = SignCertificateRequest{} }
func (m *SignCertificateRequest) String() string { return proto.CompactTextString(m) }
func (*SignCertificateRequest) ProtoMessage()    {}
func (*SignCertificateRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_08674a33113246ee, []int{0}
}
func (m *SignCertificateRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SignCertificateRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SignCertificateRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SignCertificateRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SignCertificateRequest.Merge(m, src)
}
func (m *SignCertificateRequest) XXX_Size() int {
	return m.Size()
}
func (m *SignCertificateRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SignCertificateRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SignCertificateRequest proto.InternalMessageInfo

func (m *SignCertificateRequest) GetCertificateSigningRequest() []byte {
	if m != nil {
		return m.CertificateSigningRequest
	}
	return nil
}

type SignCertificateResponse struct {
	Certificate []byte `protobuf:"bytes,1,opt,name=certificate,proto3" json:"certificate,omitempty"`
}

func (m *SignCertificateResponse) Reset()         { *m = SignCertificateResponse{} }
func (m *SignCertificateResponse) String() string { return proto.CompactTextString(m) }
func (*SignCertificateResponse) ProtoMessage()    {}
func (*SignCertificateResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_08674a33113246ee, []int{1}
}
func (m *SignCertificateResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SignCertificateResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SignCertificateResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SignCertificateResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SignCertificateResponse.Merge(m, src)
}
func (m *SignCertificateResponse) XXX_Size() int {
	return m.Size()
}
func (m *SignCertificateResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_SignCertificateResponse.DiscardUnknown(m)
}

var xxx_messageInfo_SignCertificateResponse proto.InternalMessageInfo

func (m *SignCertificateResponse) GetCertificate() []byte {
	if m != nil {
		return m.Certificate
	}
	return nil
}

func init() {
	proto.RegisterType((*SignCertificateRequest)(nil), "ocf.cloud.certificateauthority.pb.SignCertificateRequest")
	proto.RegisterType((*SignCertificateResponse)(nil), "ocf.cloud.certificateauthority.pb.SignCertificateResponse")
}

func init() {
	proto.RegisterFile("github.com/plgd-dev/cloud/certificate-authority/pb/cert.proto", fileDescriptor_08674a33113246ee)
}

var fileDescriptor_08674a33113246ee = []byte{
	// 214 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xb2, 0x4d, 0xcf, 0x2c, 0xc9,
	0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x2f, 0xc8, 0x49, 0x4f, 0xd1, 0x4d, 0x49, 0x2d, 0xd3,
	0x4f, 0xce, 0xc9, 0x2f, 0x4d, 0xd1, 0x4f, 0x4e, 0x2d, 0x2a, 0xc9, 0x4c, 0xcb, 0x4c, 0x4e, 0x2c,
	0x49, 0xd5, 0x4d, 0x2c, 0x2d, 0xc9, 0xc8, 0x2f, 0xca, 0x2c, 0xa9, 0xd4, 0x2f, 0x48, 0x02, 0x4b,
	0xe8, 0x15, 0x14, 0xe5, 0x97, 0xe4, 0x0b, 0x29, 0xe6, 0x27, 0xa7, 0xe9, 0x81, 0x95, 0xeb, 0x21,
	0x29, 0x87, 0xab, 0xd6, 0x2b, 0x48, 0x52, 0x8a, 0xe0, 0x12, 0x0b, 0xce, 0x4c, 0xcf, 0x73, 0x46,
	0x48, 0x07, 0xa5, 0x16, 0x96, 0xa6, 0x16, 0x97, 0x08, 0xd9, 0x71, 0x49, 0x23, 0x69, 0x8a, 0x2f,
	0xce, 0x4c, 0xcf, 0xcb, 0xcc, 0x4b, 0x8f, 0x2f, 0x82, 0x48, 0x4b, 0x30, 0x2a, 0x30, 0x6a, 0xf0,
	0x04, 0x49, 0x22, 0x29, 0x09, 0x86, 0xa8, 0x80, 0xea, 0x57, 0xb2, 0xe6, 0x12, 0xc7, 0x30, 0xb9,
	0xb8, 0x20, 0x3f, 0xaf, 0x38, 0x55, 0x48, 0x81, 0x8b, 0x1b, 0x49, 0x1f, 0xd4, 0x28, 0x64, 0x21,
	0x27, 0xff, 0x13, 0x8f, 0xe4, 0x18, 0x2f, 0x3c, 0x92, 0x63, 0x7c, 0xf0, 0x48, 0x8e, 0x71, 0xc2,
	0x63, 0x39, 0x86, 0x0b, 0x8f, 0xe5, 0x18, 0x6e, 0x3c, 0x96, 0x63, 0x88, 0x32, 0x25, 0x3d, 0x48,
	0xac, 0x0b, 0x92, 0x92, 0xd8, 0xc0, 0x21, 0x62, 0x0c, 0x08, 0x00, 0x00, 0xff, 0xff, 0x1f, 0xa3,
	0x68, 0xd0, 0x52, 0x01, 0x00, 0x00,
}

func (m *SignCertificateRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SignCertificateRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SignCertificateRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.CertificateSigningRequest) > 0 {
		i -= len(m.CertificateSigningRequest)
		copy(dAtA[i:], m.CertificateSigningRequest)
		i = encodeVarintCert(dAtA, i, uint64(len(m.CertificateSigningRequest)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *SignCertificateResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SignCertificateResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SignCertificateResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Certificate) > 0 {
		i -= len(m.Certificate)
		copy(dAtA[i:], m.Certificate)
		i = encodeVarintCert(dAtA, i, uint64(len(m.Certificate)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintCert(dAtA []byte, offset int, v uint64) int {
	offset -= sovCert(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *SignCertificateRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.CertificateSigningRequest)
	if l > 0 {
		n += 1 + l + sovCert(uint64(l))
	}
	return n
}

func (m *SignCertificateResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Certificate)
	if l > 0 {
		n += 1 + l + sovCert(uint64(l))
	}
	return n
}

func sovCert(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozCert(x uint64) (n int) {
	return sovCert(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *SignCertificateRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCert
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: SignCertificateRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SignCertificateRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CertificateSigningRequest", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCert
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthCert
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthCert
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.CertificateSigningRequest = append(m.CertificateSigningRequest[:0], dAtA[iNdEx:postIndex]...)
			if m.CertificateSigningRequest == nil {
				m.CertificateSigningRequest = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCert(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCert
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthCert
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *SignCertificateResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCert
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: SignCertificateResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SignCertificateResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Certificate", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCert
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthCert
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthCert
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Certificate = append(m.Certificate[:0], dAtA[iNdEx:postIndex]...)
			if m.Certificate == nil {
				m.Certificate = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCert(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCert
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthCert
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipCert(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowCert
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowCert
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowCert
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthCert
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupCert
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthCert
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthCert        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowCert          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupCert = fmt.Errorf("proto: unexpected end of group")
)
