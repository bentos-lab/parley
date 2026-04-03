package audio

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/hraban/opus"
)

const (
	oggStreamSerialMax = 1<<31 - 1
	resampleRate       = 48000
	encodeFrameSamples = 960
)

var oggCRCTable [256]uint32

func init() {
	for i := range 256 {
		r := uint32(i << 24)
		for range 8 {
			if (r & 0x80000000) != 0 {
				r = (r << 1) ^ 0x04C11DB7
			} else {
				r <<= 1
			}
		}
		oggCRCTable[i] = r
	}
}

func oggCRC(data []byte) uint32 {
	var crc uint32
	for _, b := range data {
		crc = (crc << 8) ^ oggCRCTable[byte(crc>>24)^b]
	}
	return crc
}

// ConvertWAVToOpus converts a PCM WAV file into an OGG/Opus voice note.
// Parameters: ctx governs cancellation, wavPath points to the source PCM WAV file.
// Returns: path of a temporary OGG file, duration in seconds (rounded), cleanup func that deletes the temp file, and any error.
func ConvertWAVToOpus(ctx context.Context, wavPath string) (string, uint32, func(), error) {
	data, err := os.ReadFile(wavPath)
	if err != nil {
		return "", 0, nil, fmt.Errorf("read wav for conversion: %w", err)
	}
	wav, err := ParseWAV(data)
	if err != nil {
		return "", 0, nil, fmt.Errorf("parse wav for conversion: %w", err)
	}
	if wav.BitsPerSample != 16 {
		return "", 0, nil, fmt.Errorf("unsupported bits per sample: %d", wav.BitsPerSample)
	}
	frames := len(wav.Data) / 2
	if wav.Channels <= 0 {
		return "", 0, nil, fmt.Errorf("invalid channel count: %d", wav.Channels)
	}
	sampleFrames := frames / wav.Channels
	if sampleFrames == 0 || wav.SampleRate <= 0 {
		return "", 0, nil, fmt.Errorf("invalid wav duration")
	}
	pcm := make([]int16, frames)
	for i := range pcm {
		pcm[i] = int16(binary.LittleEndian.Uint16(wav.Data[2*i:]))
	}

	monoPCM := downmixToMono(pcm, wav.Channels)
	resampled := resamplePCM(monoPCM, wav.SampleRate, resampleRate, 1)
	tmp, err := os.CreateTemp("", "parley-voice-*.ogg")
	if err != nil {
		return "", 0, nil, fmt.Errorf("create temporary ogg file: %w", err)
	}

	durationSeconds := uint32(math.Ceil(float64(len(resampled)) / float64(resampleRate)))
	if durationSeconds == 0 {
		durationSeconds = 1
	}

	tmpPath := tmp.Name()
	cleanup := func() {
		_ = os.Remove(tmpPath)
	}
	writer := newOggWriter(tmp)
	if err := writer.writePage(0x02, 0, opusHeadPacket(1, resampleRate)); err != nil {
		cleanup()
		tmp.Close()
		return "", 0, nil, err
	}
	if err := writer.writePage(0x00, 0, opusTagsPacket("github.com/bentos-lab/parley")); err != nil {
		cleanup()
		tmp.Close()
		return "", 0, nil, err
	}
	encoder, err := opus.NewEncoder(resampleRate, 1, opus.AppVoIP)
	if err != nil {
		cleanup()
		tmp.Close()
		return "", 0, nil, fmt.Errorf("initialize opus encoder: %w", err)
	}
	packetBuf := make([]byte, 4000)
	resampledFrames := len(resampled)
	const opusPreSkip = 312
	granule := uint64(opusPreSkip)

	for offset := 0; offset < resampledFrames; offset += encodeFrameSamples {
		select {
		case <-ctx.Done():
			cleanup()
			_ = tmp.Close()
			return "", 0, nil, ctx.Err()
		default:
		}

		end := min(offset+encodeFrameSamples, resampledFrames)
		chunk := resampled[offset:end]
		chunkFrames := end - offset
		chunkBuf := chunk

		if chunkFrames < encodeFrameSamples {
			padded := make([]int16, encodeFrameSamples)
			copy(padded, chunk)
			chunkBuf = padded
		}

		n, err := encoder.Encode(chunkBuf, packetBuf)
		if err != nil {
			cleanup()
			_ = tmp.Close()
			return "", 0, nil, fmt.Errorf("encode opus frame: %w", err)
		}

		granule += uint64(chunkFrames)

		headerType := byte(0)
		if end == resampledFrames {
			headerType |= 0x04
		}

		if err := writer.writePage(headerType, granule, packetBuf[:n]); err != nil {
			cleanup()
			_ = tmp.Close()
			return "", 0, nil, err
		}
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return "", 0, nil, fmt.Errorf("close ogg file: %w", err)
	}
	return tmpPath, durationSeconds, cleanup, nil
}

