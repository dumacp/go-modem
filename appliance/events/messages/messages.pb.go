// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: messages.proto

package messages

import (
	fmt "fmt"
	_ "github.com/AsynkronIT/protoactor-go/actor"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
	reflect "reflect"
	strconv "strconv"
	strings "strings"
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

//Ignition
type StateType int32

const (
	UP   StateType = 0
	DOWN StateType = 1
	NA   StateType = 2
)

var StateType_name = map[int32]string{
	0: "UP",
	1: "DOWN",
	2: "NA",
}

var StateType_value = map[string]int32{
	"UP":   0,
	"DOWN": 1,
	"NA":   2,
}

func (StateType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_4dc296cbfe5ffcd5, []int{0}
}

type IgnitionEvent struct {
	Event     StateType `protobuf:"varint,1,opt,name=Event,proto3,enum=messages.StateType" json:"Event,omitempty"`
	TimeStamp int64     `protobuf:"varint,2,opt,name=TimeStamp,proto3" json:"TimeStamp,omitempty"`
}

func (m *IgnitionEvent) Reset()      { *m = IgnitionEvent{} }
func (*IgnitionEvent) ProtoMessage() {}
func (*IgnitionEvent) Descriptor() ([]byte, []int) {
	return fileDescriptor_4dc296cbfe5ffcd5, []int{0}
}
func (m *IgnitionEvent) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *IgnitionEvent) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_IgnitionEvent.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *IgnitionEvent) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IgnitionEvent.Merge(m, src)
}
func (m *IgnitionEvent) XXX_Size() int {
	return m.Size()
}
func (m *IgnitionEvent) XXX_DiscardUnknown() {
	xxx_messageInfo_IgnitionEvent.DiscardUnknown(m)
}

var xxx_messageInfo_IgnitionEvent proto.InternalMessageInfo

func (m *IgnitionEvent) GetEvent() StateType {
	if m != nil {
		return m.Event
	}
	return UP
}

func (m *IgnitionEvent) GetTimeStamp() int64 {
	if m != nil {
		return m.TimeStamp
	}
	return 0
}

func init() {
	proto.RegisterEnum("messages.StateType", StateType_name, StateType_value)
	proto.RegisterType((*IgnitionEvent)(nil), "messages.IgnitionEvent")
}

func init() { proto.RegisterFile("messages.proto", fileDescriptor_4dc296cbfe5ffcd5) }

var fileDescriptor_4dc296cbfe5ffcd5 = []byte{
	// 245 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0xcb, 0x4d, 0x2d, 0x2e,
	0x4e, 0x4c, 0x4f, 0x2d, 0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x80, 0xf1, 0xa5, 0xcc,
	0xd2, 0x33, 0x4b, 0x32, 0x4a, 0x93, 0xf4, 0x92, 0xf3, 0x73, 0xf5, 0x1d, 0x8b, 0x2b, 0xf3, 0xb2,
	0x8b, 0xf2, 0xf3, 0x3c, 0x43, 0xf4, 0xc1, 0xca, 0x12, 0x93, 0x4b, 0xf2, 0x8b, 0x74, 0xd3, 0xf3,
	0xf5, 0xc1, 0x0c, 0x88, 0x18, 0xd4, 0x04, 0xa5, 0x08, 0x2e, 0x5e, 0xcf, 0xf4, 0xbc, 0xcc, 0x92,
	0xcc, 0xfc, 0x3c, 0xd7, 0xb2, 0xd4, 0xbc, 0x12, 0x21, 0x4d, 0x2e, 0x56, 0x30, 0x43, 0x82, 0x51,
	0x81, 0x51, 0x83, 0xcf, 0x48, 0x58, 0x0f, 0x6e, 0x65, 0x70, 0x49, 0x62, 0x49, 0x6a, 0x48, 0x65,
	0x41, 0x6a, 0x10, 0x44, 0x85, 0x90, 0x0c, 0x17, 0x67, 0x48, 0x66, 0x6e, 0x6a, 0x70, 0x49, 0x62,
	0x6e, 0x81, 0x04, 0x93, 0x02, 0xa3, 0x06, 0x73, 0x10, 0x42, 0x40, 0x4b, 0x95, 0x8b, 0x13, 0xae,
	0x43, 0x88, 0x8d, 0x8b, 0x29, 0x34, 0x40, 0x80, 0x41, 0x88, 0x83, 0x8b, 0xc5, 0xc5, 0x3f, 0xdc,
	0x4f, 0x80, 0x11, 0x24, 0xe2, 0xe7, 0x28, 0xc0, 0xe4, 0x64, 0x72, 0xe1, 0xa1, 0x1c, 0xc3, 0x8d,
	0x87, 0x72, 0x0c, 0x1f, 0x1e, 0xca, 0x31, 0x36, 0x3c, 0x92, 0x63, 0x5c, 0xf1, 0x48, 0x8e, 0xf1,
	0xc4, 0x23, 0x39, 0xc6, 0x0b, 0x8f, 0xe4, 0x18, 0x1f, 0x3c, 0x92, 0x63, 0x7c, 0xf1, 0x48, 0x8e,
	0xe1, 0xc3, 0x23, 0x39, 0xc6, 0x09, 0x8f, 0xe5, 0x18, 0x2e, 0x3c, 0x96, 0x63, 0xb8, 0xf1, 0x58,
	0x8e, 0x21, 0x89, 0x0d, 0xec, 0x7a, 0x63, 0x40, 0x00, 0x00, 0x00, 0xff, 0xff, 0x3a, 0xd7, 0x0e,
	0x66, 0x11, 0x01, 0x00, 0x00,
}

