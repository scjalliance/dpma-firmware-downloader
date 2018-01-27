# dpma-firmware-downloader

Downloads dpma-compatible phone firmware from Digium servers

## CLI

```
Usage of dpma-firmware-downloader.exe:
  -cachedir string
        Directory in which to save cache files (default "cache")
  -exclude value
        Models to exclude, comma-separated values or globs
  -excludefiles value
        Files to exclude, value or glob
  -firmwaredir string
        Directory in which to save firmware (default "fw")
  -flatten
        flatten extracted files to single directory
  -include value
        Models to include, comma-separated values or globs (default *)
  -includefiles value
        Files to include, value or glob (default *.eff)
  -latest int
        # of releases to download for each model (0 for unlimited)
  -url string
        URL of DPMA manifest (default "https://downloads.digium.com/pub/telephony/res_digium_phone/firmware/dpma-firmware.json")
```

## Docker
