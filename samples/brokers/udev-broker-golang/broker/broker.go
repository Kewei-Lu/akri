package broker

import (
	"context"

	pb "udev-broker-golang/grpc"

	"github.com/vladimirvivien/go4vl/device"
	"github.com/vladimirvivien/go4vl/v4l2"
)

var (
	preferredPixFmts   []v4l2.FourCCType = []v4l2.FourCCType{v4l2.PixelFmtMPEG, v4l2.PixelFmtMJPEG, v4l2.PixelFmtJPEG, v4l2.PixelFmtYUYV}
	PreferredPixWidth  uint32            = 800
	PreferredPixHeight uint32            = 600
)

type Broker struct {
	devPath string
	device  *device.Device
	// PreferredPixFmt    string
	// PreferredPixWidth  string
	// PreferredPixHeight string
	fmtDesc *v4l2.FormatDescription
	pixSize *v4l2.FrameSizeEnum
	Ready   bool
	pb.UnimplementedCameraServer
}

func InitBroker() *Broker {
	broker := &Broker{
		devPath: DEVPATH,
	}
	broker.OpenDevice()
	broker.NegotiateFmt()
	broker.OpenDevice()
	broker.startCapture()
	return broker
}

func (b *Broker) OpenDevice() error {
	dev, err := device.Open(b.devPath)
	if err != nil {
		Logger.Errorf("Open device %s failed, error: %s", b.devPath, err.Error())
		b.Ready = false
		return nil
	}
	b.device = dev

	return nil
}

func (b *Broker) NegotiateFmt() error {
	if b.device == nil {
		Logger.Errorf("Cannot find device %s", b.devPath)
		return nil
	}
	fmtDescs, err := b.device.GetFormatDescriptions()
	if err != nil {
		Logger.Errorf("Get desc error %v", err)
		return nil
	}
	if len(fmtDescs) == 0 {
		Logger.Error("Empty format descriptions")
		return nil
	}
	// negotiate encode format
	for _, preferredPixFmt := range preferredPixFmts {
		for _, desc := range fmtDescs {
			if desc.PixelFormat == preferredPixFmt {
				b.fmtDesc = &desc
				break
			}
		}
	}
	if b.fmtDesc == nil {
		Logger.Warnf("No preferred encoding format is supported, will use the first supported one as default: %d", fmtDescs[0].PixelFormat)
		b.fmtDesc = &fmtDescs[0]
	}

	// negotiate pix format
	frameSizes, err := v4l2.GetFormatFrameSizes(b.device.Fd(), b.fmtDesc.PixelFormat)
	if err != nil {
		Logger.Errorf("Get frameSizes error %v", err)
		return nil
	}
	if len(frameSizes) == 0 {
		Logger.Error("Empty frameSizes")
		return nil
	}
	for _, size := range frameSizes {
		Logger.Debugf("SIZE: %#v", size)
		if size.Size.MinWidth <= uint32(PreferredPixWidth) && size.Size.MinHeight <= uint32(PreferredPixHeight) {
			b.pixSize = &size
			break
		}
	}
	if b.pixSize == nil {
		Logger.Warnf("No preferred pix resolution is supported, will use the first supported one as default: WIDTH: %d, HEIGHT: %d", frameSizes[0].Size.MinWidth, frameSizes[0].Size.MinHeight)
		b.pixSize = &frameSizes[0]
	}
	Logger.Debugf("Broker Config: %v", b)
	return nil
}

func (b *Broker) startCapture() error {
	if b.device == nil {
		Logger.Errorf("nil device")
		return nil
	}
	// cancel, ctx := context.WithCancel(context.TODO())
	if err := b.device.Start(context.TODO()); err != nil {
		Logger.Fatal("Failed to start device")
		return nil
	}
	b.Ready = true
	Logger.Debugf("Start video capture success, DEVICE PATH: %s, BUFFER SIZE: %d", b.devPath, b.device.BufferCount())
	return nil
}

// GRPC function
func (b *Broker) GetFrame(ctx context.Context, notify *pb.NotifyRequest) (*pb.NotifyResponse, error) {
	if !b.Ready {
		Logger.Errorf("device not ready yet")
		return nil, nil
	}
	frame := <-b.device.GetOutput()
	// cancel, ctx := context.WithCancel(context.TODO())

	return &pb.NotifyResponse{Frame: frame, Camera: b.device.Name()}, nil
}

// func (b *Broker) GetStream(request *pb.NotifyRequest, stream pb.CameraServer) error {
// 	if !b.Ready {
// 		Logger.Errorf("device not ready yet")
// 		return nil
// 	}
// 	for frame := range b.device.GetOutput() {
// 		if err := stream.Send(&pb.NotifyResponse{Frame: frame, Camera: b.device.Name()}); err != nil {
// 			return err
// 		}
// 	}
// 	// cancel, ctx := context.WithCancel(context.TODO())

// 	return nil
// }
