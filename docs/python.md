To use Chew with python you need to first build the package to create the shared object file and header file. You can do this by running the following command:

```bash
go build -o chew.so -buildmode=c-shared ./cmd/chew/wrapper.go
```

This will create a `chew.so` and `chew.h` file in the current directory. You can then use these files in your python project to use Chew. Here is an example of how to use Chew in your python project:

```python
import ctypes

chew_lib = ctypes.CDLL('./chew.so')

chew_lib.Process.argtypes = [ctypes.c_char_p]
chew_lib.Process.restype = ctypes.c_char_p

url = "https://example.com"
result = chew_lib.Process(url.encode('utf-8'))

print(result.decode('utf-8'))
```

With the above code snippet, you can now use Chew in your python project. I can't speak for the limitations of using Chew in python as I have not extensively tested it myself.
