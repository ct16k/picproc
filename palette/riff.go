package palette

import (
	"encoding/binary"
	"fmt"
	"image/color"
	"io"

	"golang.org/x/image/riff"
)

/*
typedef struct tagLOGPALETTE {
  WORD         palVersion;
  WORD         palNumEntries;
  PALETTEENTRY palPalEntry[1];
} LOGPALETTE;

typedef struct tagPALETTEENTRY {
  BYTE peRed;
  BYTE peGreen;
  BYTE peBlue;
  BYTE peFlags;
} PALETTEENTRY;
*/

type PaletteConverter interface {
	From(color.Palette) int64
	To(color.Model) (int64, color.Palette)
}

type PaletteRIFFReaderWriter interface {
	ReadRIFF(io.Reader) (int64, error)
	WriteRIFF(io.Writer) (int64, error)
}

var (
	riffType = riff.FourCC{'R', 'I', 'F', 'F'}
	palType  = riff.FourCC{'P', 'A', 'L', ' '}
	dataType = riff.FourCC{'d', 'a', 't', 'a'}
)

func ReadFrom(r io.Reader) ([]color.Palette, error) {
	formType, rd, err := riff.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("could not open RIFF stream: %w", err)
	} else if formType != palType {
		return nil, fmt.Errorf("unsupported RIFF content type: %s", string(formType[:]))
	}

	return readPalettes(rd, string(formType[:]))
}

func readPalettes(r *riff.Reader, ident string) ([]color.Palette, error) {
	var res []color.Palette

	for {
		id, size, data, err := r.Next()
		if err != nil {
			if err == io.EOF {
				break
			}

			return res, fmt.Errorf("could not read chunk %q#%d: %w", ident, len(res), err)
		}

		if id == riff.LIST {
			listType, list, lerr := riff.NewListReader(size, data)
			if lerr != nil {
				return res, fmt.Errorf("could not read list from chunk %q#%d: %w", ident, len(res), lerr)
			} else if listType != palType {
				return nil, fmt.Errorf("chunk %q#%d unsupported type: %s", ident, len(res), string(listType[:]))
			}

			if listRes, lerr := readPalettes(list, fmt.Sprintf("%s%d.%s", ident, len(res), listType[:])); lerr != nil {
				return append(res, listRes...), lerr
			}
		} else if id != dataType {
			return res, fmt.Errorf("unsupported chunk type in %q#%d: %s", ident, len(res), id)
		}

		pal, err := readPalette(data, fmt.Sprintf("%s%d", ident, len(res)))
		if err != nil {
			return res, err
		}

		res = append(res, pal)
	}

	return res, nil
}

func readPalette(r io.Reader, ident string) (color.Palette, error) {
	buf := make([]byte, 2)

	n, err := r.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("could not read version from chunk %s: %w", ident, err)
	} else if n != 2 {
		return nil, fmt.Errorf("not enough bytes in %s to read version number: %d", ident, n)
	}

	ver := binary.BigEndian.Uint16(buf)
	if ver != 3 {
		return nil, fmt.Errorf("unsupported palette version in chunk %s: %d", ident, ver)
	}

	n, err = r.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("could not read number of entries from chunk %s: %w", ident, err)
	} else if n != 2 {
		return nil, fmt.Errorf("not enough bytes in %s to read number of entries: %d", ident, n)
	}

	count := binary.LittleEndian.Uint16(buf)
	res := make([]color.Color, count)
	buf4 := make([]byte, 4)
	for i := range count {
		n, err = r.Read(buf4)
		if err != nil {
			return res, fmt.Errorf("could not read color %d/%d from chunk %s: %w", i, count, ident, err)
		} else if n != 4 {
			return res, fmt.Errorf("not enough bytes to read color %d/%d from chunk %s: %d", i, count, ident, n)
		}

		res[i] = color.RGBA{
			R: buf4[0],
			G: buf4[1],
			B: buf4[2],
		}
	}

	return res, nil
}

func WriteTo(w io.Writer, pals []color.Palette) (int64, error) {
	n := 4
	for _, pal := range pals {
		n += 4 + 4 + 4 + len(pal)*4 // chunk id + chunk size + palVersion + palNumEntries + 4 bytes/color
	}

	if err := writeByes(w, riffType[:]); err != nil {
		return 0, fmt.Errorf("could not write RIFF magic: %w", err)
	}

	if err := writeByes(w, binary.LittleEndian.AppendUint32(nil, uint32(n))); err != nil {
		return 0, fmt.Errorf("could not write document size: %w", err)
	}

	if err := writeByes(w, palType[:]); err != nil {
		return 0, fmt.Errorf("could not write content type: %w", err)
	}

	var count int64
	for i, pal := range pals {
		if n, err := writePalette(w, pal); err != nil {
			count += n
			return count, fmt.Errorf("could not write chunk %d: %w", i, err)
		} else {
			count += n
		}
	}

	return count, nil
}

func writePalette(w io.Writer, pal color.Palette) (int64, error) {
	if err := writeByes(w, dataType[:]); err != nil {
		return 0, fmt.Errorf("could not write type: %w", err)
	}

	n := 4 + len(pal)*4
	if err := writeByes(w, binary.LittleEndian.AppendUint32(nil, uint32(n))); err != nil {
		return 0, fmt.Errorf("could not write chunk size: %w", err)
	}

	if err := writeByes(w, []byte{0, 0x03}); err != nil {
		return 0, fmt.Errorf("could not write palette version: %w", err)
	}

	if err := writeByes(w, binary.LittleEndian.AppendUint16(nil, uint16(len(pal)))); err != nil {
		return 0, fmt.Errorf("could not write number of colors: %w", err)
	}

	for i, col := range pal {
		c := color.RGBA64Model.Convert(col).(color.RGBA)
		if err := writeByes(w, []byte{c.R, c.G, c.B, 0x00}); err != nil {
			return int64(i), fmt.Errorf("could not write color %d/%d: %w", i, len(pal), err)
		}
	}

	return int64(len(pal)), nil
}

func writeByes(w io.Writer, b []byte) error {
	n, err := w.Write(b)
	if err != nil {
		return err
	} else if n != len(b) {
		return fmt.Errorf("wrote only %d/%d bytes", n, len(b))
	}

	return nil
}