func (x StateType) String() string {
	s, ok := StateType_name[int32(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}
func (this *IgnitionEvent) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*IgnitionEvent)
	if !ok {
		that2, ok := that.(IgnitionEvent)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.Event != that1.Event {
		return false
	}
	if this.TimeStamp != that1.TimeStamp {
		return false
	}
	return true
}
func (this *IgnitionEvent) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 6)
	s = append(s, "&messages.IgnitionEvent{")
	s = append(s, "Event: "+fmt.Sprintf("%#v", this.Event)+",\n")
	s = append(s, "TimeStamp: "+fmt.Sprintf("%#v", this.TimeStamp)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func valueToGoStringMessages(v interface{}, typ string) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("func(v %v) *%v { return &v } ( %#v )", typ, typ, pv)
}
func (m *IgnitionEvent) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *IgnitionEvent) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *IgnitionEvent) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.TimeStamp != 0 {
		i = encodeVarintMessages(dAtA, i, uint64(m.TimeStamp))
		i--
		dAtA[i] = 0x10
	}
	if m.Event != 0 {
		i = encodeVarintMessages(dAtA, i, uint64(m.Event))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintMessages(dAtA []byte, offset int, v uint64) int {
	offset -= sovMessages(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *IgnitionEvent) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Event != 0 {
		n += 1 + sovMessages(uint64(m.Event))
	}
	if m.TimeStamp != 0 {
		n += 1 + sovMessages(uint64(m.TimeStamp))
	}
	return n
}

func sovMessages(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozMessages(x uint64) (n int) {
	return sovMessages(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *IgnitionEvent) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&IgnitionEvent{`,
		`Event:` + fmt.Sprintf("%v", this.Event) + `,`,
		`TimeStamp:` + fmt.Sprintf("%v", this.TimeStamp) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringMessages(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *IgnitionEvent) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMessages
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
			return fmt.Errorf("proto: IgnitionEvent: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: IgnitionEvent: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Event", wireType)
			}
			m.Event = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessages
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Event |= StateType(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TimeStamp", wireType)
			}
			m.TimeStamp = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessages
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TimeStamp |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipMessages(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMessages
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthMessages
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
func skipMessages(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowMessages
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
					return 0, ErrIntOverflowMessages
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
					return 0, ErrIntOverflowMessages
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
				return 0, ErrInvalidLengthMessages
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupMessages
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthMessages
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthMessages        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMessages          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupMessages = fmt.Errorf("proto: unexpected end of group")
)
