package debate

import (
	"fmt"
	"os"

	"github.com/bentos-lab/parley/shared/audio"
)

func fileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("stat %s: %w", path, err)
	}
	return !info.IsDir(), nil
}

func loadWAV(path string) (audio.WAV, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return audio.WAV{}, false, nil
		}
		return audio.WAV{}, false, fmt.Errorf("read wav %s: %w", path, err)
	}
	wav, err := audio.ParseWAV(data)
	if err != nil {
		return audio.WAV{}, false, fmt.Errorf("parse wav %s: %w", path, err)
	}
	return wav, true, nil
}
