package audio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"
)

const (
	riffHeaderSize = 12
	fmtChunkID     = "fmt "
	dataChunkID    = "data"
)

// WAV represents PCM WAV audio data.
type WAV struct {
	SampleRate    int
	Channels      int
	BitsPerSample int
	Data          []byte
}

// ParseWAV parses PCM WAV bytes into a WAV struct.
// Parameters: input is the raw WAV byte slice to parse.
// Returns: the parsed WAV struct or an error if parsing fails.
func ParseWAV(input []byte) (WAV, error) {
	reader := bytes.NewReader(input)
	header := make([]byte, riffHeaderSize)
	if _, err := io.ReadFull(reader, header); err != nil {
		return WAV{}, fmt.Errorf("read wav header: %w", err)
	}
	if string(header[0:4]) != "RIFF" || string(header[8:12]) != "WAVE" {
		return WAV{}, fmt.Errorf("invalid wav header")
	}
	var format WAV
	var foundFmt bool
	var foundData bool
	for reader.Len() > 0 {
		var truncated bool
		chunkID := make([]byte, 4)
		if _, err := io.ReadFull(reader, chunkID); err != nil {
			return WAV{}, fmt.Errorf("read chunk id: %w", err)
		}
		var chunkSize uint32
		if err := binary.Read(reader, binary.LittleEndian, &chunkSize); err != nil {
			return WAV{}, fmt.Errorf("read chunk size: %w", err)
		}
		switch string(chunkID) {
		case fmtChunkID:
			if err := parseFmtChunk(reader, chunkSize, &format); err != nil {
				return WAV{}, err
			}
			foundFmt = true
		case dataChunkID:
			data, wasTruncated, err := readChunkData(reader, chunkSize)
			if err != nil {
				return WAV{}, fmt.Errorf("read data chunk: %w", err)
			}
			format.Data = data
			foundData = true
			truncated = wasTruncated
		default:
			if _, err := reader.Seek(int64(chunkSize), io.SeekCurrent); err != nil {
				return WAV{}, fmt.Errorf("skip chunk: %w", err)
			}
		}
		if truncated {
			break
		}
		if chunkSize%2 == 1 {
			if _, err := reader.Seek(1, io.SeekCurrent); err != nil {
				return WAV{}, fmt.Errorf("skip padding: %w", err)
			}
		}
	}
	if !foundFmt || !foundData {
		return WAV{}, fmt.Errorf("missing fmt or data chunk")
	}
	return format, nil
}

// Concat concatenates WAV samples with optional silence padding between each clip.
// Parameters: wavs are the WAV clips to join, padding is the silence duration between clips.
// Returns: a combined WAV or an error if formats mismatch.
func Concat(wavs []WAV, padding time.Duration) (WAV, error) {
	if len(wavs) == 0 {
		return WAV{}, fmt.Errorf("no wavs to concat")
	}
	base := wavs[0]
	bytesPerSample := base.Channels * base.BitsPerSample / 8
	paddingSamples := int(float64(base.SampleRate) * padding.Seconds())
	paddingBytes := paddingSamples * bytesPerSample
	if bytesPerSample <= 0 {
		return WAV{}, fmt.Errorf("invalid wav format")
	}
	var combined []byte
	for i, wav := range wavs {
		if wav.SampleRate != base.SampleRate || wav.Channels != base.Channels || wav.BitsPerSample != base.BitsPerSample {
			return WAV{}, fmt.Errorf("wav format mismatch")
		}
		combined = append(combined, wav.Data...)
		if i < len(wavs)-1 && paddingBytes > 0 {
			combined = append(combined, make([]byte, paddingBytes)...)
		}
	}
	return WAV{
		SampleRate:    base.SampleRate,
		Channels:      base.Channels,
		BitsPerSample: base.BitsPerSample,
		Data:          combined,
	}, nil
}

// SaveWAV writes a WAV file to the specified path.
// Parameters: path is the output file location, wav is the audio payload to write.
// Returns: an error if the file cannot be written.
func SaveWAV(path string, wav WAV) error {
	data := wav.Bytes()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write wav: %w", err)
	}
	return nil
}

