// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.21.9
// source: proto/option.proto

package option

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type PubSubOption struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Topic        string `protobuf:"bytes,1,opt,name=topic,proto3" json:"topic,omitempty"`
	Subscription string `protobuf:"bytes,2,opt,name=subscription,proto3" json:"subscription,omitempty"`
}

func (x *PubSubOption) Reset() {
	*x = PubSubOption{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_option_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PubSubOption) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PubSubOption) ProtoMessage() {}

func (x *PubSubOption) ProtoReflect() protoreflect.Message {
	mi := &file_proto_option_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PubSubOption.ProtoReflect.Descriptor instead.
func (*PubSubOption) Descriptor() ([]byte, []int) {
	return file_proto_option_proto_rawDescGZIP(), []int{0}
}

func (x *PubSubOption) GetTopic() string {
	if x != nil {
		return x.Topic
	}
	return ""
}

func (x *PubSubOption) GetSubscription() string {
	if x != nil {
		return x.Subscription
	}
	return ""
}

var file_proto_option_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptorpb.MethodOptions)(nil),
		ExtensionType: (*PubSubOption)(nil),
		Field:         50006,
		Name:          "events.pubSubOption",
		Tag:           "bytes,50006,opt,name=pubSubOption",
		Filename:      "proto/option.proto",
	},
}

// Extension fields to descriptorpb.MethodOptions.
var (
	// optional events.PubSubOption pubSubOption = 50006;
	E_PubSubOption = &file_proto_option_proto_extTypes[0]
)

var File_proto_option_proto protoreflect.FileDescriptor

var file_proto_option_proto_rawDesc = []byte{
	0x0a, 0x12, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x1a, 0x20, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65,
	0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x48,
	0x0a, 0x0c, 0x50, 0x75, 0x62, 0x53, 0x75, 0x62, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x14,
	0x0a, 0x05, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74,
	0x6f, 0x70, 0x69, 0x63, 0x12, 0x22, 0x0a, 0x0c, 0x73, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x73, 0x75, 0x62, 0x73,
	0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x3a, 0x5d, 0x0a, 0x0c, 0x70, 0x75, 0x62, 0x53,
	0x75, 0x62, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1e, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d, 0x65, 0x74, 0x68, 0x6f,
	0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xd6, 0x86, 0x03, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x14, 0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x50, 0x75, 0x62, 0x53, 0x75, 0x62,
	0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0c, 0x70, 0x75, 0x62, 0x53, 0x75, 0x62, 0x4f, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x88, 0x01, 0x01, 0x42, 0x4a, 0x5a, 0x48, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x54, 0x61, 0x6b, 0x61, 0x57, 0x61, 0x6b, 0x69, 0x79, 0x61,
	0x6d, 0x61, 0x2f, 0x66, 0x6f, 0x72, 0x63, 0x75, 0x73, 0x69, 0x6e, 0x67, 0x2d, 0x62, 0x61, 0x63,
	0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x63, 0x6d, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d,
	0x67, 0x65, 0x6e, 0x2d, 0x67, 0x6f, 0x2d, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x2f, 0x6f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_option_proto_rawDescOnce sync.Once
	file_proto_option_proto_rawDescData = file_proto_option_proto_rawDesc
)

func file_proto_option_proto_rawDescGZIP() []byte {
	file_proto_option_proto_rawDescOnce.Do(func() {
		file_proto_option_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_option_proto_rawDescData)
	})
	return file_proto_option_proto_rawDescData
}

var file_proto_option_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_proto_option_proto_goTypes = []interface{}{
	(*PubSubOption)(nil),               // 0: events.PubSubOption
	(*descriptorpb.MethodOptions)(nil), // 1: google.protobuf.MethodOptions
}
var file_proto_option_proto_depIdxs = []int32{
	1, // 0: events.pubSubOption:extendee -> google.protobuf.MethodOptions
	0, // 1: events.pubSubOption:type_name -> events.PubSubOption
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	1, // [1:2] is the sub-list for extension type_name
	0, // [0:1] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_proto_option_proto_init() }
func file_proto_option_proto_init() {
	if File_proto_option_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_option_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PubSubOption); i {
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
			RawDescriptor: file_proto_option_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 1,
			NumServices:   0,
		},
		GoTypes:           file_proto_option_proto_goTypes,
		DependencyIndexes: file_proto_option_proto_depIdxs,
		MessageInfos:      file_proto_option_proto_msgTypes,
		ExtensionInfos:    file_proto_option_proto_extTypes,
	}.Build()
	File_proto_option_proto = out.File
	file_proto_option_proto_rawDesc = nil
	file_proto_option_proto_goTypes = nil
	file_proto_option_proto_depIdxs = nil
}
