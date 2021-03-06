// This file was auto-generated by the vanadium vdl tool.
// Package: downloader

package downloader

import (
	"v.io/v23"
	"v.io/v23/context"
	"v.io/v23/rpc"
	"v.io/v23/vdl"
)

var _ = __VDLInit() // Must be first; see __VDLInit comments for details.

//////////////////////////////////////////////////
// Type definitions

type Chunk struct {
	Url    string
	Owner  string
	Offset int32
	Length int32
	Data   []byte
}

func (Chunk) __VDLReflect(struct {
	Name string `vdl:"downloader/protocol.Chunk"`
}) {
}

func (x Chunk) VDLIsZero() bool {
	if x.Url != "" {
		return false
	}
	if x.Owner != "" {
		return false
	}
	if x.Offset != 0 {
		return false
	}
	if x.Length != 0 {
		return false
	}
	if len(x.Data) != 0 {
		return false
	}
	return true
}

func (x Chunk) VDLWrite(enc vdl.Encoder) error {
	if err := enc.StartValue(__VDLType_struct_1); err != nil {
		return err
	}
	if x.Url != "" {
		if err := enc.NextFieldValueString(0, vdl.StringType, x.Url); err != nil {
			return err
		}
	}
	if x.Owner != "" {
		if err := enc.NextFieldValueString(1, vdl.StringType, x.Owner); err != nil {
			return err
		}
	}
	if x.Offset != 0 {
		if err := enc.NextFieldValueInt(2, vdl.Int32Type, int64(x.Offset)); err != nil {
			return err
		}
	}
	if x.Length != 0 {
		if err := enc.NextFieldValueInt(3, vdl.Int32Type, int64(x.Length)); err != nil {
			return err
		}
	}
	if len(x.Data) != 0 {
		if err := enc.NextFieldValueBytes(4, __VDLType_list_2, x.Data); err != nil {
			return err
		}
	}
	if err := enc.NextField(-1); err != nil {
		return err
	}
	return enc.FinishValue()
}

func (x *Chunk) VDLRead(dec vdl.Decoder) error {
	*x = Chunk{}
	if err := dec.StartValue(__VDLType_struct_1); err != nil {
		return err
	}
	decType := dec.Type()
	for {
		index, err := dec.NextField()
		switch {
		case err != nil:
			return err
		case index == -1:
			return dec.FinishValue()
		}
		if decType != __VDLType_struct_1 {
			index = __VDLType_struct_1.FieldIndexByName(decType.Field(index).Name)
			if index == -1 {
				if err := dec.SkipValue(); err != nil {
					return err
				}
				continue
			}
		}
		switch index {
		case 0:
			switch value, err := dec.ReadValueString(); {
			case err != nil:
				return err
			default:
				x.Url = value
			}
		case 1:
			switch value, err := dec.ReadValueString(); {
			case err != nil:
				return err
			default:
				x.Owner = value
			}
		case 2:
			switch value, err := dec.ReadValueInt(32); {
			case err != nil:
				return err
			default:
				x.Offset = int32(value)
			}
		case 3:
			switch value, err := dec.ReadValueInt(32); {
			case err != nil:
				return err
			default:
				x.Length = int32(value)
			}
		case 4:
			if err := dec.ReadValueBytes(-1, &x.Data); err != nil {
				return err
			}
		}
	}
}

//////////////////////////////////////////////////
// Interface definitions

// DownloaderClientMethods is the client interface
// containing Downloader methods.
type DownloaderClientMethods interface {
	GetChunk(*context.T, ...rpc.CallOpt) (newChunk Chunk, _ error)
	TransferChunk(_ *context.T, downloadedChunk Chunk, _ ...rpc.CallOpt) error
}

// DownloaderClientStub adds universal methods to DownloaderClientMethods.
type DownloaderClientStub interface {
	DownloaderClientMethods
	rpc.UniversalServiceMethods
}

// DownloaderClient returns a client stub for Downloader.
func DownloaderClient(name string) DownloaderClientStub {
	return implDownloaderClientStub{name}
}

type implDownloaderClientStub struct {
	name string
}

func (c implDownloaderClientStub) GetChunk(ctx *context.T, opts ...rpc.CallOpt) (o0 Chunk, err error) {
	err = v23.GetClient(ctx).Call(ctx, c.name, "GetChunk", nil, []interface{}{&o0}, opts...)
	return
}