// Bytes encodes the WAV data with a RIFF header.
// Parameters: none.
// Returns: the encoded WAV byte slice.
func (w WAV) Bytes() []byte {
	bytesPerSample := w.Channels * w.BitsPerSample / 8
	blockAlign := uint16(bytesPerSample)
	byteRate := uint32(w.SampleRate * bytesPerSample)
	dataSize := uint32(len(w.Data))
	chunkSize := 4 + (8 + 16) + (8 + dataSize)
	buf := &bytes.Buffer{}
	buf.WriteString("RIFF")
	_ = binary.Write(buf, binary.LittleEndian, uint32(chunkSize))
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	_ = binary.Write(buf, binary.LittleEndian, uint32(16))
	_ = binary.Write(buf, binary.LittleEndian, uint16(1))
	_ = binary.Write(buf, binary.LittleEndian, uint16(w.Channels))
	_ = binary.Write(buf, binary.LittleEndian, uint32(w.SampleRate))
	_ = binary.Write(buf, binary.LittleEndian, byteRate)
	_ = binary.Write(buf, binary.LittleEndian, blockAlign)
	_ = binary.Write(buf, binary.LittleEndian, uint16(w.BitsPerSample))
	buf.WriteString("data")
	_ = binary.Write(buf, binary.LittleEndian, dataSize)
	buf.Write(w.Data)
	return buf.Bytes()
}

// readChunkData reads up to size bytes from reader, tolerating truncated data chunks.
// Parameters: reader is the WAV byte reader, size is the declared chunk size.
// Returns: the data bytes, whether the chunk was truncated, and an error for read failures.
func readChunkData(reader *bytes.Reader, size uint32) ([]byte, bool, error) {
	if size == 0 {
		return []byte{}, false, nil
	}
	remaining := reader.Len()
	if remaining == 0 {
		return []byte{}, true, nil
	}
	if int(size) > remaining {
		data := make([]byte, remaining)
		if _, err := io.ReadFull(reader, data); err != nil {
			return nil, false, err
		}
		return data, true, nil
	}
	data := make([]byte, int(size))
	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, false, err
	}
	return data, false, nil
}

// parseFmtChunk reads the fmt chunk to populate WAV format metadata.
// Parameters: reader provides chunk bytes, size is the fmt chunk size, wav receives parsed metadata.
// Returns: an error if the format chunk is invalid or unsupported.
func parseFmtChunk(reader io.Reader, size uint32, wav *WAV) error {
	if size < 16 {
		return fmt.Errorf("fmt chunk too small")
	}
	var audioFormat uint16
	var numChannels uint16
	var sampleRate uint32
	var byteRate uint32
	var blockAlign uint16
	var bitsPerSample uint16
	if err := binary.Read(reader, binary.LittleEndian, &audioFormat); err != nil {
		return fmt.Errorf("read fmt audioFormat: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &numChannels); err != nil {
		return fmt.Errorf("read fmt channels: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &sampleRate); err != nil {
		return fmt.Errorf("read fmt sampleRate: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &byteRate); err != nil {
		return fmt.Errorf("read fmt byteRate: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &blockAlign); err != nil {
		return fmt.Errorf("read fmt blockAlign: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &bitsPerSample); err != nil {
		return fmt.Errorf("read fmt bitsPerSample: %w", err)
	}
	if audioFormat != 1 {
		return fmt.Errorf("unsupported wav format: %d", audioFormat)
	}
	remaining := int64(size) - 16
	if remaining > 0 {
		if _, err := io.CopyN(io.Discard, reader, remaining); err != nil {
			return fmt.Errorf("skip fmt extras: %w", err)
		}
	}
	_ = byteRate
	_ = blockAlign
	wav.SampleRate = int(sampleRate)
	wav.Channels = int(numChannels)
	wav.BitsPerSample = int(bitsPerSample)
	return nil
}
