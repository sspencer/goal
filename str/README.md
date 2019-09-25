# str

Go string utils.

Starting off some chunking:

```go
str.ChunkString("1234", 3)  -> ["123", "4"]
str.Chunkstring("1234", -3) -> ["1", "234"]
    
str.Comma(1234567) -> "1,234,567"
```

And some concurrent string workers (I use it for file transformations, input file to output file).

```go
func (m myWorker) StringWork(fn string) string {
    if err := transform(fn); err != nil {
        return ""
    }

    return fn + ".processed"
}

func (m myWorker) startProcessing(filenames []string) {
    // transform 4 files at a time
	results := str.Worker(4, filenames, m)
}
```