func (c implDownloaderClientStub) TransferChunk(ctx *context.T, i0 Chunk, opts ...rpc.CallOpt) (err error) {
	err = v23.GetClient(ctx).Call(ctx, c.name, "TransferChunk", []interface{}{i0}, nil, opts...)
	return
}

// DownloaderServerMethods is the interface a server writer
// implements for Downloader.
type DownloaderServerMethods interface {
	GetChunk(*context.T, rpc.ServerCall) (newChunk Chunk, _ error)
	TransferChunk(_ *context.T, _ rpc.ServerCall, downloadedChunk Chunk) error
}

// DownloaderServerStubMethods is the server interface containing
// Downloader methods, as expected by rpc.Server.
// There is no difference between this interface and DownloaderServerMethods
// since there are no streaming methods.
type DownloaderServerStubMethods DownloaderServerMethods

// DownloaderServerStub adds universal methods to DownloaderServerStubMethods.
type DownloaderServerStub interface {
	DownloaderServerStubMethods
	// Describe the Downloader interfaces.
	Describe__() []rpc.InterfaceDesc
}

// DownloaderServer returns a server stub for Downloader.
// It converts an implementation of DownloaderServerMethods into
// an object that may be used by rpc.Server.
func DownloaderServer(impl DownloaderServerMethods) DownloaderServerStub {
	stub := implDownloaderServerStub{
		impl: impl,
	}
	// Initialize GlobState; always check the stub itself first, to handle the
	// case where the user has the Glob method defined in their VDL source.
	if gs := rpc.NewGlobState(stub); gs != nil {
		stub.gs = gs
	} else if gs := rpc.NewGlobState(impl); gs != nil {
		stub.gs = gs
	}
	return stub
}

type implDownloaderServerStub struct {
	impl DownloaderServerMethods
	gs   *rpc.GlobState
}

func (s implDownloaderServerStub) GetChunk(ctx *context.T, call rpc.ServerCall) (Chunk, error) {
	return s.impl.GetChunk(ctx, call)
}

func (s implDownloaderServerStub) TransferChunk(ctx *context.T, call rpc.ServerCall, i0 Chunk) error {
	return s.impl.TransferChunk(ctx, call, i0)
}

func (s implDownloaderServerStub) Globber() *rpc.GlobState {
	return s.gs
}

func (s implDownloaderServerStub) Describe__() []rpc.InterfaceDesc {
	return []rpc.InterfaceDesc{DownloaderDesc}
}

// DownloaderDesc describes the Downloader interface.
var DownloaderDesc rpc.InterfaceDesc = descDownloader

// descDownloader hides the desc to keep godoc clean.
var descDownloader = rpc.InterfaceDesc{
	Name:    "Downloader",
	PkgPath: "downloader/protocol",
	Methods: []rpc.MethodDesc{
		{
			Name: "GetChunk",
			OutArgs: []rpc.ArgDesc{
				{"newChunk", ``}, // Chunk
			},
		},
		{
			Name: "TransferChunk",
			InArgs: []rpc.ArgDesc{
				{"downloadedChunk", ``}, // Chunk
			},
		},
	},
}

// Hold type definitions in package-level variables, for better performance.
var (
	__VDLType_struct_1 *vdl.Type
	__VDLType_list_2   *vdl.Type
)

var __VDLInitCalled bool

// __VDLInit performs vdl initialization.  It is safe to call multiple times.
// If you have an init ordering issue, just insert the following line verbatim
// into your source files in this package, right after the "package foo" clause:
//
//    var _ = __VDLInit()
//
// The purpose of this function is to ensure that vdl initialization occurs in
// the right order, and very early in the init sequence.  In particular, vdl
// registration and package variable initialization needs to occur before
// functions like vdl.TypeOf will work properly.
//
// This function returns a dummy value, so that it can be used to initialize the
// first var in the file, to take advantage of Go's defined init order.
func __VDLInit() struct{} {
	if __VDLInitCalled {
		return struct{}{}
	}
	__VDLInitCalled = true

	// Register types.
	vdl.Register((*Chunk)(nil))

	// Initialize type definitions.
	__VDLType_struct_1 = vdl.TypeOf((*Chunk)(nil)).Elem()
	__VDLType_list_2 = vdl.TypeOf((*[]byte)(nil))

	return struct{}{}
}
