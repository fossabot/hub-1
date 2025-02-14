// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: github.com/plgd-dev/hub/pkg/net/grpc/stub.proto

package grpc_test

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type TestRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Test string `protobuf:"bytes,1,opt,name=test,proto3" json:"test,omitempty"`
}

func (x *TestRequest) Reset() {
	*x = TestRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TestRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TestRequest) ProtoMessage() {}

func (x *TestRequest) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TestRequest.ProtoReflect.Descriptor instead.
func (*TestRequest) Descriptor() ([]byte, []int) {
	return file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_rawDescGZIP(), []int{0}
}

func (x *TestRequest) GetTest() string {
	if x != nil {
		return x.Test
	}
	return ""
}

type TestResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Test string `protobuf:"bytes,1,opt,name=test,proto3" json:"test,omitempty"`
}

func (x *TestResponse) Reset() {
	*x = TestResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TestResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TestResponse) ProtoMessage() {}

func (x *TestResponse) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TestResponse.ProtoReflect.Descriptor instead.
func (*TestResponse) Descriptor() ([]byte, []int) {
	return file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_rawDescGZIP(), []int{1}
}

func (x *TestResponse) GetTest() string {
	if x != nil {
		return x.Test
	}
	return ""
}

var File_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto protoreflect.FileDescriptor

var file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_rawDesc = []byte{
	0x0a, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x6c, 0x67,
	0x64, 0x2d, 0x64, 0x65, 0x76, 0x2f, 0x68, 0x75, 0x62, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x6e, 0x65,
	0x74, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x73, 0x74, 0x75, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x13, 0x70, 0x6b, 0x67, 0x2e, 0x6e, 0x65, 0x74, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e,
	0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x22, 0x21, 0x0a, 0x0b, 0x54, 0x65, 0x73, 0x74, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x65, 0x73, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x65, 0x73, 0x74, 0x22, 0x22, 0x0a, 0x0c, 0x54, 0x65, 0x73,
	0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x65, 0x73,
	0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x65, 0x73, 0x74, 0x32, 0xb9, 0x01,
	0x0a, 0x0b, 0x53, 0x74, 0x75, 0x62, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x51, 0x0a,
	0x08, 0x54, 0x65, 0x73, 0x74, 0x43, 0x61, 0x6c, 0x6c, 0x12, 0x20, 0x2e, 0x70, 0x6b, 0x67, 0x2e,
	0x6e, 0x65, 0x74, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e,
	0x54, 0x65, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x21, 0x2e, 0x70, 0x6b,
	0x67, 0x2e, 0x6e, 0x65, 0x74, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65,
	0x72, 0x2e, 0x54, 0x65, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00,
	0x12, 0x57, 0x0a, 0x0a, 0x54, 0x65, 0x73, 0x74, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x12, 0x20,
	0x2e, 0x70, 0x6b, 0x67, 0x2e, 0x6e, 0x65, 0x74, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x73, 0x65,
	0x72, 0x76, 0x65, 0x72, 0x2e, 0x54, 0x65, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x21, 0x2e, 0x70, 0x6b, 0x67, 0x2e, 0x6e, 0x65, 0x74, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e,
	0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x54, 0x65, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x28, 0x01, 0x30, 0x01, 0x42, 0x30, 0x5a, 0x2e, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x6c, 0x67, 0x64, 0x2d, 0x64, 0x65, 0x76,
	0x2f, 0x68, 0x75, 0x62, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x6e, 0x65, 0x74, 0x2f, 0x67, 0x72, 0x70,
	0x63, 0x3b, 0x67, 0x72, 0x70, 0x63, 0x5f, 0x74, 0x65, 0x73, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_rawDescOnce sync.Once
	file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_rawDescData = file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_rawDesc
)

func file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_rawDescGZIP() []byte {
	file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_rawDescOnce.Do(func() {
		file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_rawDescData)
	})
	return file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_rawDescData
}

var file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_goTypes = []interface{}{
	(*TestRequest)(nil),  // 0: pkg.net.grpc.server.TestRequest
	(*TestResponse)(nil), // 1: pkg.net.grpc.server.TestResponse
}
var file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_depIdxs = []int32{
	0, // 0: pkg.net.grpc.server.StubService.TestCall:input_type -> pkg.net.grpc.server.TestRequest
	0, // 1: pkg.net.grpc.server.StubService.TestStream:input_type -> pkg.net.grpc.server.TestRequest
	1, // 2: pkg.net.grpc.server.StubService.TestCall:output_type -> pkg.net.grpc.server.TestResponse
	1, // 3: pkg.net.grpc.server.StubService.TestStream:output_type -> pkg.net.grpc.server.TestResponse
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_init() }
func file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_init() {
	if File_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TestRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TestResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_goTypes,
		DependencyIndexes: file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_depIdxs,
		MessageInfos:      file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_msgTypes,
	}.Build()
	File_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto = out.File
	file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_rawDesc = nil
	file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_goTypes = nil
	file_github_com_plgd_dev_hub_pkg_net_grpc_stub_proto_depIdxs = nil
}
