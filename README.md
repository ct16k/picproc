# picproc

## Summary
picproc is a tool to process pictures. Currently only sorts between portrait and landscape photos, or resizes, applying
dithering and custom palettes. One day it may learn to do more.

## Commands
```
  orient cp    Copy images to their respective folders
  orient mv    Move images to their respective folders
  mangle       Mangle an image
```

### orient cp|mv
```
Flags:
  -h, --help                     Show context-sensitive help.
      --workers=1                Number of concurrent workers (if less than 1
                                 use number of CPUs)

      --scan="."                 Source folder to scan
      --portrait="portrait"      Destination folder for portrait images.
                                 Relative to scan dir if not absolute.
      --landscape="landscape"    Destination folder for landscape images.
                                 Relative to scan dir if not absolute.
```

This command will scan (NOT recursively) for all supported image types in the `scan` folder and either copy (`cp`) or
move (`mv`) them to the respective destination folder, based on the aspect ratio of the image.

### mangle
```
Flags:
  -h, --help                  Show context-sensitive help.
      --workers=1             Number of concurrent workers (if less than 1 use
                              number of CPUs)

      --scan="."              Source folder to scan
      --dest="mangled"        Destination folder for processed pictures.
                              Relative to scan dir if not absolute. If same as
                              scan dir, will overwrite source files.
      --format="unsup:png"    Output format of mangled image. If prefixed with
                              'unsup:' will convert only unsupported formats

resize
  --resize         Resize image
  --width=INT      Max width
  --height=INT     Max height
  --crop           Crop image to maintain requested aspect ration
  --fill=STRING    If given and not cropping, will fill background with this
                   color to maintain destination aspect ratio

palette
  --palette=STRING    Palette name (bw, spectra6, mattdm6, gray16, vga16,
                      vga256) or PAL file in RIFF format to apply
  --dither            Apply dithering
```

This command will scan (NOT recursively) for all supported image types in the `scan` folder and attempt to process them
in the following order, saving the resulted images in the `dest` folder. If source and destination folders match, it
will replace the original file:
- if `resize` is specified, at least one of the dimensions (`width` or `height`) need to be given, and the tool will
  change dimensions of the source files to match what is requested but maintaining the original aspect ratio. This means
  the resulting dimensions may be smaller than the ones requested. To get the exact dimensions, either `crop` or
  `fill` need to be specified:
  - `crop` will trim edges off the source so the resulting image fits the given aspect ratio.
  - `fill` will pad the image with bars of the color specified, to fit the given aspect ratio. The color is in web
    format (#RGB, #RGBA, #RRGGBB, #RRGGBBAA).
- if a `palette` is given, it will convert the image from its source color space to the given palette. A few are built
  in, or a custom one can be given as a file in RIFF format. The result can be dithered for better visual results.

The image type will be preserved, if possible, but not all input types can also be written to. The tool can currently
read from GIF, JPEG, PNG, BMP, TIFF, WEBP and write to GIF, JPEG, PNG, BMP, TIFF. Writing to WEBP is not supported. Use
the `format` flag to save to all files in the  given format. To convert the type only for unsupported input formats,
prefix the flag value with `unsup:`.

### Concurrency
By default the tools processes images sequentially, but some commands support parallelism. To process multiple images at
the same time, use the `workers` flag. A value less than 1 means using as many workers as the number of CPUs detected in
the system. A good value needs to balance between CPU and disk load generated, depending on the requested operations.

## Installation
Just download and `go build` or `go run main.go`. Requires the Go toolchain.

## License
Licensed under the MIT License - see LICENSE.md file for details