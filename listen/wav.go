package listen

import (
	"io"
	"os"

	concatwav "github.com/moutend/go-wav"
)

var (
	ThresholdSilence = 50
)

func IsWavEmpty(data []byte) (bool, error) {
	// Проверяем, что у нас есть достаточно данных для анализа заголовка WAV
	if len(data) < 44 {
		return false, ErrNotEnoughDataToParseWav
	}

	// Проверяем сигнатуру "RIFF" и "WAVE"
	if string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return false, ErrInvalidWav
	}

	// Получаем формат аудио из заголовка WAV
	audioFormat := int(data[20]) | int(data[21])<<8

	// Проверяем, что формат аудио - PCM
	if audioFormat != 1 {
		return false, ErrFileIsNotPCM
	}

	// Получаем число каналов из заголовка WAV
	numChannels := int(data[22]) | int(data[23])<<8

	// Получаем битность из заголовка WAV
	bitsPerSample := int(data[34]) | int(data[35])<<8

	// Вычисляем размер блока данных сэмплов
	blockSize := (bitsPerSample / 8) * numChannels

	// Анализируем данные сэмплов
	for i := 44; i < len(data); i += blockSize {
		for j := 0; j < blockSize; j++ {
			if data[i+j] != 0 {
				return false, nil
			}
		}
	}

	return true, nil
}

func WavSilence(data []byte) (int, error) {
	// Проверяем, что у нас есть достаточно данных для анализа заголовка WAV
	if len(data) < 44 {
		return 0, ErrNotEnoughDataToParseWav
	}

	// Анализируем данные сэмплов
	s := 0
	c := 0
	for i := 44; i < len(data); i += 2 {
		// Преобразуем два байта в 16-битное число
		sample := int16(data[i]) | int16(data[i+1])<<8
		s += abs(int(sample))
		c++
	}
	// fmt.Println(s / c)
	return s / c, nil
}

func IsWavSilent(data []byte) (bool, error) {
	t, err := WavSilence(data)
	return t < ThresholdSilence, err
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

func IsVoiceless(data []byte, minSample, maxSample int16) (bool, int16, error) {
	// Проверяем, что у нас есть достаточно данных для анализа заголовка WAV
	if len(data) < 44 {
		return false, 0, ErrNotEnoughDataToParseWav
	}

	// Анализируем данные сэмплов
	for i := 44; i < len(data); i += 2 {
		// Преобразуем два байта в 16-битное число
		sample := int16(data[i]) | int16(data[i+1])<<8

		// Проверяем, является ли амплитуда нулевой
		// fmt.Println("sample", sample)
		if sample > minSample && sample < maxSample {
			return true, sample, nil
		}
	}

	return false, 0, nil
}

func ConcatWav(i1, i2 []byte) []byte {

	// Create wav.File.
	a := &concatwav.File{}
	b := &concatwav.File{}

	// Unmarshal input1.wav and input2.wav.
	concatwav.Unmarshal(i1, a)
	concatwav.Unmarshal(i2, b)

	// Add input2.wav to input1.wav.
	c, _ := concatwav.New(a.SamplesPerSec(), a.BitsPerSample(), a.Channels())
	io.Copy(c, a)
	io.Copy(c, b)

	// Marshal input1.wav and save result.
	file, _ := concatwav.Marshal(c)
	return file
}

func SaveWavByte(b []byte, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	_, err = f.Write(b)
	if err != nil {
		return err
	}

	return nil
}