func resamplePCM(pcm []int16, srcRate, dstRate, channels int) []int16 {
	if srcRate == dstRate {
		return pcm
	}
	srcFrames := len(pcm) / channels
	if srcFrames == 0 {
		return pcm
	}
	dstFrames := int(math.Round(float64(srcFrames) * float64(dstRate) / float64(srcRate)))
	if dstFrames < 1 {
		dstFrames = 1
	}
	ratio := float64(srcRate) / float64(dstRate)
	result := make([]int16, dstFrames*channels)
	for frame := 0; frame < dstFrames; frame++ {
		pos := float64(frame) * ratio
		idx := int(pos)
		frac := pos - float64(idx)
		if idx >= srcFrames {
			idx = srcFrames - 1
			frac = 0
		}
		nextIdx := idx + 1
		if nextIdx >= srcFrames {
			nextIdx = srcFrames - 1
		}
		for ch := 0; ch < channels; ch++ {
			a := float64(pcm[idx*channels+ch])
			b := float64(pcm[nextIdx*channels+ch])
			value := a*(1-frac) + b*frac
			result[frame*channels+ch] = int16(math.Round(value))
		}
	}
	return result
}

type oggWriter struct {
	w      io.Writer
	seq    uint32
	stream uint32
}

func newOggWriter(w io.Writer) *oggWriter {
	var streamID uint32
	var buf [4]byte

	if _, err := rand.Read(buf[:]); err == nil {
		streamID = (binary.LittleEndian.Uint32(buf[:]) & oggStreamSerialMax) + 1
	} else {
		streamID = 1
	}

	return &oggWriter{
		w:      w,
		seq:    0,
		stream: streamID,
	}
}

func buildLacing(payloadLen int) []byte {
	if payloadLen == 0 {
		return []byte{0}
	}

	segments := make([]byte, 0, payloadLen/255+1)
	for payloadLen >= 255 {
		segments = append(segments, 255)
		payloadLen -= 255
	}
	segments = append(segments, byte(payloadLen))
	return segments
}
func (o *oggWriter) writePage(headerType byte, granule uint64, payload []byte) error {
	if len(payload) == 0 {
		return fmt.Errorf("ogg page payload empty")
	}

	laces := buildLacing(len(payload))
	header := make([]byte, 27+len(laces))

	copy(header[:4], []byte("OggS"))
	header[4] = 0
	header[5] = headerType
	binary.LittleEndian.PutUint64(header[6:], granule)
	binary.LittleEndian.PutUint32(header[14:], o.stream)
	binary.LittleEndian.PutUint32(header[18:], o.seq)
	header[26] = byte(len(laces))

	copy(header[27:], laces)

	page := append(header, payload...)

	binary.LittleEndian.PutUint32(page[22:], 0)
	checksum := oggCRC(page)
	binary.LittleEndian.PutUint32(page[22:], checksum)

	if _, err := o.w.Write(page); err != nil {
		return err
	}

	o.seq++
	return nil
}

func opusHeadPacket(channels, sampleRate int) []byte {
	packet := make([]byte, 19)
	copy(packet, []byte("OpusHead"))
	packet[8] = 1
	packet[9] = byte(channels)

	// pre-skip
	binary.LittleEndian.PutUint16(packet[10:], 312)

	binary.LittleEndian.PutUint32(packet[12:], uint32(sampleRate))
	binary.LittleEndian.PutUint16(packet[16:], 0)
	packet[18] = 0
	return packet
}

func opusTagsPacket(vendor string) []byte {
	vendorBytes := []byte(vendor)
	packet := make([]byte, 8+4+len(vendorBytes)+4)
	copy(packet, []byte("OpusTags"))
	binary.LittleEndian.PutUint32(packet[8:], uint32(len(vendorBytes)))
	copy(packet[12:], vendorBytes)
	binary.LittleEndian.PutUint32(packet[12+len(vendorBytes):], 0)
	return packet
}

func downmixToMono(pcm []int16, channels int) []int16 {
	if channels == 1 {
		return pcm
	}

	frames := len(pcm) / channels
	out := make([]int16, frames)

	for i := 0; i < frames; i++ {
		sum := 0
		for ch := 0; ch < channels; ch++ {
			sum += int(pcm[i*channels+ch])
		}
		out[i] = int16(sum / channels)
	}

	return out
}
