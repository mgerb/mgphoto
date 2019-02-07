# MGPhoto

[Download the latest release here.](https://github.com/mgerb/mgphoto/releases)

- duplicates are skipped
- preserve original files
- no files will be overwritten
- currently only works with [EXIF](https://en.wikipedia.org/wiki/Exif) for dates
- unknown dates will go in **unknown** folder

Photos are not renamed unless a file already exists with that name
e.g. **IMG_1.jpg** will be renamed to **IMG_1_1.jpg**.

## Usage

```
mgphoto -o ./outputPath ./photos
```

Recursively scans entire directory along with
nested directories and outputs image/video files
in the following format.

```
2017/
└── 2017-08-15/
    └── IMG_1.jpg

2018/
├── 2018-02-23/
│   ├── IMG_2.jpg
│   ├── IMG_3.jpg
│   ├── IMG_4.jpg
└── 2018-03-01/
    └── IMG_5.jpg
```
